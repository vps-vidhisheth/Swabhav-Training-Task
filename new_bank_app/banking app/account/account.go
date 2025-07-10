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

var accounts = make(map[int]*Account)

func NewAccount(accountID, ownerID, bankID int) (*Account, error) {
	if bankID <= 0 {
		return nil, apperror.NewValidationError("bankID", "must be greater than 0")
	}
	if ownerID <= 0 {
		return nil, apperror.NewValidationError("ownerID", "must be greater than 0")
	}
	if _, exists := accounts[accountID]; exists {
		return nil, apperror.NewValidationError("accountID", fmt.Sprintf("account %d already exists", accountID))
	}
	account := &Account{
		AccountID: accountID,
		BankID:    bankID,
		OwnerID:   ownerID,
		Balance:   1000,
		IsActive:  true,
	}
	accounts[accountID] = account
	return account, nil
}

func GetAccountById(accountID int) (*Account, error) {
	acc, ok := accounts[accountID]
	if !ok {
		return nil, apperror.NewNotFoundError("account", accountID)
	}
	return acc, nil
}

func (a *Account) DepositMoney(callerID int, amount float64) error {
	if !a.IsActive {
		return apperror.NewAccountError("deposit", fmt.Sprintf("account %d is inactive", a.AccountID))
	}
	if a.OwnerID != callerID {
		return apperror.NewAuthError("unauthorized access to deposit money")
	}
	if amount <= 0 {
		return apperror.NewValidationError("amount", "must be greater than 0")
	}
	a.Balance += amount
	return nil
}

func (a *Account) WithdrawMoney(callerID int, amount float64) error {
	if !a.IsActive {
		return apperror.NewAccountError("withdraw", fmt.Sprintf("account %d is inactive", a.AccountID))
	}
	if a.OwnerID != callerID {
		return apperror.NewAuthError("unauthorized access to withdraw money")
	}
	if amount <= 0 {
		return apperror.NewValidationError("amount", "must be greater than 0")
	}
	if a.Balance < amount {
		return apperror.NewValidationError("balance", "insufficient funds")
	}
	a.Balance -= amount
	return nil
}

func (acc *Account) TransferMoneyToExternal(targetAccID, fromCustomerID, toCustomerID int, amount float64) error {
	if acc.OwnerID != fromCustomerID {
		return apperror.NewAuthError("sender does not own the source account")
	}
	if !acc.IsActive {
		return apperror.NewAccountError("transfer", fmt.Sprintf("source account %d is inactive", acc.AccountID))
	}
	toAcc, err := GetAccountById(targetAccID)
	if err != nil {
		return err
	}
	if !toAcc.IsActive {
		return apperror.NewAccountError("transfer", fmt.Sprintf("target account %d is inactive", targetAccID))
	}
	if toAcc.OwnerID != toCustomerID {
		return apperror.NewAuthError("receiver does not own the target account")
	}
	if err := acc.WithdrawMoney(fromCustomerID, amount); err != nil {
		return err
	}
	if err := toAcc.DepositMoney(toCustomerID, amount); err != nil {
		_ = acc.DepositMoney(fromCustomerID, amount) // rollback
		return err
	}
	return nil
}

func TransferMoneyInternally(fromAccountID, toAccountID int, amount float64) error {
	fromAcc, err := GetAccountById(fromAccountID)
	if err != nil {
		return err
	}
	toAcc, err := GetAccountById(toAccountID)
	if err != nil {
		return err
	}
	if fromAcc.OwnerID != toAcc.OwnerID {
		return apperror.NewAuthError("accounts belong to different owners")
	}
	if !fromAcc.IsActive || !toAcc.IsActive {
		return apperror.NewAccountError("transfer", "both accounts must be active")
	}
	if err := fromAcc.WithdrawMoney(fromAcc.OwnerID, amount); err != nil {
		return err
	}
	if err := toAcc.DepositMoney(toAcc.OwnerID, amount); err != nil {
		_ = fromAcc.DepositMoney(fromAcc.OwnerID, amount) // rollback
		return err
	}
	return nil
}
