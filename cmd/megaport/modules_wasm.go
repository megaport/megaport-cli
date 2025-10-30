//go:build js && wasm
// +build js,wasm

package megaport

import (
	"github.com/megaport/megaport-cli/internal/commands/completion"
	"github.com/megaport/megaport-cli/internal/commands/generate_docs"
	"github.com/megaport/megaport-cli/internal/commands/locations"
	"github.com/megaport/megaport-cli/internal/commands/mcr"
	"github.com/megaport/megaport-cli/internal/commands/mve"
	"github.com/megaport/megaport-cli/internal/commands/partners"
	"github.com/megaport/megaport-cli/internal/commands/ports"
	"github.com/megaport/megaport-cli/internal/commands/servicekeys"
	"github.com/megaport/megaport-cli/internal/commands/version"
	"github.com/megaport/megaport-cli/internal/commands/vxc"
)

// registerModules registers all command modules EXCEPT config for WASM
// Config profiles are not supported in WASM - use session-based auth via browser UI instead
func registerModules() {
	// Register all modules except config
	moduleRegistry.Register(version.NewModule())
	moduleRegistry.Register(ports.NewModule())
	moduleRegistry.Register(vxc.NewModule())
	moduleRegistry.Register(mcr.NewModule())
	moduleRegistry.Register(mve.NewModule())
	moduleRegistry.Register(locations.NewModule())
	moduleRegistry.Register(partners.NewModule())
	moduleRegistry.Register(servicekeys.NewModule())
	moduleRegistry.Register(generate_docs.NewModule())
	moduleRegistry.Register(completion.NewModule())
	// NOTE: config module is NOT registered in WASM
	// Use session-based authentication via the browser UI login form instead
}
