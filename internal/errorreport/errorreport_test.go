package errorreport

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// Private network values that must never appear in an outgoing event.
var secrets = []string{
	"Living Room AP", "aa:bb:cc:dd:ee:ff", "192.168.1.42", "HomeNet-5G", "olivers-macbook",
	"/Users/oliverames",
}

type captureTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (t *captureTransport) Configure(sentry.ClientOptions) {}
func (t *captureTransport) SendEvent(e *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
}
func (t *captureTransport) Flush(time.Duration) bool              { return true }
func (t *captureTransport) FlushWithContext(context.Context) bool { return true }
func (t *captureTransport) Close()                                {}

func TestInitWithoutDSNIsInert(t *testing.T) {
	if Init("  ", "test") {
		t.Fatal("Init with an empty DSN must not enable reporting")
	}
}

func TestScrubEventRemovesPrivateData(t *testing.T) {
	leak := strings.Join(secrets, " ")
	event := &sentry.Event{
		Message:     leak,
		ServerName:  "olivers-macbook",
		User:        sentry.User{IPAddress: "192.168.1.42", Username: "oliver"},
		Request:     &sentry.Request{URL: "https://192.168.1.1/api/s/default/stat/sta", Data: leak},
		Breadcrumbs: []*sentry.Breadcrumb{{Message: leak}},
		Modules:     map[string]string{"x": "1"},
		Tags:        map[string]string{"tool": "list_clients", "ssid": "HomeNet-5G", "source": "has space"},
		Contexts:    map[string]sentry.Context{"runtime": {"name": "go"}, "device": {"name": "Living Room AP"}},
		Exception: []sentry.Exception{{
			Type:      "client " + leak,
			Value:     leak,
			Mechanism: &sentry.Mechanism{Type: "generic", Data: map[string]any{"mac": "aa:bb:cc:dd:ee:ff"}},
			Stacktrace: &sentry.Stacktrace{Frames: []sentry.Frame{
				{Module: ModulePath + "/internal/tools/core", Function: "listClients", AbsPath: "/Users/oliverames/Developer/Projects/ames-unifi-mcp/internal/tools/core/clients.go", Vars: map[string]any{"host": "olivers-macbook"}, ContextLine: leak},
				{Module: "runtime", Function: "gopanic", AbsPath: "/usr/local/go/src/runtime/panic.go"},
				{Function: "weird " + leak, AbsPath: "/tmp/x.go"},
			}},
		}},
	}
	got := ScrubEvent(event)
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range secrets {
		if strings.Contains(string(raw), s) {
			t.Errorf("scrubbed event still contains %q: %s", s, raw)
		}
	}
	if got.Message != ScrubbedMessage || got.Exception[0].Value != ScrubbedMessage || got.Exception[0].Type != "Error" {
		t.Errorf("message or exception not replaced: %+v", got.Exception[0])
	}
	if got.Tags["tool"] != "list_clients" || len(got.Tags) != 1 {
		t.Errorf("tags = %v, want only tool", got.Tags)
	}
	if _, ok := got.Contexts["device"]; ok {
		t.Error("device context should be dropped")
	}
	frames := got.Exception[0].Stacktrace.Frames
	if len(frames) != 2 || frames[0].Filename != "internal/tools/core/clients.go" || frames[1].Filename != "runtime/panic.go" {
		t.Errorf("frames = %+v", frames)
	}
}

func TestRecoverToolReportsTypeAndToolOnly(t *testing.T) {
	transport := &captureTransport{}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:        "https://public@example.invalid/1",
		Transport:  transport,
		BeforeSend: func(e *sentry.Event, _ *sentry.EventHint) *sentry.Event { return ScrubEvent(e) },
	}); err != nil {
		t.Fatal(err)
	}
	enabled.Store(true)
	t.Cleanup(func() { enabled.Store(false) })

	err := func() (err error) {
		defer RecoverTool("get_client", &err)
		panic(errors.New("no client aa:bb:cc:dd:ee:ff at 192.168.1.42 on HomeNet-5G"))
	}()
	if err == nil || err.Error() != "internal error in tool get_client" {
		t.Fatalf("err = %v", err)
	}
	sentry.Flush(time.Second)

	transport.mu.Lock()
	defer transport.mu.Unlock()
	if len(transport.events) != 1 {
		t.Fatalf("got %d events, want 1", len(transport.events))
	}
	ev := transport.events[0]
	raw, _ := json.Marshal(ev)
	for _, s := range secrets {
		if strings.Contains(string(raw), s) {
			t.Errorf("event contains %q", s)
		}
	}
	if ev.Tags["tool"] != "get_client" || ev.Exception[0].Type != "*errors.errorString" || ev.Exception[0].Value != PanicMessage {
		t.Errorf("unexpected event: tags=%v exception=%+v", ev.Tags, ev.Exception[0])
	}
	if ev.ServerName != "" || ev.Request != nil {
		t.Error("server name or request not stripped")
	}
}

func TestSafeToolName(t *testing.T) {
	if SafeToolName("list_devices") != "list_devices" || SafeToolName("bad name; drop") != "unknown" {
		t.Error("SafeToolName did not validate")
	}
}
