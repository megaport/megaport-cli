package servicekeys

import (
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// ServiceKeyOutput represents the desired fields for output
type ServiceKeyOutput struct {
	output.Output `json:"-" header:"-"`
	KeyUID        string `json:"key_uid" header:"KEY UID"`
	ProductName   string `json:"product_name" header:"PRODUCT NAME"`
	ProductUID    string `json:"product_uid" header:"PRODUCT UID"`
	Description   string `json:"description" header:"DESCRIPTION"`
	CreateDate    string `json:"create_date" header:"CREATE DATE"`
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
