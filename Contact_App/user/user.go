package user

import (
	"Contact_App/apperror"
	"Contact_App/contact"
	"Contact_App/contact_detail"
	"fmt"
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

// --- Recovery Helper ---

func safeExecUser(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in [%s]: %v\n", label, r)
		}
	}()
	fn()
}

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
	var result *User
	safeExecUser("NewUser", func() {
		if !requester.IsAdminActive() {
			panic(apperror.NewAuthError("create a new user"))
		}
		var err error
		result, err = ExposeNewUserInternal(fname, lname, isAdmin)
		if err != nil {
			panic(err)
		}
	})
	return result, nil
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
		return nil, apperror.NewValidationError("name", "first or last name cannot be empty")
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
	var result *User
	safeExecUser("GetUserByID", func() {
		if !admin.IsAdminActive() {
			panic(apperror.NewAuthError("access user details"))
		}
		user, exists := users[userID]
		if !exists || !user.IsActive {
			panic(apperror.NewNotFoundError("user", userID))
		}
		result = user
	})
	return result, nil
}

func (u *User) GetContactByID(contactID int) (contact_detail.Contact, error) {
	for _, c := range u.contacts {
		if c.ContactID == contactID {
			return c, nil
		}
	}
	return nil, apperror.NewNotFoundError("contact", contactID)
}

func (u *User) IncrementContactIDCounter() int {
	u.contactIDCounter++
	return u.contactIDCounter
}

func (u *User) AddContact(c *contact.Contact) {
	u.contacts = append(u.contacts, c)
}

func (u *User) CreateContact(fname, lname string) (c *contact.Contact, err error) {
	safeExecUser("CreateContact", func() {
		if !u.IsStaffActive() {
			panic(apperror.NewAuthError("create contacts"))
		}
		id := u.IncrementContactIDCounter()
		c = &contact.Contact{
			ContactID: id,
			FName:     strings.TrimSpace(fname),
			LName:     strings.TrimSpace(lname),
			IsActive:  true,
		}
		u.AddContact(c)
	})
	return
}

func (u *User) AddContactWithDetails(fname, lname string, inputs [][2]string) (err error) {
	safeExecUser("AddContactWithDetails", func() {
		c, e := u.CreateContact(fname, lname)
		if e != nil {
			panic(e)
		}
		for _, input := range inputs {
			typ := input[0]
			val := input[1]
			_, err := contact_detail.NewContactDetail(c, typ, val)
			if err != nil {
				panic(apperror.NewContactDetailError("creating detail", err.Error()))
			}
		}
	})
	return
}

// --- User Updates ---

func (u *User) updateFirstName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("firstname", "must be a non-empty string")
	}
	u.FName = strings.TrimSpace(v)
	return nil
}

func (u *User) updateLastName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("lastname", "must be a non-empty string")
	}
	u.LName = strings.TrimSpace(v)
	return nil
}

func (u *User) updateIsAdmin(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isadmin", "must be a boolean")
	}
	u.IsAdmin = v
	return nil
}

func (u *User) updateIsActive(value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
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
		return apperror.NewValidationError("field", "unknown update field")
	}
}

func (requester *User) UpdateUserByID(userID int, field string, value interface{}) (err error) {
	safeExecUser("UpdateUserByID", func() {
		if !requester.IsActiveUser() {
			panic(apperror.NewAuthError("update user"))
		}
		if !requester.IsAdminUser() {
			panic(apperror.NewAuthError("perform user CRUD"))
		}
		user, exists := users[userID]
		if !exists || !user.IsActive {
			panic(apperror.NewNotFoundError("user", userID))
		}
		err = user.UpdateUser(field, value)
	})
	return
}

// --- Contact/Detail Updates ---

func (u *User) UpdateContactByID(contactID int, field string, value interface{}) (err error) {
	safeExecUser("UpdateContactByID", func() {
		if !u.IsStaffActive() {
			panic(apperror.NewAuthError("update contact"))
		}
		err = contact.UpdateContactField(u, u, contactID, field, value)
	})
	return
}

// --- Deletion ---

func (u *User) DeleteContactByID(contactID int) (err error) {
	safeExecUser("DeleteContactByID", func() {
		if !u.IsAdminActive() && !u.IsStaffActive() {
			panic(apperror.NewAuthError("delete contact"))
		}
		err = u.DeleteOwnContactByID(contactID)
	})
	return
}

func (u *User) DeleteOwnContactByID(contactID int) (err error) {
	safeExecUser("DeleteOwnContactByID", func() {
		if !u.IsStaffActive() {
			panic(apperror.NewAuthError("delete their own contacts"))
		}
		for i, contact := range u.contacts {
			if contact.ContactID == contactID {
				u.contacts = append(u.contacts[:i], u.contacts[i+1:]...)
				return
			}
		}
		err = apperror.NewNotFoundError("contact", contactID)
	})
	return
}

func (u *User) DeleteDetailByID(contactID, detailID int) (err error) {
	safeExecUser("DeleteDetailByID", func() {
		if !u.IsStaffActive() {
			panic(apperror.NewAuthError("delete contact details"))
		}
		for _, c := range u.contacts {
			if c.ContactID == contactID {
				for i, d := range c.Details {
					if d.ContactDetailsID == detailID {
						c.Details = append(c.Details[:i], c.Details[i+1:]...)
						return
					}
				}
			}
		}
		err = apperror.NewNotFoundError("detail", detailID)
	})
	return
}

func (requester *User) DeleteUserByID(id int) (err error) {
	safeExecUser("DeleteUserByID", func() {
		if !requester.IsAdminActive() {
			panic(apperror.NewAuthError("delete users"))
		}
		user, exists := users[id]
		if !exists {
			err = apperror.NewNotFoundError("user", id)
			return
		}
		user.IsActive = false
	})
	return
}
