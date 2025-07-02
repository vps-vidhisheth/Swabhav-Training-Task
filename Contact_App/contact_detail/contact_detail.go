package contact_detail

import (
	"Contact_App/helper"
	"errors"
	"strings"
)

type ContactDetail struct {
	ContactDetailsID int
	Type             string
	Value            string
	IsActive         bool
}

// --- INTERFACES ---

type Contact interface {
	GetDetailCounterAndIncrement() int
	AddContactDetail(*ContactDetail)
	GetDetails() []*ContactDetail
}

type ContactOwner interface {
	IsAdminUser() bool
	IsActiveUser() bool
	GetContactByID(id int) (Contact, error)
}

type Authorizer interface {
	IsAdminUser() bool
	IsActiveUser() bool
}

// --- FACTORY METHOD ---

func NewContactDetail(c Contact, typ string, val string) (*ContactDetail, error) {
	typ, val = strings.ToLower(strings.TrimSpace(typ)), strings.TrimSpace(val)
	if (typ != "email" && typ != "phone") || val == "" {
		return nil, errors.New("invalid contact detail")
	}
	id := c.GetDetailCounterAndIncrement()
	d := &ContactDetail{
		ContactDetailsID: id,
		Type:             typ,
		Value:            val,
		IsActive:         true,
	}
	c.AddContactDetail(d)
	return d, nil
}

// --- UPDATE METHODS ---

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

// --- MAIN FIELD UPDATE FUNCTION ---

func UpdateContactDetailField(
	owner ContactOwner,
	requester Authorizer,
	contactID, detailID int,
	field string,
	value interface{},
) error {
	if !helper.IsAuthorizedStaff(helper.UserData{
		IsAdmin:  requester.IsAdminUser(),
		IsActive: requester.IsActiveUser(),
	}) {
		return errors.New("only active staff can update contact details")
	}

	contact, err := owner.GetContactByID(contactID)
	if err != nil {
		return err
	}

	for _, d := range contact.GetDetails() {
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
