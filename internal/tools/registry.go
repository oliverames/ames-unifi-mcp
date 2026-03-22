package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/version"
)

// Registry holds all registered tools and provides lookup/dispatch.
type Registry struct {
	mu          sync.RWMutex
	tools       map[string]Tool
	index       []ToolMeta
	permChecker *permissions.Checker
	version     version.Info
}

func NewRegistry(permChecker *permissions.Checker, ver version.Info) *Registry {
	return &Registry{
		tools:       make(map[string]Tool),
		permChecker: permChecker,
		version:     ver,
	}
}

// Register adds a tool to the registry. Mutating tools are automatically
// wrapped with the confirm gate. Tools requiring a higher controller version
// than detected are skipped. Tools blocked by permissions are skipped.
func (r *Registry) Register(t Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Version gate
	if minVer := t.MinVersion(); minVer != "" {
		required, err := version.Parse(minVer)
		if err == nil && !r.version.AtLeast(required.Major, required.Minor, required.Patch) {
			return nil // silently skip — controller too old
		}
	}

	// Permission gate: skip tools the profile doesn't allow
	if !r.permChecker.Allowed(t.Category(), t.Action()) {
		// Still add to index so the LLM knows it exists but can't use it
		r.index = append(r.index, ToolMeta{
			Name:        t.Name(),
			Description: t.Description() + " [PERMISSION DENIED — requires higher permission profile]",
			Category:    t.Category(),
			Mutating:    t.IsMutating(),
			MinVersion:  t.MinVersion(),
		})
		return nil
	}

	// Wrap mutating tools with confirm gate
	wrapped := WithConfirm(t)

	r.tools[wrapped.Name()] = wrapped
	r.index = append(r.index, MetaFromTool(wrapped))
	return nil
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// Index returns the full tool catalog, optionally filtered by category.
func (r *Registry) Index(category string) []ToolMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if category == "" {
		return r.index
	}

	var filtered []ToolMeta
	for _, m := range r.index {
		if strings.EqualFold(string(m.Category), category) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// All returns all registered (callable) tools.
func (r *Registry) All() map[string]Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]Tool, len(r.tools))
	for k, v := range r.tools {
		out[k] = v
	}
	return out
}

// Execute dispatches a tool call by name.
func (r *Registry) Execute(ctx context.Context, name string, input json.RawMessage) (json.RawMessage, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(ctx, input)
}

// Batch executes multiple tools in parallel.
func (r *Registry) Batch(ctx context.Context, calls []BatchCall) []BatchResult {
	results := make([]BatchResult, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c BatchCall) {
			defer wg.Done()
			data, err := r.Execute(ctx, c.Name, c.Input)
			results[idx] = BatchResult{
				Name: c.Name,
				Data: data,
			}
			if err != nil {
				results[idx].Error = err.Error()
			}
		}(i, call)
	}

	wg.Wait()
	return results
}

// BatchCall represents a single tool invocation in a batch.
type BatchCall struct {
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// BatchResult is the outcome of one batch call.
type BatchResult struct {
	Name  string          `json:"name"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}
