package servicekeys

import (
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type serviceKeyOutput struct {
	output.Output `json:"-" header:"-"`
	KeyUID        string `json:"key_uid" header:"Key UID"`
	ProductName   string `json:"product_name" header:"Product Name"`
	ProductUID    string `json:"product_uid" header:"Product UID"`
	Description   string `json:"description" header:"Description"`
	CreateDate    string `json:"create_date" header:"Create Date"`
}

func toServiceKeyOutput(sk *megaport.ServiceKey) (serviceKeyOutput, error) {
	if sk == nil {
		return serviceKeyOutput{}, fmt.Errorf("nil service key")
	}

	output := serviceKeyOutput{
		KeyUID:      sk.Key,
		ProductName: sk.ProductName,
		ProductUID:  sk.ProductUID,
		Description: sk.Description,
	}

	if sk.CreateDate != nil {
		output.CreateDate = sk.CreateDate.Format(time.RFC3339)
	}

	return output, nil
}
