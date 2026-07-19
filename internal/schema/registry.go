// Package schema provides a self-describing schema registry for glean-cli commands.
// Each command registers a CommandSchema describing its flags, required fields,
// defaults, and an example invocation. The `glean schema` command exposes this
// registry to callers (human or agent) without needing documentation in context.
package schema

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// FlagSchema describes a single CLI flag.
type FlagSchema struct {
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// Surface values for CommandSchema.Surface.
const (
	// SurfacePlatform marks commands served by the platform APIs (/api/*).
	SurfacePlatform = "platform"
	// SurfaceLegacy marks commands served by the classic client APIs (/rest/api/v1/*).
	SurfaceLegacy = "legacy"
	// SurfaceRaw marks commands that make raw user-directed HTTP requests.
	SurfaceRaw = "raw"
	// SurfaceLocal marks commands that never call the Glean API.
	SurfaceLocal = "local"
)

// CommandSchema describes one glean command for agent introspection.
type CommandSchema struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	// WhenToUse is one sentence of agent guidance: the task this command is
	// the right tool for.
	WhenToUse string `json:"when_to_use,omitempty"`
	// Surface names the API family serving the command: one of the Surface*
	// constants. Commands migrated to the platform APIs flip this to
	// SurfacePlatform in the same PR.
	Surface string                `json:"surface,omitempty"`
	Flags   map[string]FlagSchema `json:"flags"`
	Example string                `json:"example"`
}

var (
	mu       sync.RWMutex
	registry = map[string]CommandSchema{}
)

// Register adds a CommandSchema to the global registry.
// Call this from each command's init or constructor.
func Register(s CommandSchema) {
	mu.Lock()
	defer mu.Unlock()
	registry[s.Command] = s
}

// Get returns the schema for a single command, or an error if not found.
func Get(command string) (CommandSchema, error) {
	mu.RLock()
	defer mu.RUnlock()
	s, ok := registry[command]
	if !ok {
		return CommandSchema{}, fmt.Errorf("no schema registered for command %q", command)
	}
	return s, nil
}

// List returns all registered command names in sorted order.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// MarshalList returns JSON for all registered schemas, keyed by command name.
func MarshalList() ([]byte, error) {
	mu.RLock()
	defer mu.RUnlock()
	return json.MarshalIndent(registry, "", "  ")
}
