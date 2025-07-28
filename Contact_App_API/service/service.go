package service

import (
	"Contact_App/apperror"
	"Contact_App/contact"
	"Contact_App/contact_detail"
	"Contact_App/user"
)

func AuthenticateUser(email, password string) (*user.User, error) {
	u, err := user.Authenticate(email, password)
	if err != nil {
		return nil, apperror.NewAuthError("authentication failed")
	}
	return u, nil
}

func GetUserByID(id int) (*user.User, error) {
	return user.GetUserByID(id)
}

func CreateUser(requesterID int, fname, lname string, isAdmin bool, email, password string) (*user.User, error) {
	u, err := user.GetUserByID(requesterID)
	if err != nil {
		return nil, err
	}
	return u.CreateUser(fname, lname, isAdmin, email, password)
}

func GetAllUsers(requesterID int) ([]*user.User, error) {
	u, err := user.GetUserByID(requesterID)
	if err != nil {
		return nil, err
	}
	if !u.IsAdminActive() {
		return nil, apperror.NewAuthError("get all users")
	}
	return user.GetAllUsers(), nil
}

func UpdateUserByID(requesterID, userID int, field string, value interface{}) error {
	u, err := user.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	if field != "fname" && field != "lname" {
		return apperror.NewValidationError(field, "Only 'fname' or 'lname' can be updated")
	}
	return u.UpdateUserByID(userID, field, value)
}

func DeleteUserByID(requesterID, userID int) error {
	u, err := user.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	return u.DeleteUserByID(userID)
}

func CreateContact(userID int, fname, lname string) (*contact.Contact, error) {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return u.CreateContact(fname, lname)
}

func AddContactWithDetails(userID int, fname, lname string, inputs [][2]string) error {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return err
	}
	return u.AddContactWithDetails(fname, lname, inputs)
}

func UpdateContactByID(userID, contactID int, field string, value interface{}) error {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return err
	}
	return u.UpdateContactByID(contactID, field, value)
}

func DeleteContactByID(userID, contactID int) error {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return err
	}
	return u.DeleteContactByID(contactID)
}

func DeleteDetailByID(userID, contactID, detailID int) error {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return err
	}
	return u.DeleteDetailByID(contactID, detailID)
}

func AddDetailToContact(userID, contactID int, detailType, value string) (*contact_detail.ContactDetail, error) {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	c, err := u.GetContactByID(contactID)
	if err != nil {
		return nil, err
	}

	detail, err := contact_detail.NewContactDetail(c, detailType, value)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func GetContacts(userID int) ([]*contact.Contact, error) {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return u.GetContacts(), nil
}

func GetContactByID(userID, contactID int) (*contact.Contact, error) {
	u, err := user.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return u.GetContactByID(contactID)
}

func GenerateToken(u *user.User) (string, error) {
	return user.GenerateJWT(u)
}
