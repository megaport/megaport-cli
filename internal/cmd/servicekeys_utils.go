package cmd

import (
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
)

// ServiceKeyOutput represents the desired fields for output
type ServiceKeyOutput struct {
	output
	KeyUID      string `json:"key_uid"`
	ProductName string `json:"product_name"`
	ProductUID  string `json:"product_uid"`
	Description string `json:"description"`
	CreateDate  string `json:"create_date"`
}

// ToServiceKeyOutput converts a ServiceKey to ServiceKeyOutput
func ToServiceKeyOutput(sk *megaport.ServiceKey) (ServiceKeyOutput, error) {
	if sk == nil {
		return ServiceKeyOutput{}, fmt.Errorf("nil service key")
	}

	output := ServiceKeyOutput{
		KeyUID:      sk.Key,
		ProductName: sk.ProductName,
		ProductUID:  sk.ProductUID,
		Description: sk.Description,
	}

	// Handle nil CreateDate
	if sk.CreateDate != nil {
		output.CreateDate = sk.CreateDate.Time.Format(time.RFC3339)
	}

	return output, nil
}
