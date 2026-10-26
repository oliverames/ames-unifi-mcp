// Package errorreport sends optional, scrubbed panic reports to Sentry.
//
// Reporting is off unless SENTRY_DSN is set. When it is on, the only events
// sent are panics recovered from tool handlers, and each one is reduced to the
// panic's Go type, a fixed message, code-location stack frames, and the tool
// name. UniFi responses, tool arguments, and error text can carry device
// names, MAC and IP addresses, SSIDs, client hostnames, and credentials, so
// none of them are ever attached. Request data, user data, server name,
// breadcrumbs, modules, and threads are stripped from every event before it
// leaves the process.
package errorreport

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getsentry/sentry-go"
)

// ScrubbedMessage replaces every exception value and event message.
const ScrubbedMessage = "Error details withheld to protect network data."

// PanicMessage is the fixed exception value for recovered tool panics.
const PanicMessage = "Tool handler panicked; details withheld to protect network data."

// ModulePath is this module's import path. Frames inside it are reported
// relative to the repository root.
const ModulePath = "github.com/oliverames/ames-unifi-mcp"

var (
	enabled atomic.Bool

	safeTypeName = regexp.MustCompile(`^[A-Za-z_*\[\]][A-Za-z0-9_.*\[\]]{0,63}$`)
	safeTagValue = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,64}$`)
	safeModule   = regexp.MustCompile(`^[A-Za-z0-9_.~/-]{1,200}$`)
	safeFunction = regexp.MustCompile(`^[A-Za-z0-9_.*()\[\]{}/-]{1,200}$`)
	safeFileBase = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,100}\.(go|s)$`)

	safeTagKeys     = map[string]bool{"tool": true, "source": true}
	safeContextKeys = map[string]bool{"runtime": true, "os": true}
)

// Init starts Sentry when dsn is non-empty. It returns whether reporting is on.
// A bad DSN disables reporting without failing the server.
func Init(dsn, release string) bool {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return false
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:     dsn,
		Release: release,
		// DataCollection supersedes SendDefaultPII; both are set to collect nothing.
		SendDefaultPII:   false,
		TracesSampleRate: 0,
		DataCollection: &sentry.DataCollection{
			UserInfo:    sentry.Set(false),
			Cookies:     &sentry.KeyValueCollectionBehavior{Mode: sentry.CollectionOff},
			HTTPHeaders: &sentry.HeaderCollectionConfig{Request: &sentry.KeyValueCollectionBehavior{Mode: sentry.CollectionOff}, Response: &sentry.KeyValueCollectionBehavior{Mode: sentry.CollectionOff}},
			HTTPBodies:  []sentry.BodyType{},
			QueryParams: &sentry.KeyValueCollectionBehavior{Mode: sentry.CollectionOff},
		},
		// A non-empty value stops the SDK from filling in the hostname.
		ServerName:       "ames-unifi-mcp",
		MaxBreadcrumbs:   -1,
		BeforeBreadcrumb: func(*sentry.Breadcrumb, *sentry.BreadcrumbHint) *sentry.Breadcrumb { return nil },
		BeforeSend:       func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event { return ScrubEvent(event) },
		BeforeSendTransaction: func(*sentry.Event, *sentry.EventHint) *sentry.Event {
			return nil
		},
	})
	if err != nil {
		// The error text can echo the DSN, so report only that it failed.
		fmt.Fprintln(os.Stderr, "Sentry error reporting disabled: invalid SENTRY_DSN.")
		return false
	}
	enabled.Store(true)
	return true
}

// Enabled reports whether Init succeeded.
func Enabled() bool { return enabled.Load() }

// Flush waits up to timeout for queued events to be sent.
func Flush(timeout time.Duration) bool {
	if !enabled.Load() {
		return true
	}
	return sentry.Flush(timeout)
}

// RecoverTool must be deferred directly by a tool handler. It recovers a
// panic, reports it with only the tool name and panic type, and turns it into
// an error result so one broken tool cannot take down the server.
func RecoverTool(tool string, errp *error) {
	v := recover()
	if v == nil {
		return
	}
	CapturePanic(tool, v)
	if errp != nil {
		*errp = fmt.Errorf("internal error in tool %s", SafeToolName(tool))
	}
}

