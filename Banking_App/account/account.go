package account

import (
	"banking-app/apperror"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/helper"
)

type Account struct {
	AccountID    int
	Balance      float64
	IsActive     bool
	Bank         *bank.Bank
	Customer     *customer.Customer
	Transactions []Transaction
}

type Transaction struct {
	TransactionID int
	Type          string
	Amount        float64
}

var (
	accounts           = make(map[int]*Account)
	accountIDCounter   int
	transactionCounter int
)

// --- Account Creation ---

func CreateAccount(cust *customer.Customer, bank *bank.Bank) (*Account, error) {
	if cust == nil || !cust.IsActive {
		return nil, apperror.NewCustomerError("create account", "invalid customer")
	}
	if bank == nil || !bank.IsActive {
		return nil, apperror.NewBankError("create account", "invalid bank")
	}
	accountIDCounter++
	acc := &Account{
		AccountID: accountIDCounter,
		Balance:   1000,
		IsActive:  true,
		Customer:  cust,
		Bank:      bank,
	}
	acc.addTransaction("deposit", 1000)
	accounts[acc.AccountID] = acc
	cust.TotalBalance += 1000
	return acc, nil
}

func CreateAccountAuthorized(requester interface {
	IsAdminUser() bool
	IsActiveUser() bool
}, cust *customer.Customer, bank *bank.Bank) (*Account, error) {
	if !requester.IsActiveUser() || !requester.IsAdminUser() {
		return nil, apperror.NewAuthError("create account")
	}
	return CreateAccount(cust, bank)
}

// --- Getters (Admin only) ---

func GetAccountByID(accountID int, isAdmin bool) (*Account, error) {
	if !isAdmin {
		return nil, apperror.NewAuthError("only admin can access account by ID")
	}
	acc, exists := accounts[accountID]
	if !exists || !acc.IsActive {
		return nil, apperror.NewNotFoundError("account", accountID)
	}
	return acc, nil
}

func GetCustomerAccounts(customerID int, isAdmin bool) ([]*Account, error) {
	if !isAdmin {
		return nil, apperror.NewAuthError("only admin can access customer accounts")
	}
	var result []*Account
	for _, acc := range accounts {
		if acc.Customer.CustomerID == customerID && acc.IsActive {
			result = append(result, acc)
		}
	}
	if len(result) == 0 {
		return nil, apperror.NewNotFoundError("customer accounts", customerID)
	}
	return result, nil
}

func GetBankAccounts(bankID int, isAdmin bool) ([]*Account, error) {
	if !isAdmin {
		return nil, apperror.NewAuthError("only admin can access bank accounts")
	}
	var result []*Account
	for _, acc := range accounts {
		if acc.Bank.BankID == bankID && acc.IsActive {
			result = append(result, acc)
		}
	}
	if len(result) == 0 {
		return nil, apperror.NewNotFoundError("bank accounts", bankID)
	}
	return result, nil
}

func GetAllAccountsPaginated(page, pageSize int, isAdmin bool) ([]*Account, error) {
	if !isAdmin {
		return nil, apperror.NewAuthError("only admin can access all accounts")
	}
	var activeAccounts []*Account
	for _, acc := range accounts {
		if acc.IsActive {
			activeAccounts = append(activeAccounts, acc)
		}
	}
	start, end := helper.PaginationBounds(page, pageSize, len(activeAccounts))
	return activeAccounts[start:end], nil
}

// --- Deposit / Withdraw / Transfer ---

func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return apperror.NewValidationError("amount", "must be greater than 0")
	}
	a.Balance += amount
	a.Customer.TotalBalance += amount
	a.addTransaction("deposit", amount)
	return nil
}

func (a *Account) Withdraw(amount float64) error {
	if amount <= 0 || a.Balance < amount {
		return apperror.NewAccountError("withdraw", "invalid amount or insufficient funds")
	}
	a.Balance -= amount
	a.Customer.TotalBalance -= amount
	a.addTransaction("withdraw", amount)
	return nil
}

// --- Customer-Initiated Internal Transfer ---

