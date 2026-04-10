package auth

import (
	"strings"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type authStatusOutput struct {
	output.Output `json:"-" header:"-"`
	FirstName     string `json:"first_name" header:"First Name"`
	LastName      string `json:"last_name" header:"Last Name"`
	Email         string `json:"email" header:"Email"`
	Position      string `json:"position" header:"Position"`
	CompanyName   string `json:"company_name" header:"Company"`
	Active        bool   `json:"active" header:"Active"`
	Profile       string `json:"profile" header:"Profile"`
	Environment   string `json:"environment" header:"Environment"`
	APIEndpoint   string `json:"api_endpoint" header:"API Endpoint"`
}

func toAuthStatusOutput(user *megaport.User, profileName, environment, apiEndpoint, companyName string) authStatusOutput {
	out := authStatusOutput{
		Profile:     profileName,
		Environment: capitalizeFirst(environment),
		APIEndpoint: apiEndpoint,
		CompanyName: companyName,
	}

	if user != nil {
		out.FirstName = user.FirstName
		out.LastName = user.LastName
		out.Email = user.Email
		out.Position = user.Position
		out.Active = user.Active
		if user.CompanyName != "" {
			out.CompanyName = user.CompanyName
		}
	}

	return out
}

func printAuthStatus(user *megaport.User, profileName, environment, apiEndpoint, companyName, format string, noColor bool) error {
	out := toAuthStatusOutput(user, profileName, environment, apiEndpoint, companyName)
	return output.PrintOutput([]authStatusOutput{out}, format, noColor)
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
