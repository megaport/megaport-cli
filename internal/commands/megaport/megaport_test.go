package megaport

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {

	output := output.CaptureOutput(func() {
		versionCmd.Run(nil, nil)
	})

	expected := fmt.Sprintf("Megaport CLI Version: %s\n", version)
	assert.Equal(t, expected, output)
}
