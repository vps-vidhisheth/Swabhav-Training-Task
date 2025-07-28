package user

import (
	"Contact_App/apperror"
	"Contact_App/auth"
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
	Email            string
	Password         string
	contacts         []*contact.Contact
	contactIDCounter int
}

var users = make(map[int]*User)
var userIDCounter = 0

func safeExecUser(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in [%s]: %v\n", label, r)
		}
	}()
	fn()
}

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

func (requester *User) NewUserWithCredentials(fname, lname string, isAdmin bool, email, password string) (*User, error) {
	var result *User
	var err error
	safeExecUser("NewUserWithCredentials", func() {
		if !requester.IsAdminActive() {
			err = apperror.NewAuthError("create user")
			return
		}
		if fname == "" || lname == "" || email == "" || password == "" {
			err = apperror.NewValidationError("user", "all fields are required")
			return
		}
		for _, u := range users {
			if u.Email == strings.ToLower(email) {
				err = apperror.NewValidationError("email", "already registered")
				return
			}
		}

		userIDCounter++
		result = &User{
			UserID:   userIDCounter,
			FName:    strings.TrimSpace(fname),
			LName:    strings.TrimSpace(lname),
			IsAdmin:  isAdmin,
			IsActive: true,
			Email:    strings.ToLower(email),
			Password: password,
		}
		users[result.UserID] = result
	})
	return result, err
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
		Email:    fmt.Sprintf("%s.%s@example.com", strings.ToLower(fname), strings.ToLower(lname)),
		Password: "admin123",
	}
	users[user.UserID] = user
	return user, nil
}

func CreateInitialAdminUser() *User {
	admin, _ := ExposeNewUserInternal("Admin", "User", true)
	return admin
}

func (requester *User) CreateUser(fname, lname string, isAdmin bool, email, password string) (*User, error) {
	return requester.NewUserWithCredentials(fname, lname, isAdmin, email, password)
}

func Authenticate(email, password string) (*User, error) {
	for _, user := range users {
		if user.Email == strings.ToLower(email) && user.Password == password && user.IsActive {
			return user, nil
		}
	}
	return nil, fmt.Errorf("invalid email or password")
}

func GenerateJWT(u *User) (string, error) {
	return auth.GenerateToken(u.UserID, u.IsAdmin, u.IsActive)
}

func (u *User) GetContacts() []*contact.Contact {
	return u.contacts
}

func (u *User) GetAllUsers() []*User {
	if !u.IsAdminActive() {
		return nil
	}

	all := make([]*User, 0, len(users))
	for _, user := range users {
		if user.IsActive {
			all = append(all, user)
		}
	}
	return all
}

func GetUserByID(userID int) (*User, error) {
	user, exists := users[userID]
	if !exists || !user.IsActive {
		return nil, apperror.NewNotFoundError("user", userID)
	}
	return user, nil
}

func GetAllUsers() []*User {
	all := make([]*User, 0, len(users))
	for _, user := range users {
		if user.IsActive {
			all = append(all, user)
		}
	}
	return all
}

func (u *User) GetContactByID(contactID int) (*contact.Contact, error) {
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
		fmt.Printf("CreateContact called by UserID=%d IsActive=%v\n", u.UserID, u.IsActive)
		if !u.IsActive {
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
			err = e
			return
		}
		for _, input := range inputs {
			typ := input[0]
			val := input[1]
			_, err = contact_detail.NewContactDetail(c, typ, val)
			if err != nil {
				err = apperror.NewContactDetailError("creating detail", err.Error())
				return
			}
		}
	})
	return
}

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

func (u *User) UpdateField(field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "firstname", "fname":
		return u.updateFirstName(value)
	case "lastname", "lname":
		return u.updateLastName(value)
	case "isadmin":
		return u.updateIsAdmin(value)
	case "isactive":
		return u.updateIsActive(value)
	default:
		return apperror.NewValidationError("field", "unknown update field")
	}
}

func UpdateUser(u *User) error {
	if u == nil {
		return apperror.NewValidationError("user", "cannot be nil")
	}
	existing, ok := users[u.UserID]
	if !ok || !existing.IsActive {
		return apperror.NewNotFoundError("user", u.UserID)
	}
	users[u.UserID] = u
	return nil
}

func (requester *User) UpdateUserByID(userID int, field string, value interface{}) (err error) {
	safeExecUser("UpdateUserByID", func() {
		if !requester.IsActiveUser() || !requester.IsAdminUser() {
			err = apperror.NewAuthError("update user")
			return
		}
		user, exists := users[userID]
		if !exists || !user.IsActive {
			err = apperror.NewNotFoundError("user", userID)
			return
		}
		err = user.UpdateField(field, value)
	})
	return
}

func (u *User) UpdateContactByID(contactID int, field string, value interface{}) (err error) {
	safeExecUser("UpdateContactByID", func() {
		if !u.IsStaffActive() {
			err = apperror.NewAuthError("update contact")
			return
		}
		err = contact.UpdateContactField(u, u, contactID, field, value)
	})
	return
}

func (u *User) DeleteContactByID(contactID int) (err error) {
	safeExecUser("DeleteContactByID", func() {
		if !u.IsAdminActive() && !u.IsStaffActive() {
			err = apperror.NewAuthError("delete contact")
			return
		}
		err = u.DeleteOwnContactByID(contactID)
	})
	return
}

func (u *User) DeleteOwnContactByID(contactID int) (err error) {
	safeExecUser("DeleteOwnContactByID", func() {
		if !u.IsStaffActive() {
			err = apperror.NewAuthError("delete their own contacts")
			return
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
			err = apperror.NewAuthError("delete contact details")
			return
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
			err = apperror.NewAuthError("delete users")
			return
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

func (u *User) AddDetailToContact(contactID int, detailType, value string) (*contact_detail.ContactDetail, error) {
	var detail *contact_detail.ContactDetail
	var err error

	safeExecUser("AddDetailToContact", func() {
		if !u.IsStaffActive() && !u.IsAdminActive() {
			err = apperror.NewAuthError("add contact details")
			return
		}
		contactObj, e := u.GetContactByID(contactID)
		if e != nil {
			err = e
			return
		}

		detail, err = contact_detail.NewContactDetail(contactObj, detailType, value)
		if err != nil {
			err = apperror.NewContactDetailError("creating detail", err.Error())
			return
		}
	})

	return detail, err
}
