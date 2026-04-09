package registry

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Compile-time check that testModule satisfies Module.
var _ Module = &testModule{}

type testModule struct {
	name       string
	registered bool
}

func (m *testModule) Name() string { return m.name }
func (m *testModule) RegisterCommands(root *cobra.Command) {
	m.registered = true
	root.AddCommand(&cobra.Command{Use: m.name})
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)

	// Empty registry — RegisterAll is a no-op.
	root := &cobra.Command{Use: "root"}
	r.RegisterAll(root)
	assert.Empty(t, root.Commands())
}

func TestRegister_Single(t *testing.T) {
	r := NewRegistry()
	mod := &testModule{name: "ports"}
	r.Register(mod)

	root := &cobra.Command{Use: "root"}
	r.RegisterAll(root)

	assert.True(t, mod.registered)
	assert.Len(t, root.Commands(), 1)
	assert.Equal(t, "ports", root.Commands()[0].Use)
}

func TestRegister_Multiple(t *testing.T) {
	r := NewRegistry()
	mods := []*testModule{
		{name: "ports"},
		{name: "mcr"},
		{name: "vxc"},
	}
	for _, m := range mods {
		r.Register(m)
	}

	root := &cobra.Command{Use: "root"}
	r.RegisterAll(root)

	for _, m := range mods {
		assert.True(t, m.registered, "module %s should be registered", m.name)
	}
	assert.Len(t, root.Commands(), 3)

	names := make([]string, 0, len(root.Commands()))
	for _, cmd := range root.Commands() {
		names = append(names, cmd.Use)
	}
	assert.Contains(t, names, "ports")
	assert.Contains(t, names, "mcr")
	assert.Contains(t, names, "vxc")
}

func TestRegisterAll_Empty(t *testing.T) {
	r := NewRegistry()
	root := &cobra.Command{Use: "root"}

	// Should not panic.
	r.RegisterAll(root)
	assert.Empty(t, root.Commands())
}

func TestRegister_Duplicate(t *testing.T) {
	r := NewRegistry()
	mod1 := &testModule{name: "ports"}
	mod2 := &testModule{name: "ports"}
	r.Register(mod1)
	r.Register(mod2)

	root := &cobra.Command{Use: "root"}
	r.RegisterAll(root)

	// Both modules get RegisterCommands called (no dedup).
	assert.True(t, mod1.registered)
	assert.True(t, mod2.registered)
}
