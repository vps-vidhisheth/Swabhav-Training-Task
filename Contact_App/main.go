package main

import (
	"Contact_App/user"
	"fmt"
)

func displayAllUsers() {
	fmt.Println("\n--- All Users ---")
	for _, u := range user.GetAllUsers() {
		if !u.IsActiveUser() {
			continue
		}
		fmt.Printf("User ID: %d | Name: %s %s | Admin: %t\n", u.UserID, u.FName, u.LName, u.IsAdminUser())
		for _, c := range u.GetContacts() {
			if !c.IsActive {
				continue
			}
			fmt.Printf("  ContactID: %d | Name: %s %s\n", c.ContactID, c.FName, c.LName)
			for _, d := range c.GetDetails() {
				if !d.IsActive {
					continue
				}
				fmt.Printf("    DetailID: %d | Type: %s | Value: %s\n", d.ContactDetailsID, d.Type, d.Value)
			}
		}
	}
}

func main() {
	// 1. Create an initial admin user
	rootAdmin, err := user.ExposeNewUserInternal("Root", "Admin", true)
	if err != nil {
		fmt.Println("Error creating root admin:", err)
		return
	}
	fmt.Println("Created Root Admin:", rootAdmin.FName, rootAdmin.LName)
	displayAllUsers()

	// 2. Admin creates a staff user
	staff1, err := rootAdmin.CreateStaffUser("Alice", "Johnson")
	if err != nil {
		fmt.Println("Error creating staff user:", err)
		return
	}
	fmt.Println("Created Staff User:", staff1.FName, staff1.LName)
	displayAllUsers()

	// 3. Staff adds a contact
	contact1, err := staff1.CreateContact("John", "Doe")
	if err != nil {
		fmt.Println("Error creating contact:", err)
		return
	}
	fmt.Println("Staff added contact:", contact1.FName, contact1.LName)
	displayAllUsers()

	// 4. Staff adds details to the contact
	err = staff1.AddContactWithDetails("Jane", "Smith", [][2]string{
		{"email", "jane@example.com"},
		{"phone", "9876543210"},
	})
	if err != nil {
		fmt.Println("Error adding contact with details:", err)
		return
	}
	fmt.Println("Added contact with details")
	displayAllUsers()

	// 5. Staff updates contact's first name
	err = staff1.UpdateContactByID(contact1.ContactID, "firstname", "Johnny")
	if err != nil {
		fmt.Println("Update Contact Error:", err)
	} else {
		fmt.Println("Updated contact first name to Johnny")
	}
	displayAllUsers()

	// 6. Admin updates staff1's IsActive to false
	err = rootAdmin.UpdateUserByID(staff1.UserID, "isactive", false)
	if err != nil {
		fmt.Println("Admin failed to update user:", err)
	} else {
		fmt.Println("Admin set staff1 IsActive = false")
	}
	displayAllUsers()

	// 7. Attempt to update contact again (should fail due to inactive staff)
	err = staff1.UpdateContactByID(contact1.ContactID, "lastname", "McDoe")
	if err != nil {
		fmt.Println("Expected error (staff inactive):", err)
	}
	displayAllUsers()

	// 8. Admin deletes the staff user
	err = rootAdmin.DeleteUserByID(staff1.UserID)
	if err != nil {
		fmt.Println("Admin failed to delete staff user:", err)
	} else {
		fmt.Println("Admin soft-deleted the staff user")
	}
	displayAllUsers()
}
