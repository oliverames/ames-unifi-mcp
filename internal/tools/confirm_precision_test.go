package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

type capturingTool struct {
	mockTool
	input json.RawMessage
}

func (m *capturingTool) Execute(_ context.Context, input json.RawMessage) (json.RawMessage, error) {
	m.executed = true
	m.input = append(json.RawMessage(nil), input...)
	return json.RawMessage(`{"ok":true}`), nil
}

func TestConfirmGatePreservesNestedNumbers(t *testing.T) {
	const numbers = `[9007199254740993,9007199254740992,42,-9007199254740993,0.125,1.234567890123456789,1e20]`
	for _, confirmation := range []string{"", `,"confirm":false`, `,"confirm":true`} {
		t.Run("confirmation="+confirmation, func(t *testing.T) {
			inner := &capturingTool{mockTool: mockTool{name: "configure", mutating: true}}
			input := json.RawMessage(`{"mac":"aa:bb:cc:dd:ee:ff","config":{"values":` + numbers + `,"nested":[{"value":9007199254740993,"password":"private-value"}],"token":12345678901234567}` + confirmation + `}`)
			result, err := WithConfirm(inner).Execute(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			confirmed := confirmation == `,"confirm":true`
			if inner.executed != confirmed {
				t.Fatalf("executed=%v, want %v", inner.executed, confirmed)
			}
			payload := inner.input
			if !confirmed {
				var preview struct {
					Parameters           json.RawMessage `json:"parameters"`
					RequiresConfirmation bool            `json:"requires_confirmation"`
				}
				if err := json.Unmarshal(result, &preview); err != nil {
					t.Fatal(err)
				}
				if !preview.RequiresConfirmation {
					t.Fatal("preview lost confirmation requirement")
				}
				if bytes.Contains(result, []byte("private-value")) || bytes.Contains(result, []byte("12345678901234567")) {
					t.Fatal("preview disclosed sensitive values")
				}
				payload = preview.Parameters
			}
			var params map[string]json.RawMessage
			if err := json.Unmarshal(payload, &params); err != nil {
				t.Fatal(err)
			}
			if confirmed {
				if _, exists := params["confirm"]; exists {
					t.Fatal("confirm forwarded to inner tool")
				}
			}
			var config struct {
				Values json.RawMessage `json:"values"`
				Nested []struct {
					Value    json.RawMessage `json:"value"`
					Password string          `json:"password"`
				} `json:"nested"`
				Token json.RawMessage `json:"token"`
			}
			if err := json.Unmarshal(params["config"], &config); err != nil {
				t.Fatal(err)
			}
			if string(config.Values) != numbers {
				t.Errorf("numbers changed: got %s, want %s", config.Values, numbers)
			}
			if len(config.Nested) != 1 || string(config.Nested[0].Value) != "9007199254740993" {
				t.Fatalf("nested integer changed: %+v", config.Nested)
			}
			wantPassword, wantToken := "[REDACTED]", `"[REDACTED]"`
			if confirmed {
				wantPassword, wantToken = "private-value", "12345678901234567"
			}
			if config.Nested[0].Password != wantPassword || string(config.Token) != wantToken {
				t.Fatalf("sensitive value handling changed: password=%q token=%s", config.Nested[0].Password, config.Token)
			}
		})
	}
}

func TestConfirmGateRejectsInvalidConfirmationInputs(t *testing.T) {
	for _, input := range []string{`{`, `null`, `[]`, `{"confirm":true}`, `{"mac":"aa:bb:cc:dd:ee:ff","confirm":"true"}`, `{"mac":"aa:bb:cc:dd:ee:ff","confirm":1}`, `{"mac":"aa:bb:cc:dd:ee:ff","confirm":null}`, `{"mac":"aa:bb:cc:dd:ee:ff","confirm":true} {}`} {
		inner := &capturingTool{mockTool: mockTool{name: "configure", mutating: true}}
		if _, err := WithConfirm(inner).Execute(context.Background(), []byte(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
		if inner.executed {
			t.Errorf("invalid input dispatched: %s", input)
		}
	}
}
