package customer

import (
	"banking-app/account"
	"banking-app/apperror"
	"banking-app/bank"
	"banking-app/ledger"
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

func handlePanic(context string) {
	if r := recover(); r != nil {
		fmt.Printf("[RECOVERED PANIC] %s: %v\n", context, r)
	}
}

func NewCustomerManager(firstName, lastName string) *CustomerManager {
	defer handlePanic("NewCustomerManager")

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
	defer handlePanic("GetLedger")
	return cm.ledger
}

func (cm *CustomerManager) generateCustomerID() int {
	defer handlePanic("generateCustomerID")
	cm.idCounter++
	return cm.idCounter
}

func (cm *CustomerManager) isAuthorizedAdmin() bool {
	defer handlePanic("isAuthorizedAdmin")
	return cm.admin != nil && cm.admin.IsAdmin && cm.admin.IsActive
}

func (cm *CustomerManager) isAuthorizedCustomer(customerID int) bool {
	defer handlePanic("isAuthorizedCustomer")
	c := cm.customers[customerID]
	return c != nil && c.IsActive && !c.IsAdmin
}

func (cm *CustomerManager) CreateNewBank(fullname string) (*bank.Bank, error) {
	defer handlePanic("CreateNewBank")

	if !cm.isAuthorizedAdmin() {
		panic("admin authorization required")
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
	defer handlePanic("UpdateBankName")

	if !cm.isAuthorizedAdmin() {
		panic("admin authorization required")
	}
	b, ok := cm.banks[bankID]
	if !ok {
		panic(fmt.Sprintf("bank ID %d not found", bankID))
	}
	return b.UpdateBankName(newName)
}

func (cm *CustomerManager) GetBankById(id int) *bank.Bank {
	defer handlePanic("GetBankById")
	return cm.banks[id]
}

func (cm *CustomerManager) GetAllBanks() []bank.Bank {
	defer handlePanic("GetAllBanks")
	banks := make([]bank.Bank, 0, len(cm.banks))
	for _, b := range cm.banks {
		banks = append(banks, *b)
	}
	return banks
}

func (cm *CustomerManager) DeleteBank(bankID int) {
	defer handlePanic("DeleteBank")

	if !cm.isAuthorizedAdmin() {
		panic("unauthorized: only admin can delete a bank")
	}
	delete(cm.banks, bankID)
}

func (cm *CustomerManager) CreateNewCustomer(firstName, lastName string) (*Customer, error) {
	defer handlePanic("CreateNewCustomer")

	if !cm.isAuthorizedAdmin() {
		panic("admin authorization required")
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
	defer handlePanic("CreateAccountForCustomer")

	if !cm.isAuthorizedAdmin() {
		panic("admin authorization required")
	}

	cust := cm.customers[customerID]
	if cust == nil || !cust.IsActive {
		panic(fmt.Sprintf("customer ID %d not found or inactive", customerID))
	}

	bank := cm.banks[bankID]
	if bank == nil {
		panic(fmt.Sprintf("bank ID %d not found", bankID))
	}

	accountID := cm.generateCustomerID()
	acc, err := account.NewAccount(accountID, customerID, bank.BankID)
	if err != nil {
		panic(err)
	}

	cust.Accounts[acc.AccountID] = acc
	return acc, nil
}

func (cm *CustomerManager) GetAllCustomers() []Customer {
	defer handlePanic("GetAllCustomers")

	customers := make([]Customer, 0, len(cm.customers))
	for _, c := range cm.customers {
		if c.IsActive {
			customers = append(customers, *c)
		}
	}
	return customers
}

func (cm *CustomerManager) GetCustomerById(customerID int) *Customer {
	defer handlePanic("GetCustomerById")
	c := cm.customers[customerID]
	if c != nil && c.IsActive {
		return c
	}
	return nil
}

func (cm *CustomerManager) DeleteCustomer(customerID int) {
	defer handlePanic("DeleteCustomer")

	if !cm.isAuthorizedAdmin() {
		panic("unauthorized: only admin can delete a customer")
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
	defer handlePanic("DeleteCustomerAccountById")

	if !cm.isAuthorizedCustomer(customerID) {
		panic("unauthorized: only active customers can delete their accounts")
	}
	c := cm.customers[customerID]
	if acc, ok := c.Accounts[accountID]; ok {
		acc.IsActive = false
	}
}

func (cm *CustomerManager) DepositMoney(amount float64, accountID int) error {
	defer handlePanic("DepositMoney")

	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		panic(err)
	}
	return acc.DepositMoney(acc.OwnerID, amount)
}

func (cm *CustomerManager) WithDrawMoney(amount float64, accountID int) error {
	defer handlePanic("WithDrawMoney")

	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		panic(err)
	}
	return acc.WithdrawMoney(acc.OwnerID, amount)
}

func (cm *CustomerManager) WithDrawMoneyByAccount_Id(amount float64, accountID int) error {
	defer handlePanic("WithDrawMoneyByAccount_Id")

	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		panic(err)
	}
	return acc.WithdrawMoney(acc.OwnerID, amount)
}

func (cm *CustomerManager) TransferMoney_To_External(amount float64, fromCustomerID, toCustomerID, fromAccountID, toAccountID int) error {
	defer handlePanic("TransferMoney_To_External")

	if !cm.isAuthorizedCustomer(fromCustomerID) || !cm.isAuthorizedCustomer(toCustomerID) {
		panic("only active customers allowed")
	}

	fromAcc, err := account.GetAccountById(fromAccountID)
	if err != nil {
		panic(err)
	}
	toAcc, err := account.GetAccountById(toAccountID)
	if err != nil {
		panic(err)
	}

	if fromAcc.BankID != toAcc.BankID {
		if err := cm.ledger.RecordTransfer(fromAcc.BankID, toAcc.BankID, amount); err != nil {
			panic(err)
		}
	}

	return fromAcc.TransferMoneyToExternal(toAccountID, fromCustomerID, toCustomerID, amount)
}

func (cm *CustomerManager) TransferMoneyInternally(fromAccountID, toAccountID int, amount float64) error {
	defer handlePanic("TransferMoneyInternally")
	return account.TransferMoneyInternally(fromAccountID, toAccountID, amount)
}

func (cm *CustomerManager) GetPassBook_ById(customerID, accountID, pageNo int) []interface{} {
	defer handlePanic("GetPassBook_ById")
	return nil
}

func (cm *CustomerManager) GetTotalBalanceBy_Customer_Id(customerID int) float64 {
	defer handlePanic("GetTotalBalanceBy_Customer_Id")

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
	defer handlePanic("GetTotalBalance")

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
	defer handlePanic("GetAccount_BalanceBy_Id")

	acc, err := cm.GetAccountById(accountID)
	if err != nil {
		return 0
	}
	return acc.Balance
}

func (cm *CustomerManager) GetAccountById(accountID int) (*account.Account, error) {
	defer handlePanic("GetAccountById")

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
	defer handlePanic("DeleteAccountById")

	for _, c := range cm.customers {
		if acc, ok := c.Accounts[accountID]; ok {
			acc.IsActive = false
			return nil
		}
	}
	return apperror.NewNotFoundError("account", accountID)
}

func (cm *CustomerManager) UpdateAccount(accountID int, newBalance float64) error {
	defer handlePanic("UpdateAccount")

	for _, c := range cm.customers {
		if acc, ok := c.Accounts[accountID]; ok && acc.IsActive {
			acc.Balance = newBalance
			return nil
		}
	}
	return apperror.NewNotFoundError("account", accountID)
}

func (cm *CustomerManager) UpdateCustomer(customerID int, firstName, lastName string) error {
	defer handlePanic("UpdateCustomer")

	c := cm.customers[customerID]
	if c == nil || !c.IsActive {
		panic(fmt.Sprintf("customer ID %d not found or inactive", customerID))
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
	defer handlePanic("UpdateCustomerNameById")
	return cm.UpdateCustomer(customerID, firstName, lastName)
}

func TrimAndValidateName(name string) string {
	return name
}
