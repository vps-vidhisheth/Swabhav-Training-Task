package bank

import (
	"banking-app/apperror"
	"banking-app/customer"
	"banking-app/helper"
	"strings"
)

// Bank represents a single bank entity
type Bank struct {
	BankID   int
	Name     string
	IsActive bool
}

// BankManager manages bank operations and storage
type BankManager struct {
	banks         map[int]*Bank
	bankIDCounter int
}

// AccountChecker interface for checking active accounts
type AccountChecker interface {
	HasActiveAccounts(bankID int) (bool, error)
}

// NewBankManager creates a new BankManager instance
func NewBankManager() *BankManager {
	return &BankManager{
		banks:         make(map[int]*Bank),
		bankIDCounter: 0,
	}
}

// isAuthorized checks if the caller is an active admin
func isAuthorized(caller *customer.Customer) bool {
	return caller != nil && caller.IsActive && caller.IsAdmin
}

// CreateBank creates a new bank (Admin-only)
func (bm *BankManager) CreateBank(caller *customer.Customer, name string) (*Bank, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("create bank")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperror.NewValidationError("name", "bank name cannot be empty")
	}

	bm.bankIDCounter++
	bank := &Bank{
		BankID:   bm.bankIDCounter,
		Name:     name,
		IsActive: true,
	}
	bm.banks[bank.BankID] = bank
	return bank, nil
}

// GetBank retrieves a bank by ID (Admin-only)
func (bm *BankManager) GetBank(caller *customer.Customer, bankID int) (*Bank, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("get bank")
	}

	bank, exists := bm.banks[bankID]
	if !exists || !bank.IsActive {
		return nil, apperror.NewNotFoundError("bank", bankID)
	}
	return bank, nil
}

// GetAllBanksPaginated returns a paginated list of active banks (Admin-only)
func (bm *BankManager) GetAllBanksPaginated(caller *customer.Customer, page, pageSize int) ([]*Bank, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("get banks")
	}

	var activeBanks []*Bank
	for _, bank := range bm.banks {
		if bank.IsActive {
			activeBanks = append(activeBanks, bank)
		}
	}

	start, end := helper.PaginationBounds(page, pageSize, len(activeBanks))
	return activeBanks[start:end], nil
}

// UpdateName updates the bank's name
func (bank *Bank) UpdateName(value interface{}) error {
	name, ok := value.(string)
	if !ok || strings.TrimSpace(name) == "" {
		return apperror.NewValidationError("name", "must be a non-empty string")
	}
	bank.Name = strings.TrimSpace(name)
	return nil
}

// UpdateIsActive updates the bank's active status
func (bank *Bank) UpdateIsActive(value interface{}) error {
	isActive, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	bank.IsActive = isActive
	return nil
}

// UpdateField updates a specific field of the bank
func (bank *Bank) UpdateField(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "name":
		return bank.UpdateName(value)
	case "isactive":
		return bank.UpdateIsActive(value)
	default:
		return apperror.NewValidationError("field", "unknown update field")
	}
}

// UpdateBankFieldByID updates a specific field of a bank by ID (Admin-only)
func (bm *BankManager) UpdateBankFieldByID(caller *customer.Customer, bankID int, field string, value interface{}) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("update bank")
	}

	bank, exists := bm.banks[bankID]
	if !exists || !bank.IsActive {
		return apperror.NewNotFoundError("bank", bankID)
	}
	return bank.UpdateField(field, value)
}

// SoftDelete deactivates a bank after checking for active accounts
func (bank *Bank) SoftDelete(checker AccountChecker) error {
	if !bank.IsActive {
		return apperror.NewBankError("delete", "bank already inactive")
	}
	hasActive, err := checker.HasActiveAccounts(bank.BankID)
	if err != nil {
		return err
	}
	if hasActive {
		return apperror.NewBankError("delete", "bank has active accounts")
	}
	bank.IsActive = false
	return nil
}

// SoftDeleteBank deactivates a bank by ID (Admin-only)
func (bm *BankManager) SoftDeleteBank(caller *customer.Customer, checker AccountChecker, bankID int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("delete bank")
	}

	bank, exists := bm.banks[bankID]
	if !exists || !bank.IsActive {
		return apperror.NewNotFoundError("bank", bankID)
	}
	return bank.SoftDelete(checker)
}

// GetBankCount returns the total number of banks (for testing purposes)
func (bm *BankManager) GetBankCount() int {
	return len(bm.banks)
}
