package user

import (
	"Contact_App/contact"
	"Contact_App/contact_detail"
	"errors"
	"strings"
)

type User struct {
	UserID           int
	FName            string
	LName            string
	IsAdmin          bool
	IsActive         bool
	contacts         []*contact.Contact
	contactIDCounter int
}

var users = make(map[int]*User)
var userIDCounter = 0

// --- Authorization Helpers ---

func (u *User) IsAdminActive() bool {
	return u.IsAdmin && u.IsActive
}

func (u *User) IsStaffActive() bool {
	return !u.IsAdmin && u.IsActive
}

func (u *User) IsActiveUser() bool {
	return u.IsActive
}

func (u *User) IsAdminUser() bool {
	return u.IsAdmin
}

// --- User Factory ---

func (requester *User) NewUser(fname, lname string, isAdmin bool) (*User, error) {
	if !requester.IsAdminActive() {
		return nil, errors.New("only active admins can create new users")
	}
	return ExposeNewUserInternal(fname, lname, isAdmin)
}

func (requester *User) CreateAdminUser(fname, lname string) (*User, error) {
	return requester.NewUser(fname, lname, true)
}

func (requester *User) CreateStaffUser(fname, lname string) (*User, error) {
	return requester.NewUser(fname, lname, false)
}

func ExposeNewUserInternal(fname, lname string, isAdmin bool) (*User, error) {
	fname, lname = strings.TrimSpace(fname), strings.TrimSpace(lname)
	if fname == "" || lname == "" {
		return nil, errors.New("names cannot be empty")
	}

	userIDCounter++
	user := &User{
		UserID:   userIDCounter,
		FName:    fname,
		LName:    lname,
		IsAdmin:  isAdmin,
		IsActive: true,
	}
	users[user.UserID] = user
	return user, nil
}

// --- Contact Access & Management ---

func (u *User) GetContacts() []*contact.Contact {
	return u.contacts
}

func GetAllUsers() []*User {
	all := make([]*User, 0, len(users))
	for _, u := range users {
		all = append(all, u)
	}
	return all
}

func (admin *User) GetUserByID(userID int) (*User, error) {
	if !admin.IsAdminActive() {
		return nil, errors.New("only active admins can access user details")
	}
	user, exists := users[userID]
	if !exists || !user.IsActive {
		return nil, errors.New("user not found or inactive")
	}
	return user, nil
}

func (u *User) GetContactByID(contactID int) (contact_detail.Contact, error) {
	for _, c := range u.contacts {
		if c.ContactID == contactID {
			return c, nil
		}
	}
	return nil, errors.New("contact not found for this user")
}

func (u *User) IncrementContactIDCounter() int {
	u.contactIDCounter++
	return u.contactIDCounter
}

func (u *User) AddContact(c *contact.Contact) {
	u.contacts = append(u.contacts, c)
}

func (u *User) CreateContact(fname, lname string) (*contact.Contact, error) {
	if !u.IsStaffActive() {
		return nil, errors.New("only active staff can create contacts")
	}
	id := u.IncrementContactIDCounter()
	c := &contact.Contact{
		ContactID: id,
		FName:     strings.TrimSpace(fname),
		LName:     strings.TrimSpace(lname),
		IsActive:  true,
	}
	u.AddContact(c)
	return c, nil
}

func (u *User) AddContactWithDetails(fname, lname string, inputs [][2]string) error {
	c, err := u.CreateContact(fname, lname)
	if err != nil {
		return err
	}
	for _, input := range inputs {
		typ := input[0]
		val := input[1]
		_, err := contact_detail.NewContactDetail(c, typ, val)
		if err != nil {
			return err
		}
	}
	return nil
}

// --- User Updates ---

func (u *User) updateFirstName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid firstname")
	}
	u.FName = strings.TrimSpace(v)
	return nil
}

func (u *User) updateLastName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid lastname")
	}
	u.LName = strings.TrimSpace(v)
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

func (u *User) updateIsActive(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return errors.New("invalid isactive")
	}
	u.IsActive = v
	return nil
}

func (u *User) UpdateUser(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "firstname":
		return u.updateFirstName(value)
	case "lastname":
		return u.updateLastName(value)
	case "isadmin":
		return u.updateIsAdmin(value)
	case "isactive":
		return u.updateIsActive(value)
	default:
		return errors.New("unknown field")
	}
}

func (requester *User) UpdateUserByID(userID int, field string, value interface{}) error {
	if !requester.IsActiveUser() {
		return errors.New("only active users can update users")
	}
	if !requester.IsAdminUser() {
		return errors.New("only admins can do CRUD on users")
	}
	user, exists := users[userID]
	if !exists || !user.IsActive {
		return errors.New("user not found or inactive")
	}
	return user.UpdateUser(field, value)
}

// --- Contact/Detail Updates ---

func (u *User) UpdateContactByID(contactID int, field string, value interface{}) error {
	if !u.IsStaffActive() {
		return errors.New("only active staff can update contacts")
	}
	return contact.UpdateContactField(u, u, contactID, field, value)
}

// --- Deletion ---

func (u *User) DeleteContactByID(contactID int) error {
	if !u.IsAdminActive() && !u.IsStaffActive() {
		return errors.New("only active staff or admins can delete contacts")
	}
	return u.DeleteOwnContactByID(contactID)
}

func (u *User) DeleteOwnContactByID(contactID int) error {
	if !u.IsStaffActive() {
		return errors.New("only active staff can delete their own contacts")
	}
	for i, contact := range u.contacts {
		if contact.ContactID == contactID {
			u.contacts = append(u.contacts[:i], u.contacts[i+1:]...)
			return nil
		}
	}
	return errors.New("contact not found")
}

func (u *User) DeleteDetailByID(contactID, detailID int) error {
	if !u.IsStaffActive() {
		return errors.New("only active staff can delete their own contact details")
	}
	for _, c := range u.contacts {
		if c.ContactID == contactID {
			for i, d := range c.Details {
				if d.ContactDetailsID == detailID {
					c.Details = append(c.Details[:i], c.Details[i+1:]...)
					return nil
				}
			}
		}
	}
	return errors.New("detail not found")
}

func (requester *User) DeleteUserByID(id int) error {
	if !requester.IsAdminActive() {
		return errors.New("only admins can delete users")
	}
	user, exists := users[id]
	if !exists {
		return errors.New("user not found")
	}
	user.IsActive = false
	return nil
}
