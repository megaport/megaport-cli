package mcr

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// Extract the existing interactive prompting into a separate function for updating MCR
func promptForUpdateMCRDetails(mcrUID string, noColor bool) (*megaport.ModifyMCRRequest, error) {
	// Initialize request with MCR ID
	req := &megaport.ModifyMCRRequest{
		MCRID: mcrUID,
	}

	// Track if any field is updated
	fieldsUpdated := false

	// Prompt for name (can be skipped with empty input)
	namePrompt := "Enter new MCR name (leave empty to skip): "
	name, err := utils.Prompt(namePrompt, noColor)
	if err != nil {
		return nil, err
	}
	if name != "" {
		req.Name = name
		fieldsUpdated = true
	}

	// Prompt for cost centre (optional)
	costCentrePrompt := "Enter new cost centre (leave empty to skip): "
	costCentre, err := utils.Prompt(costCentrePrompt, noColor)
	if err != nil {
		return nil, err
	}
	if costCentre != "" {
		req.CostCentre = costCentre
		fieldsUpdated = true
	}

	// Prompt for marketplace visibility
	marketplaceVisibilityPrompt := "Update marketplace visibility? (yes/no, leave empty to skip): "
	marketplaceVisibilityStr, err := utils.Prompt(marketplaceVisibilityPrompt, noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(marketplaceVisibilityStr) == "yes" {
		visibilityValuePrompt := "Enter marketplace visibility (true or false): "
		visibilityValue, err := utils.Prompt(visibilityValuePrompt, noColor)
		if err != nil {
			return nil, err
		}

		marketplaceVisibility := strings.ToLower(visibilityValue) == "true"
		req.MarketplaceVisibility = &marketplaceVisibility
		fieldsUpdated = true
	}

	// Prompt for term (optional)
	termPrompt := "Enter new term (1, 12, 24, or 36 months, leave empty to skip): "
	termStr, err := utils.Prompt(termPrompt, noColor)
	if err != nil {
		return nil, err
	}
	if termStr != "" {
		term, err := strconv.Atoi(termStr)
		if err != nil {
			return nil, fmt.Errorf("invalid term: %v", err)
		}

		// Validate term value
		if term != 1 && term != 12 && term != 24 && term != 36 {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}

		req.ContractTermMonths = &term
		fieldsUpdated = true
	}

	// Make sure at least one field is being updated
	if !fieldsUpdated {
		return nil, fmt.Errorf("at least one field must be updated")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	return req, nil
}

// Extract the existing interactive prompting into a separate function for MCR
func promptForMCRDetails(noColor bool) (*megaport.BuyMCRRequest, error) {
	name, err := utils.Prompt("Enter MCR name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	termStr, err := utils.Prompt("Enter term (1, 12, 24, or 36 months) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("invalid term: %v", err)
	}

	portSpeedStr, err := utils.Prompt("Enter port speed (1000, 10000, or 100000 Mbps) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port speed: %v", err)
	}

	locationIDStr, err := utils.Prompt("Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID: %v", err)
	}

	asnStr, err := utils.Prompt("Enter MCR ASN (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	var asn int
	if asnStr != "" {
		asnValue, err := strconv.Atoi(asnStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ASN: %v", err)
		}
		asn = asnValue
	}

	// Optional fields
	diversityZone, err := utils.Prompt("Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	costCentre, err := utils.Prompt("Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	promoCode, err := utils.Prompt("Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	req := &megaport.BuyMCRRequest{
		Name:          name,
		Term:          term,
		PortSpeed:     portSpeed,
		LocationID:    locationID,
		MCRAsn:        asn,
		DiversityZone: diversityZone,
		CostCentre:    costCentre,
		PromoCode:     promoCode,
	}

	// Validate the request
	if err := validateMCRRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func promptForPrefixFilterListDetails(mcrUID string, noColor bool) (*megaport.CreateMCRPrefixFilterListRequest, error) {
	description, err := utils.Prompt("Enter description (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	addressFamily, err := utils.Prompt("Enter address family (IPv4 or IPv6) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if addressFamily != "IPv4" && addressFamily != "IPv6" {
		return nil, fmt.Errorf("invalid address family, must be IPv4 or IPv6")
	}

	// Prompt for entries
	entries := []*megaport.MCRPrefixListEntry{}
	for {
		fmt.Println("Add a new prefix filter entry (leave prefix blank to finish):")

		prefix, err := utils.Prompt("Enter prefix (e.g., 192.168.0.0/24): ", noColor)
		if err != nil {
			return nil, err
		}
		if prefix == "" {
			break
		}

		actionStr, err := utils.Prompt("Enter action (permit or deny): ", noColor)
		if err != nil {
			return nil, err
		}
		if actionStr != "permit" && actionStr != "deny" {
			return nil, fmt.Errorf("invalid action, must be permit or deny")
		}

		geStr, err := utils.Prompt("Enter GE value (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		var ge int
		if geStr != "" {
			geVal, err := strconv.Atoi(geStr)
			if err != nil {
				return nil, fmt.Errorf("invalid GE value, must be a number")
			}
			ge = geVal
		}

		leStr, err := utils.Prompt("Enter LE value (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		var le int
		if leStr != "" {
			leVal, err := strconv.Atoi(leStr)
			if err != nil {
				return nil, fmt.Errorf("invalid LE value, must be a number")
			}
			le = leVal
		}

		entries = append(entries, &megaport.MCRPrefixListEntry{
			Prefix: prefix,
			Action: actionStr,
			Ge:     ge,
			Le:     le,
		})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("at least one entry is required")
	}

	req := &megaport.CreateMCRPrefixFilterListRequest{
		MCRID: mcrUID,
		PrefixFilterList: megaport.MCRPrefixFilterList{
			Description:   description,
			AddressFamily: addressFamily,
			Entries:       entries,
		},
	}

	return req, nil
}

func promptForUpdatePrefixFilterListDetails(mcrUID string, prefixFilterListID int, noColor bool) (*megaport.MCRPrefixFilterList, error) {
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	currentPrefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current prefix filter list: %v", err)
	}

	fmt.Printf("Current description: %s\n", currentPrefixFilterList.Description)
	description, err := utils.Prompt("Enter new description (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if description == "" {
		description = currentPrefixFilterList.Description
	}

	// Just display the address family but don't allow changing it
	fmt.Printf("Address family: %s (cannot be changed after creation)\n", currentPrefixFilterList.AddressFamily)
	addressFamily := currentPrefixFilterList.AddressFamily

	// Initialize a zero-length slice with capacity to hold existing entries
	entries := make([]*megaport.MCRPrefixListEntry, 0, len(currentPrefixFilterList.Entries))

	modifyExisting, err := utils.Prompt("Do you want to modify existing entries? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(modifyExisting) != "yes" {
		// Just keep existing entries as is
		entries = append(entries, currentPrefixFilterList.Entries...)
	} else {
		for i, entry := range currentPrefixFilterList.Entries {
			fmt.Printf("Entry %d - Current: Action: %s, Prefix: %s, GE: %d, LE: %d\n",
				i+1, entry.Action, entry.Prefix, entry.Ge, entry.Le)

			keepEntry, err := utils.Prompt(fmt.Sprintf("Keep entry %d? (yes/no): ", i+1), noColor)
			if err != nil {
				return nil, err
			}

			if strings.ToLower(keepEntry) == "yes" {
				modifyEntry, err := utils.Prompt(fmt.Sprintf("Modify entry %d? (yes/no): ", i+1), noColor)
				if err != nil {
					return nil, err
				}

				if strings.ToLower(modifyEntry) == "yes" {
					prefix, err := utils.Prompt(fmt.Sprintf("Enter new prefix for entry %d (current: %s): ", i+1, entry.Prefix), noColor)
					if err != nil {
						return nil, err
					}
					if prefix == "" {
						prefix = entry.Prefix
					}

					actionStr, err := utils.Prompt(fmt.Sprintf("Enter new action for entry %d (permit or deny, current: %s): ", i+1, entry.Action), noColor)
					if err != nil {
						return nil, err
					}
					if actionStr == "" {
						actionStr = entry.Action
					} else if actionStr != "permit" && actionStr != "deny" {
						return nil, fmt.Errorf("invalid action, must be permit or deny")
					}

					geStr, err := utils.Prompt(fmt.Sprintf("Enter new GE value for entry %d (current: %d): ", i+1, entry.Ge), noColor)
					if err != nil {
						return nil, err
					}
					ge := entry.Ge
					if geStr != "" {
						geVal, err := strconv.Atoi(geStr)
						if err != nil {
							return nil, fmt.Errorf("invalid GE value, must be a number")
						}
						ge = geVal
					}

					leStr, err := utils.Prompt(fmt.Sprintf("Enter new LE value for entry %d (current: %d): ", i+1, entry.Le), noColor)
					if err != nil {
						return nil, err
					}
					le := entry.Le
					if leStr != "" {
						leVal, err := strconv.Atoi(leStr)
						if err != nil {
							return nil, fmt.Errorf("invalid LE value, must be a number")
						}
						le = leVal
					}

					entries = append(entries, &megaport.MCRPrefixListEntry{
						Prefix: prefix,
						Action: actionStr,
						Ge:     ge,
						Le:     le,
					})
				} else {
					entries = append(entries, entry)
				}
			}
		}
	}

	addNew, err := utils.Prompt("Do you want to add new entries? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(addNew) == "yes" {
		for {
			fmt.Println("Add a new prefix filter entry (leave prefix blank to finish):")

			prefix, err := utils.Prompt("Enter prefix (e.g., 192.168.0.0/24): ", noColor)
			if err != nil {
				return nil, err
			}
			if prefix == "" {
				break
			}

			actionStr, err := utils.Prompt("Enter action (permit or deny): ", noColor)
			if err != nil {
				return nil, err
			}
			if actionStr != "permit" && actionStr != "deny" {
				return nil, fmt.Errorf("invalid action, must be permit or deny")
			}

			geStr, err := utils.Prompt("Enter GE value (optional): ", noColor)
			if err != nil {
				return nil, err
			}
			var ge int
			if geStr != "" {
				geVal, err := strconv.Atoi(geStr)
				if err != nil {
					return nil, fmt.Errorf("invalid GE value, must be a number")
				}
				ge = geVal
			}

			leStr, err := utils.Prompt("Enter LE value (optional): ", noColor)
			if err != nil {
				return nil, err
			}
			var le int
			if leStr != "" {
				leVal, err := strconv.Atoi(leStr)
				if err != nil {
					return nil, fmt.Errorf("invalid LE value, must be a number")
				}
				le = leVal
			}

			entries = append(entries, &megaport.MCRPrefixListEntry{
				Prefix: prefix,
				Action: actionStr,
				Ge:     ge,
				Le:     le,
			})
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("at least one entry is required")
	}

	prefixFilterList := &megaport.MCRPrefixFilterList{
		ID:            prefixFilterListID,
		Description:   description,
		AddressFamily: addressFamily, // Always use current address family
		Entries:       entries,
	}

	if err := validateUpdatePrefixFilterList(prefixFilterList); err != nil {
		return nil, err
	}

	return prefixFilterList, nil
}
