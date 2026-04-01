package users

import (
	"encoding/json"
	"fmt"
	"os"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func processJSONCreateUserInput(jsonStr, jsonFile string) (*megaport.CreateUserRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
	}

	req := &megaport.CreateUserRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return req, nil
}

func processFlagCreateUserInput(cmd *cobra.Command) (*megaport.CreateUserRequest, error) {
	firstName, _ := cmd.Flags().GetString("first-name")
	lastName, _ := cmd.Flags().GetString("last-name")
	email, _ := cmd.Flags().GetString("email")
	position, _ := cmd.Flags().GetString("position")
	phone, _ := cmd.Flags().GetString("phone")

	userPosition := megaport.UserPosition(position)
	if position != "" && !userPosition.IsValid() {
		return nil, fmt.Errorf("invalid position: %s. Valid positions: %s", position, userPosition.ValidPositions())
	}

	req := &megaport.CreateUserRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Position:  userPosition,
		Phone:     phone,
		Active:    true,
	}

	return req, nil
}

func processJSONUpdateUserInput(jsonStr, jsonFile string) (*megaport.UpdateUserRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		jsonData = []byte(jsonStr)
	}

	req := &megaport.UpdateUserRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return req, nil
}

func processFlagUpdateUserInput(cmd *cobra.Command) (*megaport.UpdateUserRequest, error) {
	req := &megaport.UpdateUserRequest{}
	fieldsUpdated := false

	if cmd.Flags().Changed("first-name") {
		firstName, _ := cmd.Flags().GetString("first-name")
		req.FirstName = &firstName
		fieldsUpdated = true
	}
	if cmd.Flags().Changed("last-name") {
		lastName, _ := cmd.Flags().GetString("last-name")
		req.LastName = &lastName
		fieldsUpdated = true
	}
	if cmd.Flags().Changed("email") {
		email, _ := cmd.Flags().GetString("email")
		req.Email = &email
		fieldsUpdated = true
	}
	if cmd.Flags().Changed("position") {
		position, _ := cmd.Flags().GetString("position")
		req.Position = &position
		fieldsUpdated = true
	}
	if cmd.Flags().Changed("phone") {
		phone, _ := cmd.Flags().GetString("phone")
		req.Phone = &phone
		fieldsUpdated = true
	}
	if cmd.Flags().Changed("active") {
		active, _ := cmd.Flags().GetBool("active")
		req.Active = &active
		fieldsUpdated = true
	}
	if cmd.Flags().Changed("notification-enabled") {
		notif, _ := cmd.Flags().GetBool("notification-enabled")
		req.NotificationEnabled = &notif
		fieldsUpdated = true
	}

	if !fieldsUpdated {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	return req, nil
}
