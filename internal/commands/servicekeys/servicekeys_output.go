package servicekeys

import (
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type ServiceKeyOutput struct {
	output.Output `json:"-" header:"-"`
	KeyUID        string `json:"key_uid" header:"KEY UID"`
	ProductName   string `json:"product_name" header:"PRODUCT NAME"`
	ProductUID    string `json:"product_uid" header:"PRODUCT UID"`
	Description   string `json:"description" header:"DESCRIPTION"`
	CreateDate    string `json:"create_date" header:"CREATE DATE"`
}

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

	if sk.CreateDate != nil {
		output.CreateDate = sk.CreateDate.Time.Format(time.RFC3339)
	}

	return output, nil
}
