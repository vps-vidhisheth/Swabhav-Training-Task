package account

import (
	"banking-app/apperror"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/helper"
	"banking-app/ledger"
	"fmt"
)

// --- Types ---

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

// --- Internal Storage ---

var (
	accounts           = make(map[int]*Account)
	accountIDCounter   int
	transactionCounter int
	interbankLedger    *ledger.InterbankLedger
)

// --- Lookup Injectors ---

var customerLookup func(int) (*customer.Customer, error)
var bankLookup func(int) (*bank.Bank, error)

func SetCustomerLookup(fn func(int) (*customer.Customer, error)) {
	customerLookup = fn
}

func SetBankLookup(fn func(int) (*bank.Bank, error)) {
	bankLookup = fn
}

// --- Ledger Setter ---

func SetLedger(l *ledger.InterbankLedger) {
	interbankLedger = l
}

// --- Recovery Helper ---

func safely(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal panic: %v", r)
		}
	}()
	return fn()
}

// --- Account Creation ---

func CreateAccount(cust *customer.Customer, b *bank.Bank) (*Account, error) {
	if cust == nil || !cust.IsActive {
		return nil, apperror.NewCustomerError("create account", "invalid customer")
	}
	if b == nil || !b.IsActive {
		return nil, apperror.NewBankError("create account", "invalid bank")
	}

	accountIDCounter++
	acc := &Account{
		AccountID: accountIDCounter,
		Balance:   1000,
		IsActive:  true,
		Customer:  cust,
		Bank:      b,
	}
	acc.addTransaction("deposit", 1000)
	accounts[acc.AccountID] = acc
	cust.TotalBalance += 1000
	return acc, nil
}

func CreateAccountByIDs(customerID, bankID int) (*Account, error) {
	if customerLookup == nil || bankLookup == nil {
		return nil, fmt.Errorf("lookup functions not initialized")
	}
	cust, err := customerLookup(customerID)
	if err != nil {
		return nil, err
	}
	b, err := bankLookup(bankID)
	if err != nil {
		return nil, err
	}
	return CreateAccount(cust, b)
}

func (acc *Account) IsValid() bool {
	return acc != nil && acc.IsActive && acc.Customer != nil && acc.Customer.IsActive && acc.Bank != nil && acc.Bank.IsActive
}

func (acc *Account) Deposit(amount float64) error {
	return safely(func() error {
		if !acc.IsValid() {
			return apperror.NewAccountError("deposit", "invalid or inactive account")
		}
		if amount <= 0 {
			return apperror.NewValidationError("amount", "must be greater than 0")
		}
		acc.Balance += amount
		acc.Customer.TotalBalance += amount
		acc.addTransaction("deposit", amount)
		return nil
	})
}

func (acc *Account) Withdraw(amount float64) error {
	return safely(func() error {
		if !acc.IsValid() {
			return apperror.NewAccountError("withdraw", "invalid or inactive account")
		}
		if amount <= 0 || acc.Balance < amount {
			return apperror.NewAccountError("withdraw", "invalid amount or insufficient funds")
		}
		acc.Balance -= amount
		acc.Customer.TotalBalance -= amount
		acc.addTransaction("withdraw", amount)
		return nil
	})
}

func (acc *Account) addTransaction(tType string, amt float64) {
	transactionCounter++
	tx := Transaction{
		TransactionID: transactionCounter,
		Type:          tType,
		Amount:        amt,
	}
	acc.Transactions = append(acc.Transactions, tx)
}

func (acc *Account) GetPassbookPaginated(page, pageSize int) []Transaction {
	total := len(acc.Transactions)
	start, end := helper.PaginationBounds(page, pageSize, total)
	return acc.Transactions[start:end]
}

// --- Transfers ---

