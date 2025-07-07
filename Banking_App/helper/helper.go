package main

import (
	"banking-app/account"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/ledger"
	"fmt"
)

func main() {
	// Initialize the ledger
	interbankLedger := ledger.NewInterbankLedger(nil)
	account.SetLedger(interbankLedger)

	// Create a bank
	bank1, err := bank.CreateBank("Bank of Go")
	if err != nil {
		fmt.Println("Error creating bank:", err)
		return
	}
	fmt.Printf("Created bank: %s (ID: %d)\n", bank1.Name, bank1.BankID)

	// Create a customer manager
	custManager := customer.NewCustomerManager()

	// Create customers
	cust1, err := custManager.CreateCustomer("Alice", "Smith")
	if err != nil {
		fmt.Println("Error creating customer:", err)
		return
	}
	fmt.Printf("Created customer: %s %s (ID: %d)\n", cust1.FirstName, cust1.LastName, cust1.CustomerID)

	cust2, err := custManager.CreateCustomer("Bob", "Johnson")
	if err != nil {
		fmt.Println("Error creating customer:", err)
		return
	}
	fmt.Printf("Created customer: %s %s (ID: %d)\n", cust2.FirstName, cust2.LastName, cust2.CustomerID)

	// Create accounts for customers
	acc1, err := account.CreateAccount(cust1, bank1)
	if err != nil {
		fmt.Println("Error creating account for customer 1:", err)
		return
	}
	fmt.Printf("Created account for %s: Account ID %d with balance %.2f\n", cust1.FirstName, acc1.AccountID, acc1.Balance)

	acc2, err := account.CreateAccount(cust2, bank1)
	if err != nil {
		fmt.Println("Error creating account for customer 2:", err)
		return
	}
	fmt.Printf("Created account for %s: Account ID %d with balance %.2f\n", cust2.FirstName, acc2.AccountID, acc2.Balance)

	// Deposit money into account 1
	err = acc1.Deposit(500)
	if err != nil {
		fmt.Println("Error depositing into account 1:", err)
		return
	}
	fmt.Printf("Deposited 500 into account %d. New balance: %.2f\n", acc1.AccountID, acc1.Balance)

	// Withdraw money from account 2
	err = acc2.Withdraw(200)
	if err != nil {
		fmt.Println("Error withdrawing from account 2:", err)
		return
	}
	fmt.Printf("Withdrew 200 from account %d. New balance: %.2f\n", acc2.AccountID, acc2.Balance)

	// Transfer money from account 1 to account 2
	err = acc1.TransferToOwnAccount(acc2, 300)
	if err != nil {
		fmt.Println("Error transferring from account 1 to account 2:", err)
		return
	}
	fmt.Printf("Transferred 300 from account %d to account %d. New balances: Account 1: %.2f, Account 2: %.2f\n", acc1.AccountID, acc2.AccountID, acc1.Balance, acc2.Balance)

	// Admin transfer (simulated)
	admin := &customer.Customer{CustomerID: 999, IsActive: true} // Simulating an admin user
	err = account.AdminTransfer(admin, acc1, acc2, 100)
	if err != nil {
		fmt.Println("Error in admin transfer:", err)
		return
	}
	fmt.Printf("Admin transferred 100 from account %d to account %d. New balances: Account 1: %.2f, Account 2: %.2f\n", acc1.AccountID, acc2.AccountID, acc1.Balance, acc2.Balance)

	// Get account details
	accDetails, err := account.GetAccountByID(acc1.AccountID, true)
	if err != nil {
		fmt.Println("Error getting account details:", err)
		return
	}
	fmt.Printf("Account details for account ID %d: Balance: %.2f, Active: %t\n", accDetails.AccountID, accDetails.Balance, accDetails.IsActive)

	// Soft delete account 1
	err = acc1.SoftDelete()
	if err != nil {
		fmt.Println("Error soft deleting account 1:", err)
		return
	}
	fmt.Printf("Soft deleted account %d. Active status: %t\n", acc1.AccountID, acc1.IsActive)

	// Check total balance in the bank
	totalBalance, err := account.GetBankTotalBalance(bank1.BankID)
	if err != nil {
		fmt.Println("Error getting total balance for bank:", err)
		return
	}
	fmt.Printf("Total balance in bank %s: %.2f\n", bank1.Name, totalBalance)
}
