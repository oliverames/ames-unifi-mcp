// Package inputvalidation enforces tool schemas before controller dispatch.
package inputvalidation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var compiled sync.Map
var macAddress = regexp.MustCompile(`^[0-9a-f]{2}(:[0-9a-f]{2}){5}$`)

var safeIdentifier = regexp.MustCompile(`^[A-Za-z0-9_.:-]+$`)

// Validate checks the declared schema and controller identifiers. It deliberately
// accepts opaque legacy IDs as well as ObjectIds and UUIDs, but never path syntax.
func Validate(schema, input json.RawMessage) error {
	var params map[string]any
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Errorf("invalid input JSON: %w", err)
	}
	if params == nil {
		return fmt.Errorf("input must be an object")
	}
	key := string(schema)
	cached, ok := compiled.Load(key)
	if !ok {
		var doc any
		if err := json.Unmarshal(schema, &doc); err != nil {
			return fmt.Errorf("invalid tool schema: %w", err)
		}
		compiler := jsonschema.NewCompiler()
		const location = "https://schemas.invalid/tool.json"
		if err := compiler.AddResource(location, doc); err != nil {
			return fmt.Errorf("invalid tool schema: %w", err)
		}
		resolved, err := compiler.Compile(location)
		if err != nil {
			return fmt.Errorf("invalid tool schema: %w", err)
		}
		cached, _ = compiled.LoadOrStore(key, resolved)
	}
	if err := cached.(*jsonschema.Schema).Validate(params); err != nil {
		return fmt.Errorf("input does not match tool schema: %w", err)
	}
	var declared struct {
		Properties map[string]struct {
			Type any `json:"type"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(schema, &declared); err != nil {
		return fmt.Errorf("invalid tool schema: %w", err)
	}
	required := make(map[string]bool, len(declared.Required))
	for _, key := range declared.Required {
		required[key] = true
	}
	for key, value := range params {
		field, known := declared.Properties[key]
		if !known || field.Type != "string" {
			continue
		}
		s, stringValue := value.(string)
		if !stringValue {
			continue
		} // The declared schema already rejected wrong types.
		if s == "" && !required[key] {
			continue
		} // Optional empty filters mean no filter.

		// Only top-level routing identifiers are checked. Arbitrary config payloads
		// retain their API-defined structure and are validated by their declared schema.
		switch {
		case key == "mac" || strings.HasSuffix(key, "_mac"):
			if !macAddress.MatchString(s) {
				return fmt.Errorf("%s must be a lowercase, colon-separated six-byte MAC address", key)
			}
		case key == "id" || strings.HasSuffix(key, "_id"):
			if s == "." || s == ".." || !safeIdentifier.MatchString(s) {
				return fmt.Errorf("%s must be a nonempty identifier without path separators", key)
			}
		}
	}
	return nil
}
