package contact

import (
	"Contact_App/contact_detail"
	"Contact_App/helper"
	"errors"
	"strings"
)

// Interface to avoid import cycle
type ContactOwner interface {
	GetContacts() []*Contact
	IncrementContactIDCounter() int
	AddContact(*Contact)
}

type Contact struct {
	ContactID            int
	FName                string
	LName                string
	IsActive             bool
	Details              []*contact_detail.ContactDetail
	contactDetailCounter int
}

// Factory function removed. Contact creation is done via User.CreateContact()

func (c *Contact) updateFirstName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid firstname")
	}
	c.FName = strings.TrimSpace(v)
	return nil
}

func (c *Contact) updateLastName(value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return errors.New("invalid lastname")
	}
	c.LName = strings.TrimSpace(v)
	return nil
}

func UpdateContactField(owner ContactOwner, requester helper.Authorizer, contactID int, field string, value interface{}) error {
	if !requester.IsStaffActive() {
		return errors.New("only active staff can update contacts")
	}

	for _, c := range owner.GetContacts() {
		if c.ContactID == contactID {
			switch strings.ToLower(field) {
			case "firstname":
				return c.updateFirstName(value)
			case "lastname":
				return c.updateLastName(value)
			default:
				return errors.New("unknown field")
			}
		}
	}
	return errors.New("contact not found")
}

func (c *Contact) GetDetailCounterAndIncrement() int {
	c.contactDetailCounter++
	return c.contactDetailCounter
}

func (c *Contact) AddContactDetail(d *contact_detail.ContactDetail) {
	c.Details = append(c.Details, d)
}

func (c *Contact) GetDetails() []*contact_detail.ContactDetail {
	return c.Details
}
