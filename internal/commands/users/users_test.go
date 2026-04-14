package users

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUsersUpdateHasGenerateSkeleton(t *testing.T) {
	root := &cobra.Command{Use: "megaport-cli"}
	AddCommandsTo(root)
	updateCmd, _, _ := root.Find([]string{"users", "update"})
	assert.NotNil(t, updateCmd.Flags().Lookup("generate-skeleton"))
}
