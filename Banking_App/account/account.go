package account

import (
	"banking-app/apperror"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/helper"
	"banking-app/ledger"
	"fmt"
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

func safely(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal panic: %v", r)
		}
	}()
	return fn()
}

func CreateAccount(caller *customer.Customer, cust *customer.Customer, b *bank.Bank) (*Account, error) {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return nil, apperror.NewAuthError("create account: caller not authorized")
	}
	if cust == nil || !cust.IsActive {
		return nil, apperror.NewCustomerError("create account: invalid or inactive customer", "")
	}
	if b == nil || !b.IsActive {
		return nil, apperror.NewBankError("create account: invalid or inactive bank", "")
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

// CreateAccountByIDs now explicitly accepts a bank.BankManager and customer.Manager
// to retrieve bank and customer details, respectively.
func CreateAccountByIDs(caller *customer.Customer, customerID, bankID int, bankManager *bank.BankManager, customerManager *customer.Manager) (*Account, error) {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return nil, apperror.NewAuthError("create account by IDs: caller not authorized")
	}

	// Use the provided customerManager to get the customer.
	if customerManager == nil {
		return nil, fmt.Errorf("create account by IDs: customer manager is nil")
	}
	// Pass the caller to GetByID if it performs authorization checks.
	// If GetByID is strictly for internal lookup and doesn't need caller auth, 'nil' could be passed.
	// Based on customer.go, GetByID checks if caller is nil, so it's safe to pass caller.
	cust, err := customerManager.GetByID(caller, customerID)
	if err != nil {
		return nil, err
	}

	// Use the provided bankManager to get the bank.
	if bankManager == nil {
		return nil, fmt.Errorf("create account by IDs: bank manager is nil")
	}
	b, err := bankManager.GetBank(caller, bankID)
	if err != nil {
		return nil, err
	}
	return CreateAccount(caller, cust, b)
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
		if amount <= 0 {
			return apperror.NewValidationError("amount", "must be greater than 0")
		}
		if acc.Balance < amount {
			return apperror.NewAccountError("withdraw", "insufficient funds")
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

func (acc *Account) TransferToOwnAccount(to *Account, amount float64) error {
	return safely(func() error {
		if acc == nil {
			return fmt.Errorf("transfer source account is nil")
		}
		if to == nil {
			return fmt.Errorf("transfer destination account is nil")
		}
		if !acc.IsValid() || !to.IsValid() {
			return apperror.NewAuthError("transfer: source or destination account invalid or inactive")
		}
		if acc.Customer.CustomerID != to.Customer.CustomerID {
			return apperror.NewAuthError("transfer: unauthorized account transfer to a different customer")
		}
		if acc.AccountID == to.AccountID {
			return apperror.NewValidationError("transfer", "cannot transfer to the same account")
		}
		if amount <= 0 {
			return apperror.NewValidationError("amount", "must be greater than 0")
		}
		if acc.Balance < amount {
			return apperror.NewAccountError("transfer", "insufficient balance in source account")
		}

		acc.Balance -= amount
		to.Balance += amount
		acc.addTransaction("transfer_out", amount)
		to.addTransaction("transfer_in", amount)
		return nil
	})
}

func AdminTransfer(requester *customer.Customer, from, to *Account, amount float64, l *ledger.InterbankLedger) error {
	return safely(func() error {
		if requester == nil || !requester.IsActive || !requester.IsAdmin {
			return apperror.NewAuthError("admin transfer: requester not authorized")
		}
		if from == nil {
			return fmt.Errorf("admin transfer: source account is nil")
		}
		if to == nil {
			return fmt.Errorf("admin transfer: destination account is nil")
		}
		if amount <= 0 {
			return apperror.NewValidationError("amount", "must be greater than 0")
		}
		if from.Balance < amount {
			return apperror.NewAccountError("admin transfer", "insufficient funds in source account")
		}

		from.Balance -= amount
		from.Customer.TotalBalance -= amount
		to.Balance += amount
		to.Customer.TotalBalance += amount
		from.addTransaction("transfer_out", amount)
		to.addTransaction("transfer_in", amount)

		if l != nil && from.Bank != nil && to.Bank != nil &&
			from.Bank.BankID != to.Bank.BankID {
			l.RecordTransfer(from.Bank.BankID, to.Bank.BankID, amount)
		}
		return nil
	})
}

func GetAccountByID(caller *customer.Customer, accountID int) (*Account, error) {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return nil, apperror.NewAuthError("get account by ID: caller not authorized")
	}

	acc, ok := accounts[accountID]
	if !ok || !acc.IsActive {
		return nil, apperror.NewNotFoundError("account", accountID)
	}
	return acc, nil
}

func GetCustomerAccounts(caller *customer.Customer, customerID int) ([]*Account, error) {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return nil, apperror.NewAuthError("get customer accounts: caller not authorized")
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

func GetBankAccounts(caller *customer.Customer, bankID int) ([]*Account, error) {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return nil, apperror.NewAuthError("get bank accounts: caller not authorized")
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

func GetAllAccountsPaginated(caller *customer.Customer, page, pageSize int) ([]*Account, error) {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return nil, apperror.NewAuthError("get all accounts: caller not authorized")
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

func UpdateAccountByID(caller *customer.Customer, accountID int, field string, value interface{}) error {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return apperror.NewAuthError("update account by ID: caller not authorized")
	}
	acc, err := GetAccountByID(caller, accountID)
	if err != nil {
		return err
	}
	return acc.UpdateAccount(field, value)
}

func (acc *Account) SoftDelete(caller *customer.Customer) error {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return apperror.NewAuthError("soft delete account: caller not authorized")
	}
	if acc == nil {
		return fmt.Errorf("soft delete: account is nil")
	}

	if !acc.IsActive {
		return fmt.Errorf("soft delete: account already inactive (double deletion detected)")
	}
	if acc.Bank == nil || !acc.Bank.IsActive {
		return apperror.NewBankError("soft delete", "associated bank is invalid or inactive")
	}
	acc.IsActive = false
	acc.Customer.TotalBalance -= acc.Balance
	return nil
}

func SoftDeleteAccountByID(caller *customer.Customer, accountID int) error {
	if caller == nil || !caller.IsActive || !caller.IsAdmin {
		return apperror.NewAuthError("soft delete account by ID: caller not authorized")
	}
	acc, err := GetAccountByID(caller, accountID)
	if err != nil {
		return err
	}
	return acc.SoftDelete(caller)
}

func HasActiveAccounts(bankID int) (bool, error) {
	if bankID <= 0 {
		return false, apperror.NewValidationError("bankID", "must be positive")
	}

	for _, acc := range accounts {
		if acc.Bank != nil && acc.Bank.BankID == bankID && acc.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func GetBankTotalBalance(bankID int) (float64, error) {
	var total float64

	for _, acc := range accounts {
		if acc.Bank != nil && acc.Bank.BankID == bankID && acc.IsActive {
			total += acc.Balance
		}
	}
	return total, nil
}