func (a *Account) TransferToOwnAccount(to *Account, amount float64) error {
	if a.AccountID == to.AccountID {
		return apperror.NewValidationError("transfer", "cannot transfer to same account")
	}
	if a.Customer.CustomerID != to.Customer.CustomerID {
		return apperror.NewValidationError("transfer", "can only transfer between your own accounts")
	}
	if amount <= 0 || a.Balance < amount {
		return apperror.NewAccountError("transfer", "invalid amount or insufficient balance")
	}
	a.Balance -= amount
	a.Customer.TotalBalance -= amount
	to.Balance += amount
	to.Customer.TotalBalance += amount
	a.addTransaction("transfer_out", amount)
	to.addTransaction("transfer_in", amount)
	return nil
}

// --- Admin-Initiated Transfer (between any accounts) ---

func AdminTransfer(from *Account, to *Account, amount float64) error {
	if from.AccountID == to.AccountID {
		return apperror.NewValidationError("transfer", "cannot transfer to same account")
	}
	if amount <= 0 || from.Balance < amount {
		return apperror.NewAccountError("transfer", "invalid amount or insufficient balance")
	}
	from.Balance -= amount
	from.Customer.TotalBalance -= amount
	to.Balance += amount
	to.Customer.TotalBalance += amount
	from.addTransaction("transfer_out", amount)
	to.addTransaction("transfer_in", amount)
	return nil
}

func (a *Account) addTransaction(tType string, amt float64) {
	transactionCounter++
	a.Transactions = append(a.Transactions, Transaction{
		TransactionID: transactionCounter,
		Type:          tType,
		Amount:        amt,
	})
}

func (a *Account) GetPassbookPaginated(page, pageSize int) []Transaction {
	total := len(a.Transactions)
	start, end := helper.PaginationBounds(page, pageSize, total)
	return a.Transactions[start:end]
}

// --- Admin-only Update / Delete ---

func (a *Account) updateBalance(value interface{}) error {
	v, ok := value.(float64)
	if !ok || v < 0 {
		return apperror.NewValidationError("balance", "must be a non-negative float")
	}
	diff := v - a.Balance
	a.Balance = v
	a.Customer.TotalBalance += diff
	return nil
}

func (a *Account) updateIsActive(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	a.IsActive = v
	return nil
}

func (a *Account) UpdateAccount(field string, value interface{}) error {
	switch field {
	case "balance":
		return a.updateBalance(value)
	case "isactive":
		return a.updateIsActive(value)
	default:
		return apperror.NewValidationError("field", "unknown update field")
	}
}

func UpdateAccountByID(requester interface {
	IsAdminUser() bool
	IsActiveUser() bool
}, accountID int, field string, value interface{}) error {
	if !requester.IsActiveUser() || !requester.IsAdminUser() {
		return apperror.NewAuthError("update account")
	}
	acc, err := GetAccountByID(accountID, true)
	if err != nil {
		return err
	}
	return acc.UpdateAccount(field, value)
}

func (a *Account) SoftDelete() error {
	if !a.Bank.IsActive {
		return apperror.NewBankError("delete account", "cannot delete account from inactive bank")
	}
	a.IsActive = false
	a.Customer.TotalBalance -= a.Balance
	return nil
}

func SoftDeleteAccountByID(requester interface {
	IsAdminUser() bool
	IsActiveUser() bool
}, accountID int) error {
	if !requester.IsActiveUser() || !requester.IsAdminUser() {
		return apperror.NewAuthError("delete account")
	}
	acc, err := GetAccountByID(accountID, true)
	if err != nil {
		return err
	}
	return acc.SoftDelete()
}

// --- Bank Cleanup Helper ---
func HasActiveAccounts(bankID int) (bool, error) {
	if bankID <= 0 {
		return false, apperror.NewValidationError("bankID", "must be a positive integer")
	}

	for _, acc := range accounts {
		if acc.Bank.BankID == bankID && acc.IsActive {
			return true, nil
		}
	}
	return false, nil
}
