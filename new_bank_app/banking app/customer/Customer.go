package customer

import (
	"banking-app/account"
	"banking-app/apperror"
	"banking-app/bank"
	"banking-app/ledger"
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
	ledger    *ledger.Ledger
	idCounter int
	admin     *Customer
}

func NewCustomerManager(firstName, lastName string) *CustomerManager {
	cm := &CustomerManager{
		customers: make(map[int]*Customer),
		banks:     make(map[int]*bank.Bank),
		idCounter: 1000,
	}

	cm.ledger = ledger.NewLedger(func(bankID int) (float64, error) {
		var total float64
		for _, c := range cm.customers {
			if !c.IsActive {
				continue
			}
			for _, acc := range c.Accounts {
				if acc.IsActive && cm.banks[acc.BankID] != nil && acc.BankID == bankID {
					total += acc.Balance
				}
			}
		}
		return total, nil
	})

	admin := &Customer{
		CustomerID: cm.generateCustomerID(),
		FirstName:  TrimAndValidateName(firstName),
		LastName:   TrimAndValidateName(lastName),
		IsAdmin:    true,
		IsActive:   true,
		Accounts:   make(map[int]*account.Account),
	}
	cm.customers[admin.CustomerID] = admin
	cm.admin = admin

	return cm
}

func (cm *CustomerManager) GetLedger() *ledger.Ledger {
	return cm.ledger
}

func (cm *CustomerManager) generateCustomerID() int {
	cm.idCounter++
	return cm.idCounter
}

func (cm *CustomerManager) isAuthorizedAdmin() bool {
	return cm.admin != nil && cm.admin.IsAdmin && cm.admin.IsActive
}

func (cm *CustomerManager) isAuthorizedCustomer(customerID int) bool {
	c := cm.customers[customerID]
	return c != nil && c.IsActive && !c.IsAdmin
}

func (cm *CustomerManager) CreateNewBank(fullname string) (*bank.Bank, error) {
	if !cm.isAuthorizedAdmin() {
		return nil, apperror.NewAuthError("admin authorization required")
	}
	id := cm.generateCustomerID()
	b, err := bank.NewBank(id, fullname)
	if err != nil {
		return nil, err
	}
	cm.banks[id] = b
	return b, nil
}

func (cm *CustomerManager) UpdateBankName(bankID int, newName string) error {
	if !cm.isAuthorizedAdmin() {
		return apperror.NewAuthError("admin authorization required")
	}
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
	if !cm.isAuthorizedAdmin() {
		return
	}
	delete(cm.banks, bankID)
}

func (cm *CustomerManager) CreateNewCustomer(firstName, lastName string) (*Customer, error) {
	if !cm.isAuthorizedAdmin() {
		return nil, apperror.NewAuthError("admin authorization required")
	}

	customerID := cm.generateCustomerID()
	c := &Customer{
		CustomerID: customerID,
		FirstName:  TrimAndValidateName(firstName),
		LastName:   TrimAndValidateName(lastName),
		IsAdmin:    false,
		IsActive:   true,
		Accounts:   make(map[int]*account.Account),
	}

	cm.customers[customerID] = c
	return c, nil
}

func (cm *CustomerManager) CreateAccountForCustomer(customerID, bankID int) (*account.Account, error) {
	if !cm.isAuthorizedAdmin() {
		return nil, apperror.NewAuthError("admin authorization required")
	}

	cust, ok := cm.customers[customerID]
	if !ok || !cust.IsActive {
		return nil, apperror.NewNotFoundError("customerID", customerID)

	}

	bank, ok := cm.banks[bankID]
	if !ok {
		return nil, apperror.NewValidationError("bank", fmt.Sprintf("bank ID %d not found", bankID))
	}

	accountID := cm.generateCustomerID() // Or use a separate account ID generator if preferred
	acc, err := account.NewAccount(accountID, customerID, bank.BankID)
	if err != nil {
		if valErr, ok := err.(*apperror.ValidationError); ok {
			return nil, valErr
		}
		return nil, apperror.NewValidationError("account", err.Error())
	}

	cust.Accounts[acc.AccountID] = acc
	return acc, nil
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
	if !cm.isAuthorizedAdmin() {
		return
	}
	c := cm.customers[customerID]
	if c != nil {
		c.IsActive = false
		for _, acc := range c.Accounts {
			acc.IsActive = false
		}
	}
}

func (cm *CustomerManager) DeleteCustomerAccountById(customerID, accountID int) {
	if !cm.isAuthorizedCustomer(customerID) {
		return
	}
	c := cm.customers[customerID]
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
	if !cm.isAuthorizedCustomer(fromCustomerID) || !cm.isAuthorizedCustomer(toCustomerID) {
		return apperror.NewAuthError("only active customers allowed")
	}
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
