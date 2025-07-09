package bank

import (
	"banking-app/apperror"
	"strings"
)

type Bank struct {
	BankID       int
	Name         string
	Abbreviation string
	IsActive     bool
}

func NewBank(bankID int, name string) (*Bank, *apperror.ValidationError) {
	name = strings.TrimSpace(name)
	if bankID < 0 {
		return nil, apperror.NewValidationError("bankID", "please provide a valid bank ID")
	}
	if name == "" {
		return nil, apperror.NewValidationError("name", "fullname of bank cannot be empty")
	}
	if len(name) < 4 {
		return nil, apperror.NewValidationError("name", "bank fullname cannot be less than 4 letters")
	}
	firstTwo := name[:2]
	lastTwo := name[len(name)-2:]
	abbreviation := strings.ToLower(firstTwo + lastTwo)
	return &Bank{
		BankID:       bankID,
		Name:         name,
		Abbreviation: abbreviation,
		IsActive:     true,
	}, nil
}

func (b *Bank) UpdateBankName(newName string) *apperror.ValidationError {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return apperror.NewValidationError("name", "bank new fullname cannot be empty")
	}
	if len(newName) < 4 {
		return apperror.NewValidationError("name", "bank new fullname cannot be less than 4 letters")
	}
	b.Name = newName
	firstTwo := newName[:2]
	lastTwo := newName[len(newName)-2:]
	b.Abbreviation = strings.ToLower(firstTwo + lastTwo)
	return nil
}
