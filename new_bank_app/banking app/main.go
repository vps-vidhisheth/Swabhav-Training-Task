package main

import (
	"banking-app/customer"
	"fmt"
)

func main() {
	manager := customer.NewCustomerManager("Pragnesh", "Sheth")
	fmt.Println("Admin created successfully.")

	bank1, err := manager.CreateNewBank("State Bank of India")
	if err != nil {
		fmt.Println("Error creating State Bank of India:", err)
	}

	bank2, err := manager.CreateNewBank("Bank of Baroda")
	if err != nil {
		fmt.Println("Error creating Bank of Baroda:", err)
	}

	customer1, err := manager.CreateNewCustomer("Riya", "Parekh", bank1.BankID)
	if err != nil {
		fmt.Println("Error creating customer Riya:", err)
	}

	customer2, err := manager.CreateNewCustomer("Shruti", "Sahu", bank2.BankID)
	if err != nil {
		fmt.Println("Error creating customer Shruti:", err)
	}

	fmt.Println("\n--- All Banks ---")
	for _, b := range manager.GetAllBanks() {
		if b.IsActive {
			fmt.Printf("ID: %d | Name: %s | Abbreviation: %s\n", b.BankID, b.Name, b.Abbreviation)
		}
	}

	fmt.Println("\n--- All Customers ---")
	for _, c := range manager.GetAllCustomers() {
		role := "Customer"
		if c.IsAdmin {
			role = "Admin"
		}
		fmt.Printf("ID: %d | Name: %s %s | Role: %s\n", c.CustomerID, c.FirstName, c.LastName, role)
		for _, acc := range c.Accounts {
			if acc.IsActive {
				fmt.Printf("   AccountID: %d | Balance: %.2f | BankID: %d\n", acc.AccountID, acc.Balance, acc.BankID)
			}
		}
	}

	err = manager.UpdateBankName(bank2.BankID, "Bank of Bharat")
	if err != nil {
		fmt.Println("Error updating bank name:", err)
	}

	if customer1 != nil {
		err = manager.UpdateCustomerNameById(customer1.CustomerID, "Riya", "Sharma")
		if err != nil {
			fmt.Println("Error updating customer name:", err)
		}
	}

	for accID := range customer1.Accounts {
		err = manager.DepositMoney(5000, accID)
		if err != nil {
			fmt.Println("Error depositing:", err)
		}
		break
	}

	for accID := range customer1.Accounts {
		err = manager.WithDrawMoney(1000, accID)
		if err != nil {
			fmt.Println("Error withdrawing:", err)
		}
		break
	}

	accIDs := []int{}
	for accID := range customer1.Accounts {
		accIDs = append(accIDs, accID)
	}
	if len(accIDs) >= 2 {
		err = manager.TransferMoneyInternally(accIDs[0], accIDs[1], 500)
		if err != nil {
			fmt.Println("Error in intra-bank transfer:", err)
		}
	}

	var accID1, accID2 int
	for accID := range customer1.Accounts {
		accID1 = accID
		break
	}
	for accID := range customer2.Accounts {
		accID2 = accID
		break
	}
	err = manager.TransferMoney_To_External(1500, customer1.CustomerID, customer2.CustomerID, accID1, accID2)
	if err != nil {
		fmt.Println("Error in interbank transfer:", err)
	}

	fmt.Println("\n--- Interbank Ledger Balances ---")
	allBalances := manager.GetLedger().AllBalances()
	for fromBankID, debts := range allBalances {
		for toBankID, amt := range debts {
			fmt.Printf("Bank %d owes Bank %d: ₹%.2f\n", fromBankID, toBankID, amt)
		}
	}

	fmt.Println("\n--- Net Bank Positions ---")
	for _, bank := range manager.GetAllBanks() {
		if bank.IsActive {
			actual, receivable, owed, err := manager.GetLedger().GetNetBankPosition(bank.BankID)
			if err != nil {
				fmt.Printf("Error getting net position for Bank ID %d: %v\n", bank.BankID, err)
				continue
			}
			fmt.Printf("Bank ID: %d | Actual: ₹%.2f | Receivable: ₹%.2f | Owed: ₹%.2f\n", bank.BankID, actual, receivable, owed)
		}
	}

	fmt.Println("\n--- Updated Banks ---")
	for _, b := range manager.GetAllBanks() {
		if b.IsActive {
			fmt.Printf("ID: %d | Name: %s | Abbreviation: %s\n", b.BankID, b.Name, b.Abbreviation)
		}
	}

	fmt.Println("\n--- Updated Customers ---")
	for _, c := range manager.GetAllCustomers() {
		fmt.Printf("ID: %d | Name: %s %s\n", c.CustomerID, c.FirstName, c.LastName)
		for _, acc := range c.Accounts {
			if acc.IsActive {
				fmt.Printf("    AccountID: %d | Balance: %.2f | BankID: %d\n", acc.AccountID, acc.Balance, acc.BankID)
			}
		}
	}
}
