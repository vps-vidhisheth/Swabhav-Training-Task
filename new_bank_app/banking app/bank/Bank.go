package bank

import (
	"banking-app/apperror"
	"banking-app/customer"
	"banking-app/helper"
	"fmt"
	"strings"
)

type Bank struct {
	BankID       int
	Name         string
	Abbreviation string
	IsActive     bool
}

type BankManager struct {
	banks           map[int]*Bank
	bankIDCounter   int
	getCustomerByID func(int) (*customer.Customer, error)
}

// Factory
func NewBankManager(getCustomerByID func(int) (*customer.Customer, error)) *BankManager {
	return &BankManager{
		banks:           make(map[int]*Bank),
		bankIDCounter:   0,
		getCustomerByID: getCustomerByID,
	}
}

// Auth check
func (bm *BankManager) isAuthorized(callerID int) bool {
	caller, err := bm.getCustomerByID(callerID)
	return err == nil && caller != nil && caller.IsActive && caller.IsAdmin
}

func GenerateAbbreviation(bankName string) string {
	words := strings.Fields(bankName)
	var initials []string
	for _, word := range words {
		if len(word) > 0 {
			initials = append(initials, string(word[0]))
		}
	}
	return strings.ToUpper(strings.Join(initials, ""))
}

func (bm *BankManager) CreateBank(callerID int, name string) (*Bank, error) {
	if !bm.isAuthorized(callerID) {
		return nil, apperror.NewAuthError("create bank: caller not authorized")
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, apperror.NewValidationError("name", "bank name cannot be empty")
	}

	abbr := GenerateAbbreviation(trimmedName)
	if abbr == "" {
		return nil, apperror.NewValidationError("abbreviation", "could not generate abbreviation")
	}

	bm.bankIDCounter++
	bank := &Bank{
		BankID:       bm.bankIDCounter,
		Name:         trimmedName,
		Abbreviation: abbr,
		IsActive:     true,
	}
	bm.banks[bank.BankID] = bank
	return bank, nil
}

func (bm *BankManager) GetBank(callerID int, bankID int) (*Bank, error) {
	if !bm.isAuthorized(callerID) {
		return nil, apperror.NewAuthError("get bank: caller not authorized")
	}

	bank, exists := bm.banks[bankID]
	if !exists {
		return nil, apperror.NewNotFoundError("bank", bankID)
	}
	if !bank.IsActive {
		return nil, apperror.NewBankError("get bank", fmt.Sprintf("bank with ID %d is inactive", bankID))
	}
	return bank, nil
}

func (bm *BankManager) GetAllBanksPaginated(callerID int, page, pageSize int) ([]*Bank, error) {
	if !bm.isAuthorized(callerID) {
		return nil, apperror.NewAuthError("get all banks: caller not authorized")
	}

	var activeBanks []*Bank
	for _, bank := range bm.banks {
		if bank.IsActive {
			activeBanks = append(activeBanks, bank)
		}
	}

	start, end := helper.PaginationBounds(page, pageSize, len(activeBanks))
	if start >= end {
		return []*Bank{}, nil
	}
	return activeBanks[start:end], nil
}

func (b *Bank) UpdateName(value interface{}) error {
	name, ok := value.(string)
	if !ok || strings.TrimSpace(name) == "" {
		return apperror.NewValidationError("name", "bank name must be a non-empty string")
	}
	b.Name = strings.TrimSpace(name)
	return nil
}

func (b *Bank) UpdateIsActive(value interface{}) error {
	isActive, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "active status must be a boolean")
	}
	b.IsActive = isActive
	return nil
}

func (b *Bank) UpdateAbbreviation(value interface{}) error {
	abbr, ok := value.(string)
	if !ok || strings.TrimSpace(abbr) == "" {
		return apperror.NewValidationError("abbreviation", "bank abbreviation must be a non-empty string")
	}
	b.Abbreviation = strings.TrimSpace(abbr)
	return nil
}

func (b *Bank) UpdateField(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "name":
		return b.UpdateName(value)
	case "isactive":
		return b.UpdateIsActive(value)
	case "abbreviation":
		return b.UpdateAbbreviation(value)
	default:
		return apperror.NewValidationError("field", fmt.Sprintf("unknown update field: '%s'", field))
	}
}

func (bm *BankManager) UpdateBankFieldByID(callerID int, bankID int, field string, value interface{}) error {
	if !bm.isAuthorized(callerID) {
		return apperror.NewAuthError("update bank: caller not authorized")
	}

	bank, exists := bm.banks[bankID]
	if !exists {
		return apperror.NewNotFoundError("bank", bankID)
	}
	if !bank.IsActive {
		return apperror.NewBankError("update bank", fmt.Sprintf("cannot update inactive bank with ID %d", bankID))
	}
	return bank.UpdateField(field, value)
}

func (bm *BankManager) SoftDeleteBank(
	callerID int,
	hasActiveAccounts func(bankID int) (bool, error),
	bankID int,
) error {
	if !bm.isAuthorized(callerID) {
		return apperror.NewAuthError("soft delete bank: caller not authorized")
	}

	bank, exists := bm.banks[bankID]
	if !exists {
		return apperror.NewNotFoundError("bank", bankID)
	}
	return bank.softDelete(hasActiveAccounts)
}

func (b *Bank) softDelete(hasActiveAccounts func(bankID int) (bool, error)) error {
	if !b.IsActive {
		return apperror.NewBankError("soft delete", fmt.Sprintf("bank '%s' (ID: %d) is already inactive", b.Name, b.BankID))
	}

	hasActive, err := hasActiveAccounts(b.BankID)
	if err != nil {
		return fmt.Errorf("soft delete: failed to check active accounts for bank ID %d: %w", b.BankID, err)
	}
	if hasActive {
		return apperror.NewBankError("soft delete", fmt.Sprintf("bank '%s' (ID: %d) still has active accounts", b.Name, b.BankID))
	}

	b.IsActive = false
	return nil
}

func (bm *BankManager) GetBankByID(bankID int) (*Bank, error) {
	bank, exists := bm.banks[bankID]
	if !exists || !bank.IsActive {
		return nil, apperror.NewBankError("get bank by ID", fmt.Sprintf("bank with ID %d not found or inactive", bankID))
	}
	return bank, nil
}

func (bm *BankManager) GetBankCount() int {
	return len(bm.banks)
}
