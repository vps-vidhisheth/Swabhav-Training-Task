package main

import (
	"errors"
	"fmt"
	"strings"
)

type User struct {
	UserID           int
	FName            string
	LName            string
	IsAdmin          bool
	IsActive         bool
	contacts         []*Contact
	contactIDCounter int
}

type Contact struct {
	ContactID            int
	FName                string
	LName                string
	IsActive             bool
	details              []*ContactDetail
	contactDetailCounter int
}

type ContactDetail struct {
	ContactDetailsID int
	Type             string
	Value            string
}

type ContactDetailInput struct {
	Type  string
	Value string
}

var (
	userIDCounter = 0
	users         = make(map[int]*User)
)

func NewUser(creator *User, fname, lname string, isAdmin bool) (*User, error) {
	if !isAuthorizedAdmin(creator) {
		return nil, errors.New("only active admins can create new users")
	}
	fname, lname = strings.TrimSpace(fname), strings.TrimSpace(lname)
	if fname == "" || lname == "" {
		return nil, errors.New("names cannot be empty")
	}
	userIDCounter++
	u := &User{
		UserID:           userIDCounter,
		FName:            fname,
		LName:            lname,
		IsAdmin:          isAdmin,
		IsActive:         true,
		contactIDCounter: 0,
	}
	users[u.UserID] = u
	return u, nil
}

func NewContact(u *User, fname, lname string) (*Contact, error) {
	if !isAuthorizedStaff(u) {
		return nil, errors.New("only active staff can add contacts")
	}
	fname, lname = strings.TrimSpace(fname), strings.TrimSpace(lname)
	if fname == "" || lname == "" {
		return nil, errors.New("contact names cannot be empty")
	}
	u.contactIDCounter++
	c := &Contact{
		ContactID:            u.contactIDCounter,
		FName:                fname,
		LName:                lname,
		IsActive:             true,
		contactDetailCounter: 0,
	}
	u.contacts = append(u.contacts, c)
	return c, nil
}

func NewContactDetail(c *Contact, typ, val string) (*ContactDetail, error) {
	typ, val = strings.ToLower(strings.TrimSpace(typ)), strings.TrimSpace(val)
	if (typ != "email" && typ != "phone") || val == "" {
		return nil, errors.New("invalid contact detail")
	}
	c.contactDetailCounter++
	d := &ContactDetail{
		ContactDetailsID: c.contactDetailCounter,
		Type:             typ,
		Value:            val,
	}
	c.details = append(c.details, d)
	return d, nil
}

func (u *User) UpdateUser(requester *User, field string, value interface{}) error {
	if !isAuthorizedAdmin(requester) {
		return errors.New("only active admins can update users")
	}
	switch strings.ToLower(field) {
	case "firstname":
		return u.updateFirstName(value)
	case "lastname":
		return u.updateLastName(value)
	case "isadmin":
		return u.updateIsAdmin(value)
	default:
		return errors.New("unknown field")
	}
}

func (u *User) updateFirstName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid firstname")
	}
	u.FName = v
	return nil
}

func (u *User) updateLastName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid lastname")
	}
	u.LName = v
	return nil
}

func (u *User) updateIsAdmin(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return errors.New("invalid isadmin")
	}
	u.IsAdmin = v
	return nil
}

func (u *User) AddContact(fname, lname string) (*Contact, error) {
	return NewContact(u, fname, lname)
}

func (u *User) AddContactWithDetails(fname, lname string, inputs []ContactDetailInput) error {
	contact, err := NewContact(u, fname, lname)
	if err != nil {
		return err
	}
	for _, input := range inputs {
		_, err := NewContactDetail(contact, input.Type, input.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) GetContactByID(contactID int) (*Contact, error) {
	if !isAuthorizedStaff(u) {
		return nil, errors.New("only active staff can access contacts")
	}
	for _, c := range u.contacts {
		if c.ContactID == contactID {
			return c, nil
		}
	}
	return nil, errors.New("contact not found for this user")
}

func (u *User) UpdateContact(requester *User, contactID int, field string, value interface{}) error {
	if !isAuthorizedStaff(requester) || requester.UserID != u.UserID {
		return errors.New("only active staff can update their own contacts")
	}
	c, err := u.GetContactByID(contactID)
	if err != nil {
		return err
	}
	switch strings.ToLower(field) {
	case "firstname":
		return c.updateFirstName(value)
	case "lastname":
		return c.updateLastName(value)
	default:
		return errors.New("unknown field")
	}
}

func (c *Contact) updateFirstName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid firstname")
	}
	c.FName = v
	return nil
}