func (acc *Account) TransferToOwnAccount(to *Account, amount float64) error {
	return safely(func() error {
		if acc == nil || to == nil {
			panic("nil account in transfer")
		}
		if !acc.IsValid() || !to.IsValid() {
			return apperror.NewAuthError("transfer from/to invalid or inactive account")
		}
		if acc.Customer.CustomerID != to.Customer.CustomerID {
			return apperror.NewAuthError("unauthorized account transfer")
		}
		if acc.AccountID == to.AccountID {
			return apperror.NewValidationError("transfer", "cannot transfer to same account")
		}
		if amount <= 0 || acc.Balance < amount {
			return apperror.NewAccountError("transfer", "invalid amount or insufficient balance")
		}
		acc.Balance -= amount
		to.Balance += amount
		acc.addTransaction("transfer_out", amount)
		to.addTransaction("transfer_in", amount)
		return nil
	})
}

// Admin transfer: requester is customer (admin or authorized user)
func AdminTransfer(requester *customer.Customer, from, to *Account, amount float64) error {
	return safely(func() error {
		if requester == nil || from == nil || to == nil {
			panic("nil parameter in admin transfer")
		}
		if !requester.IsActive {
			return apperror.NewAuthError("inactive customer")
		}
		if from.Customer.CustomerID != requester.CustomerID {
			return apperror.NewAuthError("unauthorized source account")
		}
		if amount <= 0 || from.Balance < amount {
			return apperror.NewAccountError("transfer", "invalid amount or insufficient funds")
		}
		from.Balance -= amount
		from.Customer.TotalBalance -= amount
		to.Balance += amount
		to.Customer.TotalBalance += amount
		from.addTransaction("transfer_out", amount)
		to.addTransaction("transfer_in", amount)

		if interbankLedger != nil && from.Bank != nil && to.Bank != nil &&
			from.Bank.BankID != to.Bank.BankID {
			interbankLedger.UpdateLedgerFromTransaction(from.Bank.BankID, to.Bank.BankID, amount)
		}
		return nil
	})
}

// --- Getters ---

func GetAccountByID(accountID int, isAdmin bool) (*Account, error) {
	if !isAdmin {
		return nil, apperror.NewAuthError("only admin can access account by ID")
	}
	acc, ok := accounts[accountID]
	if !ok || !acc.IsActive {
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

// --- Updates / Deletes ---

func (acc *Account) updateIsActive(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	acc.IsActive = v
	return nil
}

func (acc *Account) UpdateAccount(field string, value interface{}) error {
	switch field {
	case "isactive":
		return acc.updateIsActive(value)
	default:
		return apperror.NewValidationError("field", "unknown update field")
	}
}

func UpdateAccountByID(requester interface {
	IsAdminUser() bool
	IsActiveUser() bool
}, accountID int, field string, value interface{}) error {
	if !requester.IsAdminUser() || !requester.IsActiveUser() {
		return apperror.NewAuthError("update account")
	}
	acc, err := GetAccountByID(accountID, true)
	if err != nil {
		return err
	}
	return acc.UpdateAccount(field, value)
}

func (acc *Account) SoftDelete() error {
	if acc == nil {
		panic("delete nil account")
	}
	if !acc.IsActive {
		panic("double deletion detected")
	}
	if !acc.Bank.IsActive {
		return apperror.NewBankError("delete", "bank is inactive")
	}
	acc.IsActive = false
	acc.Customer.TotalBalance -= acc.Balance
	return nil
}

func SoftDeleteAccountByID(requester interface {
	IsAdminUser() bool
	IsActiveUser() bool
}, accountID int) error {
	if !requester.IsAdminUser() || !requester.IsActiveUser() {
		return apperror.NewAuthError("delete account")
	}
	acc, err := GetAccountByID(accountID, true)
	if err != nil {
		return err
	}
	return acc.SoftDelete()
}

// --- Helpers ---

func HasActiveAccounts(bankID int) (bool, error) {
	if bankID <= 0 {
		return false, apperror.NewValidationError("bankID", "must be positive")
	}
	for _, acc := range accounts {
		if acc.Bank.BankID == bankID && acc.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func GetBankTotalBalance(bankID int) (float64, error) {
	var total float64
	for _, acc := range accounts {
		if acc.Bank.BankID == bankID && acc.IsActive {
			total += acc.Balance
		}
	}
	return total, nil
}
