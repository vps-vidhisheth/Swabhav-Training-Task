package bank

import (
	"banking-app/apperror"
	"banking-app/helper"
	"strings"
)

// --- Bank Definition & Storage ---

type Bank struct {
	BankID   int
	Name     string
	IsActive bool
}

var (
	banks         = make(map[int]*Bank)
	bankIDCounter int
)

// --- Interfaces for external systems ---

type AccountChecker interface {
	HasActiveAccounts(bankID int) (bool, error)
}

// --- CRUD Operations ---

func CreateBank(name string) (*Bank, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperror.NewValidationError("name", "bank name cannot be empty")
	}
	bankIDCounter++
	b := &Bank{BankID: bankIDCounter, Name: name, IsActive: true}
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
	var all []*Bank
	for _, b := range banks {
		if b.IsActive {
			all = append(all, b)
		}
	}
	start, end := helper.PaginationBounds(page, pageSize, len(all))
	return all[start:end], nil
}

// --- Bank Methods for Update and Deactivate ---

func (b *Bank) UpdateName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("name", "must be a non-empty string")
	}
	b.Name = strings.TrimSpace(v)
	return nil
}

func (b *Bank) UpdateIsActive(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	b.IsActive = v
	return nil
}

func (b *Bank) UpdateField(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "name":
		return b.UpdateName(value)
	case "isactive":
		return b.UpdateIsActive(value)
	default:
		return apperror.NewValidationError("field", "unknown update field")
	}
}

// --- Package-level methods that wrap Bank methods and enforce authorization ---

func UpdateBankFieldByID(requester helper.Authorizer, bankID int, field string, value interface{}) error {
	if !requester.IsActiveUser() || !requester.IsAdminUser() {
		return apperror.NewAuthError("update bank")
	}
	b, err := GetBank(bankID)
	if err != nil {
		return err
	}
	return b.UpdateField(field, value)
}

func (b *Bank) SoftDelete(checker AccountChecker) error {
	if !b.IsActive {
		return apperror.NewBankError("delete", "bank already inactive")
	}
	hasActive, err := checker.HasActiveAccounts(b.BankID)
	if err != nil {
		return err
	}
	if hasActive {
		return apperror.NewBankError("delete", "bank has active accounts")
	}
	b.IsActive = false
	return nil
}

func SoftDeleteBank(requester helper.Authorizer, checker AccountChecker, bankID int) error {
	if !requester.IsActiveUser() || !requester.IsAdminUser() {
		return apperror.NewAuthError("delete bank")
	}
	b, err := GetBank(bankID)
	if err != nil {
		return err
	}
	return b.SoftDelete(checker)
}
