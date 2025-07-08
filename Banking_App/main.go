package main

import (
	"banking-app/account"
	"banking-app/bank"
	"banking-app/customer"
	"banking-app/ledger"
	"fmt"
)

type AccountBalanceReaderAdapter struct{}

func (a AccountBalanceReaderAdapter) GetBankTotalBalance(bankID int) (float64, error) {
	return account.GetBankTotalBalance(bankID)
}

func main() {
	fmt.Println("Banking System Demonstration")

	customerManager := customer.NewManager()
	bankManager := bank.NewBankManager()

	accountBalanceReader := AccountBalanceReaderAdapter{}
	interbankLedger := ledger.NewInterbankLedger(accountBalanceReader)

	admin := &customer.Customer{
		CustomerID: 1,
		FirstName:  "Admin",
		LastName:   "User",
		IsActive:   true,
		IsAdmin:    true,
	}
	customerManager.ForceAdd(admin)
	fmt.Printf("Admin user created with ID: %d\n", admin.CustomerID)

	customer1 := &customer.Customer{CustomerID: 101, FirstName: "John", LastName: "Doe", IsActive: true}
	customerManager.ForceAdd(customer1)
	fmt.Printf("Customer 1 created with ID: %d\n", customer1.CustomerID)

	customer2 := &customer.Customer{CustomerID: 102, FirstName: "Jane", LastName: "Smith", IsActive: true}
	customerManager.ForceAdd(customer2)
	fmt.Printf("Customer 2 created with ID: %d\n", customer2.CustomerID)

	bank1, err := bankManager.CreateBank(admin, "National Bank")
	if err != nil {
		panic("Failed to create National Bank: " + err.Error())
	}
	bank2, err := bankManager.CreateBank(admin, "City Bank")
	if err != nil {
		panic("Failed to create City Bank: " + err.Error())
	}

	account1, err := account.CreateAccountByIDs(admin, customer1.CustomerID, bank1.BankID, bankManager, customerManager)
	if err != nil {
		panic("Failed to create account1: " + err.Error())
	}
	account2, err := account.CreateAccountByIDs(admin, customer1.CustomerID, bank2.BankID, bankManager, customerManager)
	if err != nil {
		panic("Failed to create account2: " + err.Error())
	}
	account3, err := account.CreateAccountByIDs(admin, customer2.CustomerID, bank1.BankID, bankManager, customerManager)
	if err != nil {
		panic("Failed to create account3: " + err.Error())
	}

	account1.Deposit(1000)
	account2.Deposit(500)
	account3.Deposit(2000)

	fmt.Println("\nInitial Account Status")
	fmt.Printf("Account %d (Bank %d, Customer %s): Balance Rs %.2f\n",
		account1.AccountID, account1.Bank.BankID, account1.Customer.FirstName, account1.Balance)
	fmt.Printf("Account %d (Bank %d, Customer %s): Balance Rs %.2f\n",
		account2.AccountID, account2.Bank.BankID, account2.Customer.FirstName, account2.Balance)
	fmt.Printf("Account %d (Bank %d, Customer %s): Balance Rs %.2f\n",
		account3.AccountID, account3.Bank.BankID, account3.Customer.FirstName, account3.Balance)

	fmt.Println("\nPerforming Transactions")
	account1.Deposit(500)
	fmt.Printf("Deposited Rs %.2f to account %d\n", 500.0, account1.AccountID)

	account1.Withdraw(200)
	fmt.Printf("Withdrew Rs %.2f from account %d\n", 200.0, account1.AccountID)

	account1.TransferToOwnAccount(account2, 100)
	fmt.Printf("Transferred Rs %.2f between accounts %d and %d (same customer)\n", 100.0, account1.AccountID, account2.AccountID)

	account.AdminTransfer(admin, account1, account3, 50, interbankLedger)
	fmt.Printf("Admin transferred Rs %.2f from account %d to account %d (within same bank)\n", 50.0, account1.AccountID, account3.AccountID)

	fmt.Println("\nUpdated Account Status After Initial Transactions")
	fmt.Printf("Account %d (Bank %d, Customer %s): Balance Rs %.2f\n",
		account1.AccountID, account1.Bank.BankID, account1.Customer.FirstName, account1.Balance)
	fmt.Printf("Account %d (Bank %d, Customer %s): Balance Rs %.2f\n",
		account2.AccountID, account2.Bank.BankID, account2.Customer.FirstName, account2.Balance)
	fmt.Printf("Account %d (Bank %d, Customer %s): Balance Rs %.2f\n",
		account3.AccountID, account3.Bank.BankID, account3.Customer.FirstName, account3.Balance)

	fmt.Println("\nInterbank Transfers")

	account.AdminTransfer(admin, account1, account3, 100, interbankLedger)
	fmt.Printf("Admin transferred Rs %.2f within bank %d (Account %d -> Account %d)\n", 100.0, bank1.BankID, account1.AccountID, account3.AccountID)

	account.AdminTransfer(admin, account1, account2, 150, interbankLedger)
	fmt.Printf("Admin transferred Rs %.2f from bank %d (Account %d) to bank %d (Account %d)\n", 150.0, bank1.BankID, account1.AccountID, bank2.BankID, account2.AccountID)

	fmt.Println("\nTransaction History")

	fmt.Printf("\nAccount %d Transactions:\n", account1.AccountID)
	for _, t := range account1.Transactions {
		fmt.Printf("  %s: Rs %.2f (ID: %d)\n", t.Type, t.Amount, t.TransactionID)
	}

	fmt.Printf("\nAccount %d Transactions:\n", account2.AccountID)
	for _, t := range account2.Transactions {
		fmt.Printf("  %s: Rs %.2f (ID: %d)\n", t.Type, t.Amount, t.TransactionID)
	}

	fmt.Printf("\nAccount %d Transactions:\n", account3.AccountID)
	for _, t := range account3.Transactions {
		fmt.Printf("  %s: Rs %.2f (ID: %d)\n", t.Type, t.Amount, t.TransactionID)
	}

	fmt.Println("\nLedger Settlement Report")
	for _, b := range []*bank.Bank{bank1, bank2} {
		actual, receivable, owed, err := interbankLedger.GetNetBankPosition(b.BankID)
		if err != nil {
			fmt.Printf("Error retrieving net position for bank %d (%s): %v\n", b.BankID, b.Name, err)
			continue
		}
		settlementScore := actual + receivable - owed

		fmt.Printf("\nBank %d - %s\n", b.BankID, b.Name)
		fmt.Println("--------------------------------------------------")
		fmt.Printf("Actual Balance   : Rs %.2f\n", actual)
		fmt.Printf("Receivables      : Rs %.2f\n", receivable)
		fmt.Printf("Payables (Owed)  : Rs %.2f\n", owed)
		fmt.Printf("Settlement Score : Rs %.2f\n", settlementScore)
	}

	fmt.Println("\nBanking System Demonstration Complete")
}
