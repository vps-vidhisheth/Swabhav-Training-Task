package customer

import (
	"banking-app/account"
	"banking-app/apperror"
	"banking-app/bank"
	"errors"
	"fmt"
)

type Customer struct {
	CustomerID int
	FirstName  string
	LastName   string
	IsAdmin    bool
	IsActive   bool
	Accounts   map[int]*account.Account
}

type CustomerManager struct {
	customers map[int]*Customer
	banks     map[int]*bank.Bank
	idCounter int
}

func NewCustomerManager() *CustomerManager {
	return &CustomerManager{
		customers: make(map[int]*Customer),
		banks:     make(map[int]*bank.Bank),
		idCounter: 1000,
	}
}

func (cm *CustomerManager) generateCustomerID() int {
	cm.idCounter++
	return cm.idCounter
}

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
		IsActive:   true,
		Accounts:   make(map[int]*account.Account),
	}

	if bankAccount != nil {
		bankAccount.IsActive = true
		c.Accounts[bankAccount.AccountID] = bankAccount
	}

	cm.customers[custID] = c
	return c, nil
}

func (cm *CustomerManager) NewAdmin(firstName, lastName string) *Customer {
	c, _ := cm.newCustomer(firstName, lastName, true, nil)
	return c
}

func (cm *CustomerManager) CreateNewBank(fullname string) (*bank.Bank, *apperror.ValidationError) {
	id := cm.generateCustomerID()
	b, err := bank.NewBank(id, fullname)
	if err != nil {
		return nil, err
	}
	cm.banks[id] = b
	return b, nil
}

func (cm *CustomerManager) UpdateBankName(bankID int, newName string) *apperror.ValidationError {
	b, ok := cm.banks[bankID]
	if !ok {
		return apperror.NewValidationError("bank", fmt.Sprintf("bank ID %d not found", bankID))
	}
	return b.UpdateBankName(newName)
}

func (cm *CustomerManager) GetBankById(id int) *bank.Bank {
	return cm.banks[id]
}

func (cm *CustomerManager) GetAllBanks() []bank.Bank {
	banks := make([]bank.Bank, 0, len(cm.banks))
	for _, b := range cm.banks {
		banks = append(banks, *b)
	}
	return banks
}

func (cm *CustomerManager) DeleteBank(bankID int) {
	delete(cm.banks, bankID)
}

func (cm *CustomerManager) CreateNewCustomer(firstName, lastName string, bankID int) (*Customer, *apperror.ValidationError) {
	b, bankExists := cm.banks[bankID]
	if !bankExists {
		return nil, apperror.NewValidationError("bank", fmt.Sprintf("bank ID %d not found", bankID))
	}

	accID := cm.generateCustomerID()
	acc, err := account.NewAccount(accID, 0, b.BankID)
	if err != nil {
		if valErr, ok := err.(*apperror.ValidationError); ok {
			return nil, valErr
		}
		return nil, apperror.NewValidationError("account", err.Error())
	}

	c, errVal := cm.newCustomer(firstName, lastName, false, acc)
	if errVal != nil {
		return nil, errVal
	}
	acc.OwnerID = c.CustomerID
	acc.IsActive = true

	c.Accounts[acc.AccountID] = acc
	cm.customers[c.CustomerID] = c
	return c, nil
}

func (cm *CustomerManager) GetAllCustomers() []Customer {
	customers := make([]Customer, 0, len(cm.customers))
	for _, c := range cm.customers {
		if c.IsActive {
			customers = append(customers, *c)
		}
	}
	return customers
}

func (cm *CustomerManager) GetCustomerById(customerID int) *Customer {
	c := cm.customers[customerID]
	if c != nil && c.IsActive {
		return c
	}
	return nil
}

func (cm *CustomerManager) DeleteCustomer(customerID int) {
	c := cm.customers[customerID]
	if c != nil {
		c.IsActive = false
		for _, acc := range c.Accounts {
			acc.IsActive = false
		}
	}
}

func (cm *CustomerManager) DeleteCustomerAccountById(customerID, accountID int) {
	c := cm.customers[customerID]
	if c == nil {
		return
	}
	if acc, ok := c.Accounts[accountID]; ok {
		acc.IsActive = false
	}
}

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

func (cm *CustomerManager) WithDrawMoneyByAccount_Id(amount float64, accountID int) error {
	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		return err
	}
	err2 := acc.WithdrawMoney(acc.OwnerID, amount)
	if err2 != nil {
		return apperror.NewAccountError("withdraw", "failed to withdraw money", err2)
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
		_ = fromAcc.DepositMoney(fromAcc.OwnerID, amount)
		return err
	}
	return nil
}

func (cm *CustomerManager) GetPassBook_ById(customerID, accountID, pageNo int) []interface{} {
	return nil
}

func (cm *CustomerManager) GetTotalBalanceBy_Customer_Id(customerID int) float64 {
	c := cm.customers[customerID]
	if c == nil || !c.IsActive {
		return 0
	}
	total := 0.0
	for _, acc := range c.Accounts {
		if acc.IsActive {
			total += acc.Balance
		}
	}
	return total
}

func (cm *CustomerManager) GetTotalBalance() float64 {
	total := 0.0
	for _, c := range cm.customers {
		if !c.IsActive {
			continue
		}
		for _, acc := range c.Accounts {
			if acc.IsActive {
				total += acc.Balance
			}
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

func (cm *CustomerManager) GetAccountById(accountID int) (*account.Account, error) {
	for _, c := range cm.customers {
		if !c.IsActive {
			continue
		}
		if acc, ok := c.Accounts[accountID]; ok && acc.IsActive {
			return acc, nil
		}
	}
	return nil, apperror.NewNotFoundError("account", accountID)
}

func (cm *CustomerManager) AddNewAccount(bankID int) (*account.Account, *apperror.ValidationError) {
	return nil, apperror.NewValidationError("AddNewAccount", "customer ID required")
}

func (cm *CustomerManager) DeleteAccountById(accountID int) error {
	for _, c := range cm.customers {
		if acc, ok := c.Accounts[accountID]; ok {
			acc.IsActive = false
			return nil
		}
	}
	return apperror.NewNotFoundError("account", accountID)
}

func (cm *CustomerManager) UpdateAccount(accountID int, newBalance float64) error {
	for _, c := range cm.customers {
		if acc, ok := c.Accounts[accountID]; ok && acc.IsActive {
			acc.Balance = newBalance
			return nil
		}
	}
	return apperror.NewNotFoundError("account", accountID)
}

func (cm *CustomerManager) UpdateCustomer(customerID int, firstName, lastName string) error {
	c, ok := cm.customers[customerID]
	if !ok || !c.IsActive {
		return apperror.NewNotFoundError("customer", customerID)
	}

	if firstName != "" {
		c.FirstName = firstName
	}
	if lastName != "" {
		c.LastName = lastName
	}
	return nil
}

func (cm *CustomerManager) UpdateCustomerNameById(customerID int, firstName, lastName string) error {
	return cm.UpdateCustomer(customerID, firstName, lastName)
}

func TrimAndValidateName(name string) string {
	return name
}
