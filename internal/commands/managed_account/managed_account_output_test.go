package managed_account

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintManagedAccounts_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "ACCOUNT NAME")
	assert.Contains(t, out, "ACCOUNT REF")
	assert.Contains(t, out, "COMPANY UID")

	assert.Contains(t, out, "Acme Corp")
	assert.Contains(t, out, "REF-001")
	assert.Contains(t, out, "company-uid-1")

	assert.Contains(t, out, "Beta Inc")
	assert.Contains(t, out, "REF-002")
	assert.Contains(t, out, "company-uid-2")

	assert.Contains(t, out, "┌")
	assert.Contains(t, out, "┐")
	assert.Contains(t, out, "└")
	assert.Contains(t, out, "┘")
	assert.Contains(t, out, "├")
	assert.Contains(t, out, "┤")
	assert.Contains(t, out, "│")
	assert.Contains(t, out, "─")
}

func TestPrintManagedAccounts_JSON(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
	"account_name": "Acme Corp",
	"account_ref": "REF-001",
	"company_uid": "company-uid-1"
  },
  {
	"account_name": "Beta Inc",
	"account_ref": "REF-002",
	"company_uid": "company-uid-2"
  }
]`
	assert.JSONEq(t, expected, out)
}

func TestPrintManagedAccounts_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `account_name,account_ref,company_uid
Acme Corp,REF-001,company-uid-1
Beta Inc,REF-002,company-uid-2
`
	assert.Equal(t, expected, out)
}

func TestPrintManagedAccounts_XML(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "xml", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, `<?xml version="1.0" encoding="UTF-8"?>`)
	assert.Contains(t, out, "<account_name>Acme Corp</account_name>")
	assert.Contains(t, out, "<account_ref>REF-001</account_ref>")
	assert.Contains(t, out, "<company_uid>company-uid-1</company_uid>")
	assert.Contains(t, out, "<account_name>Beta Inc</account_name>")

	// Output must be well-formed XML.
	decoder := xml.NewDecoder(strings.NewReader(out))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
}

func TestPrintManagedAccounts_Invalid(t *testing.T) {
	var err error
	out := output.CaptureOutput(func() {
		err = printManagedAccounts(testAccounts, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, out)
}

func TestPrintManagedAccounts_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		accounts []*megaport.ManagedAccount
		format   string
	}{
		{name: "empty slice table format", accounts: []*megaport.ManagedAccount{}, format: "table"},
		{name: "empty slice csv format", accounts: []*megaport.ManagedAccount{}, format: "csv"},
		{name: "empty slice json format", accounts: []*megaport.ManagedAccount{}, format: "json"},
		{name: "empty slice xml format", accounts: []*megaport.ManagedAccount{}, format: "xml"},
		{name: "nil slice table format", accounts: nil, format: "table"},
		{name: "nil slice csv format", accounts: nil, format: "csv"},
		{name: "nil slice json format", accounts: nil, format: "json"},
		{name: "nil slice xml format", accounts: nil, format: "xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := output.CaptureOutput(func() {
				err := printManagedAccounts(tt.accounts, tt.format, noColor)
				assert.NoError(t, err)
			})

			switch tt.format {
			case "table":
				assert.Contains(t, out, "ACCOUNT NAME")
				assert.Contains(t, out, "ACCOUNT REF")
				assert.Contains(t, out, "COMPANY UID")
				assert.Contains(t, out, "┌")
				assert.Contains(t, out, "┐")
				assert.Contains(t, out, "└")
				assert.Contains(t, out, "┘")
				assert.Contains(t, out, "│")
				assert.Contains(t, out, "─")
			case "csv":
				assert.Equal(t, "account_name,account_ref,company_uid\n", out)
			case "json":
				assert.Equal(t, "[]\n", out)
			case "xml":
				assert.Contains(t, out, `<?xml version="1.0" encoding="UTF-8"?>`)
				assert.Contains(t, out, "<items></items>")
			}
		})
	}
}

func TestPrintManagedAccounts_NilAccountInSlice(t *testing.T) {
	accounts := []*megaport.ManagedAccount{
		{
			AccountName: "Test Account",
			AccountRef:  "REF-001",
			CompanyUID:  "uid-001",
		},
		nil,
	}

	var err error
	output.CaptureOutput(func() {
		err = printManagedAccounts(accounts, "table", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid managed account: nil value")
}

func TestToManagedAccountOutput_EdgeCases(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		_, err := toManagedAccountOutput(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid managed account: nil value")
	})

	t.Run("zero value account", func(t *testing.T) {
		out, err := toManagedAccountOutput(&megaport.ManagedAccount{})
		assert.NoError(t, err)
		assert.Equal(t, "", out.AccountName)
		assert.Equal(t, "", out.AccountRef)
		assert.Equal(t, "", out.CompanyUID)
	})

	t.Run("full account", func(t *testing.T) {
		account := &megaport.ManagedAccount{
			AccountName: "Test Account",
			AccountRef:  "REF-123",
			CompanyUID:  "uid-456",
		}
		out, err := toManagedAccountOutput(account)
		assert.NoError(t, err)
		assert.Equal(t, "Test Account", out.AccountName)
		assert.Equal(t, "REF-123", out.AccountRef)
		assert.Equal(t, "uid-456", out.CompanyUID)
	})
}

func TestDisplayManagedAccountChanges(t *testing.T) {
	tests := []struct {
		name        string
		original    *megaport.ManagedAccount
		updated     *megaport.ManagedAccount
		expectedOut []string
	}{
		{
			name:        "name changed",
			original:    &megaport.ManagedAccount{AccountName: "Old Name", AccountRef: "REF-001"},
			updated:     &megaport.ManagedAccount{AccountName: "New Name", AccountRef: "REF-001"},
			expectedOut: []string{"Account Name:", "Old Name", "New Name"},
		},
		{
			name:        "ref changed",
			original:    &megaport.ManagedAccount{AccountName: "Same Name", AccountRef: "OLD-REF"},
			updated:     &megaport.ManagedAccount{AccountName: "Same Name", AccountRef: "NEW-REF"},
			expectedOut: []string{"Account Ref:", "OLD-REF", "NEW-REF"},
		},
		{
			name:        "both changed",
			original:    &megaport.ManagedAccount{AccountName: "Old Name", AccountRef: "OLD-REF"},
			updated:     &megaport.ManagedAccount{AccountName: "New Name", AccountRef: "NEW-REF"},
			expectedOut: []string{"Account Name:", "Account Ref:"},
		},
		{
			name:        "no changes",
			original:    &megaport.ManagedAccount{AccountName: "Same", AccountRef: "SAME"},
			updated:     &megaport.ManagedAccount{AccountName: "Same", AccountRef: "SAME"},
			expectedOut: []string{"No changes detected"},
		},
		{
			name:        "nil original",
			original:    nil,
			updated:     &megaport.ManagedAccount{AccountName: "Test"},
			expectedOut: []string{},
		},
		{
			name:        "nil updated",
			original:    &megaport.ManagedAccount{AccountName: "Test"},
			updated:     nil,
			expectedOut: []string{},
		},
		{
			name:        "both nil",
			original:    nil,
			updated:     nil,
			expectedOut: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedOutput := output.CaptureOutput(func() {
				displayManagedAccountChanges(tt.original, tt.updated, true)
			})

			for _, expected := range tt.expectedOut {
				assert.Contains(t, capturedOutput, expected)
			}
		})
	}
}