func (c *Contact) updateLastName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid lastname")
	}
	c.LName = v
	return nil
}

func (u *User) GetContactDetailByID(contactID, detailID int) (*ContactDetail, error) {
	contact, err := u.GetContactByID(contactID)
	if err != nil {
		return nil, err
	}
	for _, d := range contact.details {
		if d.ContactDetailsID == detailID {
			return d, nil
		}
	}
	return nil, errors.New("contact detail not found for this contact")
}

func (u *User) UpdateDetail(requester *User, contactID, detailID int, field string, value interface{}) error {
	if !isAuthorizedStaff(requester) || requester.UserID != u.UserID {
		return errors.New("only active staff can update their own contact details")
	}
	contact, err := u.GetContactByID(contactID)
	if err != nil {
		return err
	}
	for _, d := range contact.details {
		if d.ContactDetailsID == detailID {
			switch strings.ToLower(field) {
			case "type":
				return d.updateType(value)
			case "value":
				return d.updateValue(value)
			default:
				return errors.New("unknown field")
			}
		}
	}
	return errors.New("detail not found")
}

func (d *ContactDetail) updateType(value interface{}) error {
	v, ok := value.(string)
	v = strings.ToLower(strings.TrimSpace(v))
	if !ok || (v != "email" && v != "phone") {
		return errors.New("invalid type")
	}
	d.Type = v
	return nil
}

func (d *ContactDetail) updateValue(value interface{}) error {
	v, ok := value.(string)
	v = strings.TrimSpace(v)
	if !ok || v == "" {
		return errors.New("invalid value")
	}
	d.Value = v
	return nil
}

func isAuthorizedAdmin(u *User) bool {
	return u != nil && u.IsAdmin && u.IsActive
}

func isAuthorizedStaff(u *User) bool {
	return u != nil && !u.IsAdmin && u.IsActive
}

func DeleteUserByID(requester *User, id int) error {
	if !isAuthorizedAdmin(requester) {
		return errors.New("only admins can delete users")
	}
	if _, exists := users[id]; !exists {
		return errors.New("user not found")
	}
	delete(users, id)
	return nil
}

func DeleteContactByID(requester *User, userID, contactID int) error {
	user, exists := users[userID]
	if !exists {
		return errors.New("user not found")
	}
	if isAuthorizedAdmin(requester) || (isAuthorizedStaff(requester) && requester.UserID == userID) {
		return user.DeleteOwnContactByID(contactID)
	}
	return errors.New("unauthorized to delete this contact")
}

func (u *User) DeleteOwnContactByID(contactID int) error {
	for i, c := range u.contacts {
		if c.ContactID == contactID {
			u.contacts = append(u.contacts[:i], u.contacts[i+1:]...)
			return nil
		}
	}
	return errors.New("contact not found")
}

func DeleteDetailByID(requester *User, userID, contactID, detailID int) error {
	user, exists := users[userID]
	if !exists {
		return errors.New("user not found")
	}
	if !isAuthorizedAdmin(requester) && !(isAuthorizedStaff(requester) && requester.UserID == userID) {
		return errors.New("unauthorized to delete this contact detail")
	}
	for _, c := range user.contacts {
		if c.ContactID == contactID {
			for i, d := range c.details {
				if d.ContactDetailsID == detailID {
					c.details = append(c.details[:i], c.details[i+1:]...)
					return nil
				}
			}
		}
	}
	return errors.New("detail not found")
}

func displayAllUsers() {
	fmt.Println("\n--- All Users ---")
	for _, u := range users {
		fmt.Printf("UserID: %d | Name: %s %s | Admin: %t\n", u.UserID, u.FName, u.LName, u.IsAdmin)
		for _, c := range u.contacts {
			fmt.Printf("  ContactID: %d | Name: %s %s\n", c.ContactID, c.FName, c.LName)
			for _, d := range c.details {
				fmt.Printf("    DetailID: %d | Type: %s | Value: %s\n", d.ContactDetailsID, d.Type, d.Value)
			}
		}
	}
}

