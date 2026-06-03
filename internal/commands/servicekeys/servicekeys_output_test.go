package servicekeys

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/stretchr/testify/assert"
)

func TestServiceKeyOutput_XML(t *testing.T) {
	outputs := make([]serviceKeyOutput, 0, len(mockServiceKeys))
	for _, sk := range mockServiceKeys {
		skOutput, err := toServiceKeyOutput(sk)
		assert.NoError(t, err)
		outputs = append(outputs, skOutput)
	}

	out := output.CaptureOutput(func() {
		err := output.PrintOutput(outputs, "xml", false)
		assert.NoError(t, err)
	})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "<items>")
	assert.Contains(t, out, "<key_uid>")
	assert.Contains(t, out, "abcd-1234-efgh-5678")
	assert.Contains(t, out, "Product One")
}
