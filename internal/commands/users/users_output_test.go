package users

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToUserOutput_NilUser(t *testing.T) {
	_, err := toUserOutput(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil value")
}

func TestToUserOutput_Valid(t *testing.T) {
	user := &megaport.User{
		PartyId:   12345,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Position:  "Technical Admin",
		Active:    true,
	}

	out, err := toUserOutput(user)
	assert.NoError(t, err)
	assert.Equal(t, 12345, out.EmployeeID)
	assert.Equal(t, "John", out.FirstName)
	assert.Equal(t, "Doe", out.LastName)
	assert.Equal(t, "john@example.com", out.Email)
	assert.Equal(t, "Technical Admin", out.Position)
	assert.True(t, out.Active)
}

func TestToUserOutput_UsesPersonIdFallback(t *testing.T) {
	user := &megaport.User{
		PartyId:  0,
		PersonId: 67890,
	}

	out, err := toUserOutput(user)
	assert.NoError(t, err)
	assert.Equal(t, 67890, out.EmployeeID)
}

func TestDisplayUserChanges(t *testing.T) {
	tests := []struct {
		name             string
		original         *megaport.User
		updated          *megaport.User
		expectedContains []string
	}{
		{
			name:     "nil original",
			original: nil,
			updated:  &megaport.User{},
		},
		{
			name:     "nil updated",
			original: &megaport.User{},
			updated:  nil,
		},
		{
			name:             "no changes",
			original:         &megaport.User{FirstName: "Same", LastName: "User", Email: "same@test.com"},
			updated:          &megaport.User{FirstName: "Same", LastName: "User", Email: "same@test.com"},
			expectedContains: []string{"No changes detected"},
		},
		{
			name:             "first name changed",
			original:         &megaport.User{FirstName: "Old"},
			updated:          &megaport.User{FirstName: "New"},
			expectedContains: []string{"First Name", "Old", "New"},
		},
		{
			name:             "last name changed",
			original:         &megaport.User{LastName: "OldLast"},
			updated:          &megaport.User{LastName: "NewLast"},
			expectedContains: []string{"Last Name", "OldLast", "NewLast"},
		},
		{
			name:             "email changed",
			original:         &megaport.User{Email: "old@test.com"},
			updated:          &megaport.User{Email: "new@test.com"},
			expectedContains: []string{"Email", "old@test.com", "new@test.com"},
		},
		{
			name:             "position changed",
			original:         &megaport.User{Position: "Finance"},
			updated:          &megaport.User{Position: "Company Admin"},
			expectedContains: []string{"Position", "Finance", "Company Admin"},
		},
		{
			name:             "active changed",
			original:         &megaport.User{Active: true},
			updated:          &megaport.User{Active: false},
			expectedContains: []string{"Active", "Yes", "No"},
		},
		{
			name:             "multiple changes",
			original:         &megaport.User{FirstName: "Old", Email: "old@test.com", Active: true},
			updated:          &megaport.User{FirstName: "New", Email: "new@test.com", Active: false},
			expectedContains: []string{"First Name", "Email", "Active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := op.CaptureOutput(func() {
				displayUserChanges(tt.original, tt.updated, true)
			})

			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected)
			}
		})
	}
}