func GetUserByID(id int) (*User, error) {
	u, exists := users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func UpdateUserByID(requester *User, userID int, field string, value interface{}) error {
	if !isAuthorizedAdmin(requester) {
		return errors.New("only active admins can update users")
	}
	u, err := GetUserByID(userID)
	if err != nil {
		return err
	}
	return u.UpdateUser(requester, field, value)
}

func UpdateContactByID(requester *User, userID, contactID int, field string, value interface{}) error {
	u, err := GetUserByID(userID)
	if err != nil {
		return err
	}
	if isAuthorizedAdmin(requester) {
		c, err := u.GetContactByID(contactID)
		if err != nil {
			return err
		}
		switch strings.ToLower(field) {
		case "firstname":
			return c.updateFirstName(value)
		case "lastname":
			return c.updateLastName(value)
		default:
			return errors.New("unknown field")
		}
	}
	if isAuthorizedStaff(requester) && requester.UserID == userID {
		return u.UpdateContact(requester, contactID, field, value)
	}
	return errors.New("unauthorized to update this contact")
}

func UpdateContactDetailByID(requester *User, userID, contactID, detailID int, field string, value interface{}) error {
	u, err := GetUserByID(userID)
	if err != nil {
		return err
	}
	contact, err := u.GetContactByID(contactID)
	if err != nil {
		return err
	}
	if isAuthorizedAdmin(requester) {
		for _, d := range contact.details {
			if d.ContactDetailsID == detailID {
				switch strings.ToLower(field) {
				case "type":
					return d.updateType(value)
				case "value":
					return d.updateValue(value)
				default:
					return errors.New("unknown field")
				}
			}
		}
		return errors.New("contact detail not found")
	}
	if isAuthorizedStaff(requester) && requester.UserID == userID {
		return requester.UpdateDetail(requester, contactID, detailID, field, value)
	}
	return errors.New("unauthorized to update this contact detail")
}

func main() {
	fmt.Println("--- Contact Management System ---")

	admin := &User{UserID: 0, FName: "Super", LName: "Admin", IsAdmin: true, IsActive: true}
	users[admin.UserID] = admin

	user1, _ := NewUser(admin, "Alice", "Smith", false)
	fmt.Println("\nAdded User: Alice Smith")
	displayAllUsers()

	user2, _ := NewUser(admin, "Bob", "Jones", false)
	fmt.Println("\nAdded User: Bob Jones")
	displayAllUsers()

	err := UpdateUserByID(admin, user1.UserID, "firstname", "Alicia")
	if err != nil {
		fmt.Println("UpdateUserByID error:", err)
	} else {
		fmt.Println("\nUpdated User1's firstname to Alicia")
	}
	displayAllUsers()

	c1, _ := user1.AddContact("John", "Doe")
	fmt.Println("\nUser1 added Contact: John Doe")
	displayAllUsers()

	err = user1.AddContactWithDetails("Jane", "Doe", []ContactDetailInput{
		{"email", "jane@example.com"},
		{"phone", "1234567890"},
	})
	if err != nil {
		fmt.Println("AddContactWithDetails error:", err)
	} else {
		fmt.Println("\nUser1 added Contact with details: Jane Doe")
	}
	displayAllUsers()

	err = UpdateContactByID(user1, user1.UserID, c1.ContactID, "lastname", "Dover")
	if err != nil {
		fmt.Println("UpdateContactByID error:", err)
	} else {
		fmt.Println("\nUpdated Contact John Doe's lastname to Dover")
	}
	displayAllUsers()

	d1, _ := user1.GetContactDetailByID(2, 2)
	err = UpdateContactDetailByID(user1, user1.UserID, 2, d1.ContactDetailsID, "value", "jane.doe@example.com")
	if err != nil {
		fmt.Println("UpdateContactDetailByID error:", err)
	} else {
		fmt.Println("\nUpdated Jane Doe's contact detail email value")
	}
	displayAllUsers()

	err = DeleteDetailByID(user1, user1.UserID, 2, d1.ContactDetailsID)
	if err != nil {
		fmt.Println("DeleteDetailByID error:", err)
	} else {
		fmt.Println("\nDeleted Jane Doe's contact detail")
	}
	displayAllUsers()

	err = DeleteContactByID(user1, user1.UserID, 2)
	if err != nil {
		fmt.Println("DeleteContactByID error:", err)
	} else {
		fmt.Println("\nDeleted Contact Jane Doe")
	}
	displayAllUsers()

	err = DeleteUserByID(admin, user2.UserID)
	if err != nil {
		fmt.Println("DeleteUserByID error:", err)
	} else {
		fmt.Println("\nDeleted User Bob Jones")
	}
	displayAllUsers()

	if foundUser, err := GetUserByID(user1.UserID); err == nil {
		fmt.Printf("\nFound User: ID=%d, Name=%s %s\n", foundUser.UserID, foundUser.FName, foundUser.LName)
	} else {
		fmt.Println("User not found")
	}
	displayAllUsers()
}
