package users

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type UserOutput struct {
	output.Output `json:"-" header:"-"`
	EmployeeID    int    `json:"employee_id" header:"Employee ID"`
	FirstName     string `json:"first_name" header:"First Name"`
	LastName      string `json:"last_name" header:"Last Name"`
	Email         string `json:"email" header:"Email"`
	Position      string `json:"position" header:"Position"`
	Active        bool   `json:"active" header:"Active"`
}

type UserActivityOutput struct {
	output.Output `json:"-" header:"-"`
	LoginName     string `json:"login_name" header:"Login Name"`
	Description   string `json:"description" header:"Description"`
	Name          string `json:"name" header:"Activity"`
	CreateDate    string `json:"create_date" header:"Date"`
	UserType      string `json:"user_type" header:"User Type"`
}

func ToUserOutput(user *megaport.User) (UserOutput, error) {
	if user == nil {
		return UserOutput{}, fmt.Errorf("invalid user: nil value")
	}

	employeeID := user.PartyId
	if employeeID == 0 {
		employeeID = user.PersonId
	}

	return UserOutput{
		EmployeeID: employeeID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Position:   user.Position,
		Active:     user.Active,
	}, nil
}

func printUsers(users []*megaport.User, format string, noColor bool) error {
	outputs := make([]UserOutput, 0, len(users))
	for _, user := range users {
		out, err := ToUserOutput(user)
		if err != nil {
			return err
		}
		outputs = append(outputs, out)
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printUserActivities(activities []*megaport.UserActivity, format string, noColor bool) error {
	outputs := make([]UserActivityOutput, 0, len(activities))
	for _, activity := range activities {
		if activity == nil {
			continue
		}
		outputs = append(outputs, UserActivityOutput{
			LoginName:   activity.LoginName,
			Description: activity.Description,
			Name:        activity.Name,
			CreateDate:  activity.CreateDate.String(),
			UserType:    activity.UserType,
		})
	}
	return output.PrintOutput(outputs, format, noColor)
}

func displayUserChanges(original, updated *megaport.User, noColor bool) {
	if original == nil || updated == nil {
		return
	}

	fmt.Println()
	output.PrintInfo("Changes applied:", noColor)
	changesFound := false

	if original.FirstName != updated.FirstName {
		changesFound = true
		fmt.Printf("  • First Name: %s → %s\n",
			output.FormatOldValue(original.FirstName, noColor),
			output.FormatNewValue(updated.FirstName, noColor))
	}

	if original.LastName != updated.LastName {
		changesFound = true
		fmt.Printf("  • Last Name: %s → %s\n",
			output.FormatOldValue(original.LastName, noColor),
			output.FormatNewValue(updated.LastName, noColor))
	}

	if original.Email != updated.Email {
		changesFound = true
		fmt.Printf("  • Email: %s → %s\n",
			output.FormatOldValue(original.Email, noColor),
			output.FormatNewValue(updated.Email, noColor))
	}

	if original.Position != updated.Position {
		changesFound = true
		fmt.Printf("  • Position: %s → %s\n",
			output.FormatOldValue(original.Position, noColor),
			output.FormatNewValue(updated.Position, noColor))
	}

	if original.Active != updated.Active {
		changesFound = true
		oldActive := "No"
		if original.Active {
			oldActive = "Yes"
		}
		newActive := "No"
		if updated.Active {
			newActive = "Yes"
		}
		fmt.Printf("  • Active: %s → %s\n",
			output.FormatOldValue(oldActive, noColor),
			output.FormatNewValue(newActive, noColor))
	}

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}
