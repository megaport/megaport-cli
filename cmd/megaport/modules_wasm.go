//go:build js && wasm
// +build js,wasm

package megaport

import (
	"github.com/megaport/megaport-cli/internal/commands/locations"
	"github.com/megaport/megaport-cli/internal/commands/mcr"
	"github.com/megaport/megaport-cli/internal/commands/mve"
	"github.com/megaport/megaport-cli/internal/commands/partners"
	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/commands/servicekeys"
	"github.com/megaport/megaport-cli/internal/commands/vxc"
)

// registerModules registers all WASM-supported command modules
// The following commands are NOT supported in WASM:
// - config: Config profiles are not supported; use session-based auth via browser UI instead
// - completion: Shell completion is not applicable in browser environment
// - generate-docs: Documentation generation is a development-time tool, not needed in WASM
// - version: Version information is not applicable in browser WASM environment
func registerModules() {
	// Register only WASM-compatible modules
	moduleRegistry.Register(ports.NewModule())
	moduleRegistry.Register(vxc.NewModule())
	moduleRegistry.Register(mcr.NewModule())
	moduleRegistry.Register(mve.NewModule())
	moduleRegistry.Register(locations.NewModule())
	moduleRegistry.Register(partners.NewModule())
	moduleRegistry.Register(servicekeys.NewModule())
}
