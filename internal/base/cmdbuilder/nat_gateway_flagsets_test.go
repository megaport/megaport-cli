package cmdbuilder

import (
	"testing"
)

func TestWithNATGatewayCreateFlags(t *testing.T) {
	cmd := NewCommand("test", "test").WithNATGatewayCreateFlags().Build()

	assertFlagExists(t, cmd, "name", "")
	assertFlagType(t, cmd, "name", "string")

	assertFlagExists(t, cmd, "term", "0")
	assertFlagType(t, cmd, "term", "int")

	assertFlagExists(t, cmd, "speed", "0")
	assertFlagType(t, cmd, "speed", "int")

	assertFlagExists(t, cmd, "location-id", "0")
	assertFlagType(t, cmd, "location-id", "int")

	assertFlagExists(t, cmd, "session-count", "0")
	assertFlagType(t, cmd, "session-count", "int")

	assertFlagExists(t, cmd, "diversity-zone", "")
	assertFlagType(t, cmd, "diversity-zone", "string")

	assertFlagExists(t, cmd, "promo-code", "")
	assertFlagType(t, cmd, "promo-code", "string")

	assertFlagExists(t, cmd, "service-level-reference", "")
	assertFlagType(t, cmd, "service-level-reference", "string")

	assertFlagExists(t, cmd, "auto-renew", "false")
	assertFlagType(t, cmd, "auto-renew", "bool")

	assertFlagExists(t, cmd, "resource-tags", "")
	assertFlagType(t, cmd, "resource-tags", "string")

	assertFlagExists(t, cmd, "resource-tags-file", "")
	assertFlagType(t, cmd, "resource-tags-file", "string")
}

func TestWithNATGatewayUpdateFlags(t *testing.T) {
	cmd := NewCommand("test", "test").WithNATGatewayUpdateFlags().Build()

	assertFlagExists(t, cmd, "name", "")
	assertFlagType(t, cmd, "name", "string")

	assertFlagExists(t, cmd, "term", "0")
	assertFlagType(t, cmd, "term", "int")

	assertFlagExists(t, cmd, "speed", "0")
	assertFlagType(t, cmd, "speed", "int")

	assertFlagExists(t, cmd, "location-id", "0")
	assertFlagType(t, cmd, "location-id", "int")

	assertFlagExists(t, cmd, "session-count", "0")
	assertFlagType(t, cmd, "session-count", "int")

	assertFlagExists(t, cmd, "diversity-zone", "")
	assertFlagType(t, cmd, "diversity-zone", "string")

	assertFlagExists(t, cmd, "promo-code", "")
	assertFlagType(t, cmd, "promo-code", "string")

	assertFlagExists(t, cmd, "service-level-reference", "")
	assertFlagType(t, cmd, "service-level-reference", "string")

	assertFlagExists(t, cmd, "auto-renew", "false")
	assertFlagType(t, cmd, "auto-renew", "bool")

	assertFlagExists(t, cmd, "resource-tags", "")
	assertFlagType(t, cmd, "resource-tags", "string")

	assertFlagExists(t, cmd, "resource-tags-file", "")
	assertFlagType(t, cmd, "resource-tags-file", "string")
}

func TestWithNATGatewayFilterFlags(t *testing.T) {
	cmd := NewCommand("test", "test").WithNATGatewayFilterFlags().Build()

	assertFlagExists(t, cmd, "location-id", "0")
	assertFlagType(t, cmd, "location-id", "int")

	assertFlagExists(t, cmd, "name", "")
	assertFlagType(t, cmd, "name", "string")

	assertFlagExists(t, cmd, "include-inactive", "false")
	assertFlagType(t, cmd, "include-inactive", "bool")
}

func TestWithNATGatewayTelemetryFlags(t *testing.T) {
	cmd := NewCommand("test", "test").WithNATGatewayTelemetryFlags().Build()

	assertFlagExists(t, cmd, "types", "")
	assertFlagType(t, cmd, "types", "string")

	assertFlagExists(t, cmd, "days", "0")
	assertFlagType(t, cmd, "days", "int")

	assertFlagExists(t, cmd, "from", "")
	assertFlagType(t, cmd, "from", "string")

	assertFlagExists(t, cmd, "to", "")
	assertFlagType(t, cmd, "to", "string")
}
