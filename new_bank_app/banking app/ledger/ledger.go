package customer

import (
	"banking-app/apperror"
	"banking-app/account"
	"banking-app/bank"
	"errors"
	"fmt"
	"sync"
)

// Customer struct with fields
type Customer struct {
	CustomerID int
	FirstName  string
	LastName   string
	IsAdmin    bool
	Accounts   map[int]*account.Account // accountID -> Account pointer
}

// CustomerManager manages customers and banks
type CustomerManager struct {
	customers map[int]*Customer
	banks     map[int]*bank.Bank
	mu        sync.RWMutex
	idCounter int
}

// NewCustomerManager creates a new CustomerManager instance
func NewCustomerManager() *CustomerManager {
	return &CustomerManager{
		customers: make(map[int]*Customer),
		banks:     make(map[int]*bank.Bank),
		idCounter: 1000,
	}
}

// Generates a new unique customer ID
func (cm *CustomerManager) generateCustomerID() int {
	cm.idCounter++
	return cm.idCounter
}

// ===================== Factory Methods ====================

// newCustomer factory method to create a Customer with account
func (cm *CustomerManager) newCustomer(firstName, lastName string, isAdmin bool, bankAccount *account.Account) (*Customer, *apperror.ValidationError) {
	firstName = TrimAndValidateName(firstName)
	lastName = TrimAndValidateName(lastName)

	if firstName == "" || lastName == "" {
		return nil, apperror.NewValidationError("name", "first or last name cannot be empty")
	}

	custID := cm.generateCustomerID()
	c := &Customer{
		CustomerID: custID,
		FirstName:  firstName,
		LastName:   lastName,
		IsAdmin:    isAdmin,
		Accounts:   make(map[int]*account.Account),
	}

	if bankAccount != nil {
		c.Accounts[bankAccount.AccountID] = bankAccount
	}

	cm.mu.Lock()
	cm.customers[custID] = c
	cm.mu.Unlock()

	return c, nil
}

// NewAdmin factory method
func (cm *CustomerManager) NewAdmin(firstName, lastName string) *Customer {
	c, _ := cm.newCustomer(firstName, lastName, true, nil)
	return c
}

// ===================== Bank Management ====================

func (cm *CustomerManager) CreateNewBank(fullname string) (*bank.Bank, *apperror.ValidationError) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	id := cm.generateCustomerID()
	b, err := bank.NewBank(id, fullname)
	if err != nil {
		return nil, err
	}
	cm.banks[id] = b
	return b, nil
}

func (cm *CustomerManager) UpdateBankName(bankID int, newName string) *apperror.ValidationError {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	b, ok := cm.banks[bankID]
	if !ok {
		return apperror.NewNotFoundError("bank", bankID)
	}
	return b.UpdateBankName(newName)
}

func (cm *CustomerManager) GetBankById(id int) *bank.Bank {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.banks[id]
}

func (cm *CustomerManager) GetAllBanks() []bank.Bank {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	banks := make([]bank.Bank, 0, len(cm.banks))
	for _, b := range cm.banks {
		banks = append(banks, *b)
	}
	return banks
}

func (cm *CustomerManager) DeleteBank(bankID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.banks, bankID)
}

// ===================== Customer Management ====================

func (cm *CustomerManager) CreateNewCustomer(firstName, lastName string, bankID int) (*Customer, *apperror.ValidationError) {
	cm.mu.RLock()
	b, bankExists := cm.banks[bankID]
	cm.mu.RUnlock()
	if !bankExists {
		return nil, apperror.NewNotFoundError("bank", bankID)
	}

	accID := cm.generateCustomerID() // reuse for account id
	acc, err := account.NewAccount(accID, 0, b.BankID) // OwnerID will be customerID after creation, set 0 now temporarily
	if err != nil {
		return nil, err
	}

	c, errVal := cm.newCustomer(firstName, lastName, false, acc)
	if errVal != nil {
		return nil, errVal
	}

	// Update OwnerID of account now that CustomerID is known
	acc.OwnerID = c.CustomerID

	cm.mu.Lock()
	c.Accounts[acc.AccountID] = acc
	cm.customers[c.CustomerID] = c
	cm.mu.Unlock()

	return c, nil
}

func (cm *CustomerManager) GetAllCustomers() []Customer {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	customers := make([]Customer, 0, len(cm.customers))
	for _, c := range cm.customers {
		customers = append(customers, *c)
	}
	return customers
}

func (cm *CustomerManager) GetCustomerById(customerID int) *Customer {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.customers[customerID]
}

