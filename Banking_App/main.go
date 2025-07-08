package main

import (
	"banking-app/account"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/ledger"
	"fmt"
)

// Implements the AccountChecker interface from bank package
type accountCheckerImpl struct{}

func (accountCheckerImpl) HasActiveAccounts(bankID int) (bool, error) {
	return account.HasActiveAccounts(bankID)
}

func main() {
	// --- Initialization ---
	customerManager := customer.NewCustomerManager()
	bankManager := bank.NewBankManager()
	ledger := &ledger.InterbankLedger{}
	account.SetLedger(ledger)

	var admin *customer.Customer
	if customerManager.GetActiveCustomerCount() == 0 {
		admin = &customer.Customer{
			FirstName: "Admin",
			LastName:  "User",
			IsActive:  true,
			IsAdmin:   true,
		}
		customerManager.ForceAddCustomer(admin)
		fmt.Println("Admin bootstrapped.")
	}

	// --- Lookup Injectors ---
	account.SetCustomerLookup(func(id int) (*customer.Customer, error) {
		return customerManager.GetCustomerByID(admin, id)
	})
	account.SetBankLookup(func(id int) (*bank.Bank, error) {
		return bankManager.GetBank(admin, id)
	})

	// --- Create Customers ---
	cust1, _ := customerManager.CreateCustomer(admin, "John", "Doe", false)
	cust2, _ := customerManager.CreateCustomer(admin, "Jane", "Smith", false)
	fmt.Printf("Created customers: %s, %s\n", cust1.FirstName, cust2.FirstName)

	// --- Create Banks ---
	bank1, _ := bankManager.CreateBank(admin, "Bank of Go")
	bank2, _ := bankManager.CreateBank(admin, "Bank of Code")
	fmt.Printf("Created banks: %s, %s\n", bank1.Name, bank2.Name)

	// --- Create Accounts ---
	acc1, _ := account.CreateAccount(admin, cust1, bank1)
	acc2, _ := account.CreateAccount(admin, cust1, bank1)
	acc3, _ := account.CreateAccountByIDs(admin, cust2.CustomerID, bank2.BankID)
	fmt.Printf("Created 3 accounts. Account IDs: %d, %d, %d\n", acc1.AccountID, acc2.AccountID, acc3.AccountID)

	// --- Deposit / Withdraw ---
	acc1.Deposit(500)
	acc1.Withdraw(300)
	fmt.Printf("Account %d Balance after deposit/withdraw: %.2f\n", acc1.AccountID, acc1.Balance)

	// --- Own Account Transfer ---
	acc1.TransferToOwnAccount(acc2, 100)
	fmt.Printf("Transferred 100 from acc1 to acc2. Balances: acc1=%.2f, acc2=%.2f\n", acc1.Balance, acc2.Balance)

	// --- Admin Transfer (interbank) ---
	account.AdminTransfer(admin, acc2, acc3, 200)
	fmt.Printf("Admin transferred 200 from acc2 to acc3. Balances: acc2=%.2f, acc3=%.2f\n", acc2.Balance, acc3.Balance)

	// --- Paginated Passbook ---
	passbook := acc2.GetPassbookPaginated(1, 2)
	fmt.Printf("Acc2 Passbook (Page 1):\n")
	for _, tx := range passbook {
		fmt.Printf("  - %s: %.2f\n", tx.Type, tx.Amount)
	}

	// --- Update Fields ---
	account.UpdateAccountByID(admin, acc1.AccountID, "isactive", false)
	customerManager.UpdateCustomerField(admin, cust1.CustomerID, "firstname", "Johnny")
	bankManager.UpdateBankFieldByID(admin, bank1.BankID, "name", "Bank of Golang")
	fmt.Println("Updated acc1 active status, cust1 name, and bank1 name.")

	// --- Get All Accounts Paginated ---
	allAccounts, _ := account.GetAllAccountsPaginated(admin, 1, 10)
	fmt.Printf("All active accounts (paginated):\n")
	for _, a := range allAccounts {
		fmt.Printf("  - Account %d, Customer: %s, Bank: %s, Balance: %.2f\n",
			a.AccountID, a.Customer.FirstName, a.Bank.Name, a.Balance)
	}

	// --- Fetch Customer Accounts ---
	custAccs, _ := account.GetCustomerAccounts(admin, cust1.CustomerID)
	fmt.Printf("Accounts for Customer %s: %d\n", cust1.FirstName, len(custAccs))

	// --- Fetch Bank Accounts ---
	bankAccs, _ := account.GetBankAccounts(admin, bank1.BankID)
	fmt.Printf("Accounts at Bank %s: %d\n", bank1.Name, len(bankAccs))

	// --- Soft Delete Account ---
	account.SoftDeleteAccountByID(admin, acc2.AccountID)
	fmt.Printf("Soft deleted Account %d\n", acc2.AccountID)

	// --- Soft Delete Customer ---
	customerManager.DeleteCustomer(admin, cust2.CustomerID)
	fmt.Printf("Soft deleted Customer %s\n", cust2.FirstName)

	// --- Attempt Soft Delete Bank with Active Account (should fail) ---
	err := bankManager.SoftDeleteBank(admin, accountCheckerImpl{}, bank1.BankID)
	if err != nil {
		fmt.Println("Cannot delete bank1 (has active accounts):", err)
	}

	// --- Delete All Accounts of Bank1 ---
	for _, acc := range bankAccs {
		_ = account.SoftDeleteAccountByID(admin, acc.AccountID)
	}

	// --- Now try soft delete bank again ---
	err = bankManager.SoftDeleteBank(admin, accountCheckerImpl{}, bank1.BankID)
	if err == nil {
		fmt.Printf("Successfully soft deleted bank: %s\n", bank1.Name)
	}
}
