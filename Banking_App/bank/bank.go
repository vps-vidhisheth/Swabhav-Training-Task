package bank

import (
	"banking-app/apperror"
	"banking-app/helper"
	"strings"
)

type Bank struct {
	BankID   int
	Name     string
	IsActive bool
}

var (
	banks         = make(map[int]*Bank)
	bankIDCounter int
)

// --- AccountChecker interface (remains) ---

type AccountChecker interface {
	HasActiveAccounts(bankID int) (bool, error)
}

// --- Bank CRUD Functions ---

func CreateBank(name string) (*Bank, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperror.NewValidationError("name", "bank name cannot be empty")
	}

	bankIDCounter++
	b := &Bank{
		BankID:   bankIDCounter,
		Name:     name,
		IsActive: true,
	}
	banks[b.BankID] = b
	return b, nil
}

func GetBank(bankID int) (*Bank, error) {
	b, exists := banks[bankID]
	if !exists {
		return nil, apperror.NewNotFoundError("bank", bankID)
	}
	return b, nil
}

func GetAllBanksPaginated(requester helper.Authorizer, page, pageSize int) ([]*Bank, error) {
	if !requester.IsAdminUser() {
		return nil, apperror.NewAuthError("get banks")
	}

	var allBanks []*Bank
	for _, b := range banks {
		if b.IsActive {
			allBanks = append(allBanks, b)
		}
	}

	start, end := helper.PaginationBounds(page, pageSize, len(allBanks))
	return allBanks[start:end], nil
}

// --- Bank Field Updates ---

func (b *Bank) updateName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("name", "must be a non-empty string")
	}
	b.Name = strings.TrimSpace(v)
	return nil
}

func (b *Bank) updateIsActive(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	b.IsActive = v
	return nil
}

func (b *Bank) UpdateBankField(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "name":
		return b.updateName(value)
	case "isactive":
		return b.updateIsActive(value)
	default:
		return apperror.NewValidationError("field", "unknown update field")
	}
}

func UpdateBank(bankID int, field string, value interface{}, requester helper.Authorizer, checker AccountChecker) error {
	if !requester.IsAdminUser() || !requester.IsActiveUser() {
		return apperror.NewAuthError("update bank")
	}

	bank, err := GetBank(bankID)
	if err != nil {
		return err
	}

	// Validate before deactivation
	if strings.ToLower(field) == "isactive" {
		active, ok := value.(bool)
		if !ok {
			return apperror.NewValidationError("isactive", "must be a boolean")
		}
		if !active {
			if checker == nil {
				return apperror.NewBankError("deactivation", "account checker not provided")
			}
			hasActive, err := checker.HasActiveAccounts(bankID)
			if err != nil {
				return err
			}
			if hasActive {
				return apperror.NewBankError("deactivate", "bank has active accounts")
			}
		}
	}

	return bank.UpdateBankField(field, value)
}

// --- Soft Delete & Reactivate ---

func SoftDeleteBank(bankID int, requester helper.Authorizer, checker AccountChecker) error {
	return UpdateBank(bankID, "isactive", false, requester, checker)
}

func ReactivateBank(bankID int, requester helper.Authorizer) error {
	return UpdateBank(bankID, "isactive", true, requester, nil)
}
