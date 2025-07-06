package main

import (
	"banking-app/account"
	"banking-app/bank"
	"banking-app/customer"
	"fmt"
)

type User struct {
	isAdmin  bool
	isActive bool
}

func (u User) IsAdminUser() bool  { return u.isAdmin }
func (u User) IsActiveUser() bool { return u.isActive }

func main() {
	admin := User{isAdmin: true, isActive: true}
	nonAdmin := User{isAdmin: false, isActive: true}

	// --- Since bank functions are package-level, no constructor is needed ---
	// --- customerService uses struct receiver, so we initialize it ---
	customerService := customer.NewCustomerManager()

	// --- Create Bank ---
	b1, _ := bank.CreateBank("SBI")
	fmt.Println("Bank Created:", b1)

	// --- Create Customers ---
	c1, _ := customerService.CreateCustomer("Alice", "Smith")
	c2, _ := customerService.CreateCustomer("Bob", "Jones")
	fmt.Println("Customers Created:", c1, c2)

	// --- Create Accounts ---
	a1, _ := account.CreateAccountAuthorized(admin, c1, b1)
	a2, _ := account.CreateAccountAuthorized(admin, c1, b1)
	a3, _ := account.CreateAccountAuthorized(admin, c2, b1)
	fmt.Println("Accounts Created:", a1.AccountID, a2.AccountID, a3.AccountID)

	// --- Deposit ---
	a1.Deposit(5000)
	a2.Deposit(2000)
	a3.Deposit(3000)

	// --- Withdraw ---
	a1.Withdraw(1000)

	// --- Internal Transfer ---
	a1.TransferToOwnAccount(a2, 1500)

	// --- Admin Transfer ---
	account.AdminTransfer(a2, a3, 500)

	// --- View Balances ---
	fmt.Println("Total Balance c1:", c1.GetTotalBalance())
	fmt.Println("Total Balance c2:", c2.GetTotalBalance())

	// --- Passbook ---
	fmt.Println("Passbook for a1:")
	for _, tx := range a1.GetPassbookPaginated(1, 10) {
		fmt.Println(tx)
	}

	// --- Soft Delete and Reactivate Customer ---
	customerService.DeleteCustomer(c2.CustomerID)
	customerService.ReactivateCustomer(c2.CustomerID)

	// --- Soft Delete Account ---
	account.SoftDeleteAccountByID(admin, a3.AccountID)

	// --- Update Account as Admin ---
	account.UpdateAccountByID(admin, a1.AccountID, "balance", 8000.0)

	// --- Get All Paginated ---
	banks, _ := bank.GetAllBanksPaginated(admin, 1, 10)
	fmt.Println("Banks:", banks)

	customers, _ := customerService.GetAllCustomersPaginated(1, 10)
	fmt.Println("Customers:", customers)

	// --- Non-Admin Tests ---

	// Get account (should fail for non-admin)
	_, err := account.GetAccountByID(a1.AccountID, nonAdmin.IsAdminUser())
	fmt.Println("Non-admin GetAccountByID Error:", err)

	// Transfer between own accounts
	a1.TransferToOwnAccount(a2, 500)
	fmt.Println("After Non-admin Transfer: A1:", a1.Balance, "A2:", a2.Balance)

	// Deposit to own account
	a1.Deposit(1000)
	fmt.Println("After Non-admin Deposit A1:", a1.Balance)

	// Withdraw from own account
	a2.Withdraw(1000)
	fmt.Println("After Non-admin Withdraw A2:", a2.Balance)
}