func (cm *CustomerManager) DeleteCustomer(customerID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.customers, customerID)
}

func (cm *CustomerManager) DeleteCustomerAccountById(customerID, accountID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	c, ok := cm.customers[customerID]
	if !ok {
		return
	}
	delete(c.Accounts, accountID)
}

// ===================== Transaction Operations ====================

func (cm *CustomerManager) DepositMoney(amount float64, accountID int) error {
	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		return err
	}
	return acc.DepositMoney(acc.OwnerID, amount)
}

func (cm *CustomerManager) WithDrawMoney(amount float64, accountID int) error {
	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		return err
	}
	return acc.WithdrawMoney(acc.OwnerID, amount)
}

func (cm *CustomerManager) WithDrawMoneyByAccount_Id(amount float64, accountID int) *apperror.ValidationError {
	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		return err
	}
	err2 := acc.WithdrawMoney(acc.OwnerID, amount)
	if err2 != nil {
		return apperror.NewAccountError("withdraw", err2.Error())
	}
	return nil
}

func (cm *CustomerManager) TransferMoney_To_External(amount float64, fromCustomerID, toCustomerID, fromAccountID, toAccountID int) error {
	fromCust := cm.GetCustomerById(fromCustomerID)
	toCust := cm.GetCustomerById(toCustomerID)
	if fromCust == nil || toCust == nil {
		return errors.New("customer not found")
	}

	fromAcc, err := cm.GetAccountById(fromAccountID)
	if err != nil {
		return err
	}
	toAcc, err := cm.GetAccountById(toAccountID)
	if err != nil {
		return err
	}

	if fromAcc.OwnerID != fromCustomerID || toAcc.OwnerID != toCustomerID {
		return errors.New("account ownership mismatch")
	}

	if err := fromAcc.WithdrawMoney(fromCustomerID, amount); err != nil {
		return err
	}
	if err := toAcc.DepositMoney(toCustomerID, amount); err != nil {
		// rollback withdrawal on error
		_ = fromAcc.DepositMoney(fromCustomerID, amount)
		return err
	}
	return nil
}

func (cm *CustomerManager) TransferMoneyInternally(fromAccountID, toAccountID int, amount float64) error {
	fromAcc, err := cm.GetAccountById(fromAccountID)
	if err != nil {
		return err
	}
	toAcc, err := cm.GetAccountById(toAccountID)
	if err != nil {
		return err
	}

	if fromAcc.OwnerID != toAcc.OwnerID {
		return errors.New("accounts belong to different customers")
	}

	if err := fromAcc.WithdrawMoney(fromAcc.OwnerID, amount); err != nil {
		return err
	}
	if err := toAcc.DepositMoney(toAcc.OwnerID, amount); err != nil {
		// rollback withdrawal on error
		_ = fromAcc.DepositMoney(fromAcc.OwnerID, amount)
		return err
	}
	return nil
}

// ===================== Passbook / Balance / Account ====================

// Placeholder: Implement transactions package & store if needed.
func (cm *CustomerManager) GetPassBook_ById(customerID, accountID, pageNo int) []interface{} {
	// Return empty for now or implement with transaction storage
	return nil
}

func (cm *CustomerManager) GetTotalBalanceBy_Customer_Id(customerID int) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	c, ok := cm.customers[customerID]
	if !ok {
		return 0
	}
	total := 0.0
	for _, acc := range c.Accounts {
		total += acc.Balance
	}
	return total
}

func (cm *CustomerManager) GetTotalBalance() float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	total := 0.0
	for _, c := range cm.customers {
		for _, acc := range c.Accounts {
			total += acc.Balance
		}
	}
	return total
}

func (cm *CustomerManager) GetAccount_BalanceBy_Id(accountID int) float64 {
	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		return 0
	}
	return acc.Balance
}

func (cm *CustomerManager) GetAccountById(accountID int) (*account.Account, *apperror.ValidationError) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, c := range cm.customers {
		if acc, ok := c.Accounts[accountID]; ok {
			return acc, nil
		}
	}
	return nil, apperror.NewNotFoundError("account", accountID)
}

func (cm *CustomerManager) AddNewAccount(bankID int) (*account.Account, *apperror.ValidationError) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// No current customer context here - you might want to pass customerID for ownership
	return nil, apperror.NewValidationError("AddNewAccount", "customer ID required")
}

// ===================== Account / Customer Updates ====================

func (cm *CustomerManager) DeleteAccountById(accountID int) *apperror.ValidationError {
	cm.mu.Lock()
	defer cm
