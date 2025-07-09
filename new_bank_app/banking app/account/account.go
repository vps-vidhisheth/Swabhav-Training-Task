package account

import (
	"banking-app/apperror"
	"fmt"
)

type Account struct {
	AccountID int
	BankID    int
	OwnerID   int
	Balance   float64
	IsActive  bool
}

// Factory function to create a new account
func NewAccount(accountID, ownerID, bankID int) (*Account, error) {
	if bankID <= 0 {
		return nil, apperror.NewValidationError("bankID", "invalid bank ID")
	}
	if ownerID <= 0 {
		return nil, apperror.NewValidationError("ownerID", "invalid owner ID")
	}
	return &Account{
		AccountID: accountID,
		BankID:    bankID,
		OwnerID:   ownerID,
		Balance:   1000,
		IsActive:  true,
	}, nil
}

func (a *Account) DepositMoney(callerID int, amount float64) error {
	if !a.IsActive {
		return apperror.NewAccountError("deposit", fmt.Sprintf("account %d is inactive", a.AccountID))
	}
	if a.OwnerID != callerID {
		return apperror.NewAuthError("deposit caller does not own this account")
	}
	if amount <= 0 {
		return apperror.NewValidationError("amount", "deposit must be positive")
	}
	a.Balance += amount
	return nil
}

func (a *Account) WithdrawMoney(callerID int, amount float64) error {
	if !a.IsActive {
		return apperror.NewAccountError("withdraw", fmt.Sprintf("account %d is inactive", a.AccountID))
	}
	if a.OwnerID != callerID {
		return apperror.NewAuthError("withdraw caller does not own this account")
	}
	if amount <= 0 {
		return apperror.NewValidationError("amount", "withdrawal must be positive")
	}
	if a.Balance < amount {
		return apperror.NewValidationError("balance", "insufficient funds")
	}
	a.Balance -= amount
	return nil
}
