package main

import (
	"banking-app/account"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/ledger"
	"fmt"
)

func main() {
	// Initialize managers
	customerManager := customer.NewCustomerManager()
	bankManager := bank.NewBankManager()
	ledger := &ledger.InterbankLedger{} // Assuming you have a ledger implementation

	// Set the ledger in the account package
	account.SetLedger(ledger)

	// Create an admin customer
	admin, err := customerManager.CreateCustomer(nil, "Admin", "User ", true)
	if err != nil {
		fmt.Println("Error creating admin customer:", err)
		return
	}

	// Create a regular customer
	customer1, err := customerManager.CreateCustomer(admin, "John", "Doe", false)
	if err != nil {
		fmt.Println("Error creating customer:", err)
		return
	}

	// Create a bank
	bank1, err := bankManager.CreateBank(admin, "Bank of Go")
	if err != nil {
		fmt.Println("Error creating bank:", err)
		return
	}

	// Create an account for the customer
	account1, err := account.CreateAccount(admin, customer1, bank1)
	if err != nil {
		fmt.Println("Error creating account:", err)
		return
	}

	// Deposit money into the account
	err = account1.Deposit(500)
	if err != nil {
		fmt.Println("Error depositing money:", err)
		return
	}

	// Withdraw money from the account
	err = account1.Withdraw(200)
	if err != nil {
		fmt.Println("Error withdrawing money:", err)
		return
	}

	// Create another account for the same customer
	account2, err := account.CreateAccount(admin, customer1, bank1)
	if err != nil {
		fmt.Println("Error creating second account:", err)
		return
	}

	// Transfer money between accounts
	err = account1.TransferToOwnAccount(account2, 100)
	if err != nil {
		fmt.Println("Error transferring money:", err)
		return
	}

	// Get account details
	accDetails, err := account.GetAccountByID(admin, account1.AccountID)
	if err != nil {
		fmt.Println("Error getting account details:", err)
		return
	}
	fmt.Printf("Account ID: %d, Balance: %.2f\n", accDetails.AccountID, accDetails.Balance)

	// Soft delete the account
	err = account.SoftDeleteAccountByID(admin, account1.AccountID)
	if err != nil {
		fmt.Println("Error deleting account:", err)
		return
	}

	// Check remaining accounts for the customer
	customerAccounts, err := account.GetCustomerAccounts(admin, customer1.CustomerID)
	if err != nil {
		fmt.Println("Error getting customer accounts:", err)
		return
	}
	fmt.Printf("Remaining accounts for customer %s: %d\n", customer1.FirstName, len(customerAccounts))
}