// CapturePanic reports a recovered panic value. The value itself is never
// sent, only its Go type.
func CapturePanic(tool string, v any) {
	if !enabled.Load() {
		return
	}
	name := SafeToolName(tool)
	typ := PanicType(v)
	handled := true
	event := sentry.NewEvent()
	event.Level = sentry.LevelError
	event.Exception = []sentry.Exception{{
		Type:       typ,
		Value:      PanicMessage,
		Stacktrace: sentry.NewStacktrace(),
		Mechanism:  &sentry.Mechanism{Type: "recover", Handled: &handled},
	}}
	event.Tags = map[string]string{"tool": name, "source": "tool_panic"}
	event.Fingerprint = []string{"tool-panic", name, typ}
	sentry.CaptureEvent(event)
}

// SafeToolName returns name if it looks like a tool identifier, else "unknown".
func SafeToolName(name string) string {
	if safeTagValue.MatchString(name) {
		return name
	}
	return "unknown"
}

// PanicType names the Go type of a panic value without its contents.
func PanicType(v any) string {
	name := fmt.Sprintf("%T", v)
	if safeTypeName.MatchString(name) {
		return name
	}
	return "panic"
}

// ScrubEvent reduces any outgoing event to the safe shape. It runs on every
// event, including ones built by CapturePanic, as defense in depth against
// SDK defaults or future capture calls.
func ScrubEvent(event *sentry.Event) *sentry.Event {
	if event == nil {
		return nil
	}
	event.Request = nil
	event.User = sentry.User{}
	event.ServerName = ""
	event.Breadcrumbs = nil
	event.Modules = nil
	event.Threads = nil
	event.Attachments = nil
	event.Transaction = ""
	event.Logger = ""
	if event.Message != "" {
		event.Message = ScrubbedMessage
	}
	for key := range event.Contexts {
		if !safeContextKeys[key] {
			delete(event.Contexts, key)
		}
	}
	for key, value := range event.Tags {
		if !safeTagKeys[key] || !safeTagValue.MatchString(value) {
			delete(event.Tags, key)
		}
	}
	for i := range event.Exception {
		ex := &event.Exception[i]
		if !safeTypeName.MatchString(ex.Type) {
			ex.Type = "Error"
		}
		if ex.Value != PanicMessage {
			ex.Value = ScrubbedMessage
		}
		ex.Module = ""
		if ex.Mechanism != nil {
			ex.Mechanism.Data = nil
			ex.Mechanism.Description = ""
		}
		if ex.Stacktrace != nil {
			ex.Stacktrace.Frames = scrubFrames(ex.Stacktrace.Frames)
		}
	}
	return event
}

// scrubFrames keeps only code locations. Absolute paths name the build
// machine's home directory, so each file is rebuilt from the frame's package
// path and the file's base name.
func scrubFrames(frames []sentry.Frame) []sentry.Frame {
	out := frames[:0]
	for _, f := range frames {
		file := f.AbsPath
		if file == "" {
			file = f.Filename
		}
		base := path.Base(strings.ReplaceAll(file, `\`, "/"))
		if !safeFileBase.MatchString(base) {
			continue
		}
		if f.Module != "" && !safeModule.MatchString(f.Module) {
			continue
		}
		if f.Function != "" && !safeFunction.MatchString(f.Function) {
			continue
		}
		switch {
		case f.Module == ModulePath:
			f.Filename = base
		case strings.HasPrefix(f.Module, ModulePath+"/"):
			f.Filename = strings.TrimPrefix(f.Module, ModulePath+"/") + "/" + base
		case f.Module != "":
			f.Filename = f.Module + "/" + base
		default:
			f.Filename = base
		}
		f.AbsPath = ""
		f.Vars = nil
		f.PreContext = nil
		f.ContextLine = ""
		f.PostContext = nil
		f.Package = ""
		out = append(out, f)
	}
	return out
}
