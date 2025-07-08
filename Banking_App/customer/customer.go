package customer

import (
	"banking-app/apperror"
	"banking-app/helper"
	"fmt"
	"strings"
)

type Customer struct {
	CustomerID   int
	FirstName    string
	LastName     string
	TotalBalance float64
	IsActive     bool
	IsAdmin      bool
}

type Manager struct {
	customers map[int]*Customer
	nextID    int
}

func NewManager() *Manager {
	return &Manager{
		customers: make(map[int]*Customer),
		nextID:    0,
	}
}

func isAuthorized(c *Customer) bool {
	return c != nil && c.IsActive && c.IsAdmin
}

func (m *Manager) Create(caller *Customer, fname, lname string, isAdmin bool) (*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("create customer: caller not authorized")
	}

	trimmedFname := strings.TrimSpace(fname)
	trimmedLname := strings.TrimSpace(lname)

	if trimmedFname == "" || trimmedLname == "" {
		return nil, apperror.NewValidationError("name", "first and last name cannot be empty")
	}

	m.nextID++
	cust := &Customer{
		CustomerID:   m.nextID,
		FirstName:    trimmedFname,
		LastName:     trimmedLname,
		IsActive:     true,
		IsAdmin:      isAdmin,
		TotalBalance: 0,
	}
	m.customers[cust.CustomerID] = cust
	return cust, nil
}

func GetCustomerByID(id int) (*Customer, error) {
	return defaultManager.GetByID(nil, id)
}

var defaultManager = NewManager()

func (m *Manager) GetByID(caller *Customer, id int) (*Customer, error) {
	if caller != nil && !isAuthorized(caller) {
		return nil, apperror.NewAuthError("get customer: caller not authorized")
	}

	cust, ok := m.customers[id]
	if !ok {
		return nil, apperror.NewNotFoundError("customer", id)
	}
	if !cust.IsActive {
		return nil, apperror.NewCustomerError("get customer", fmt.Sprintf("customer with ID %d is inactive", id))
	}
	return cust, nil
}

func (m *Manager) GetAllPaginated(caller *Customer, page, pageSize int) ([]*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("list customers: caller not authorized")
	}

	var active []*Customer
	for _, cust := range m.customers {
		if cust.IsActive {
			active = append(active, cust)
		}
	}

	start, end := helper.PaginationBounds(page, pageSize, len(active))
	if start > len(active) {
		return []*Customer{}, nil
	}
	if end > len(active) {
		end = len(active)
	}
	return active[start:end], nil
}

func (m *Manager) Update(caller *Customer, updated *Customer) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("update customer: caller not authorized")
	}
	if updated == nil {
		return apperror.NewValidationError("customer", "customer cannot be nil for update")
	}

	cust, ok := m.customers[updated.CustomerID]
	if !ok {
		return apperror.NewNotFoundError("customer", updated.CustomerID)
	}
	if !cust.IsActive {
		return apperror.NewCustomerError("update customer", fmt.Sprintf("customer with ID %d is inactive", updated.CustomerID))
	}

	updated.FirstName = strings.TrimSpace(updated.FirstName)
	updated.LastName = strings.TrimSpace(updated.LastName)

	if updated.FirstName == "" || updated.LastName == "" {
		return apperror.NewValidationError("name", "first and last name are required for update")
	}

	cust.FirstName = updated.FirstName
	cust.LastName = updated.LastName
	return nil
}

func (m *Manager) UpdateField(caller *Customer, id int, field string, value interface{}) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("update customer field: caller not authorized")
	}

	cust, ok := m.customers[id]
	if !ok {
		return apperror.NewNotFoundError("customer", id)
	}
	if !cust.IsActive {
		return apperror.NewCustomerError("update customer field", fmt.Sprintf("customer with ID %d is inactive", id))
	}

	switch strings.ToLower(field) {
	case "firstname":
		return m.updateFirstName(cust, value)
	case "lastname":
		return m.updateLastName(cust, value)
	case "isactive":
		return m.updateIsActive(cust, value)
	default:
		return apperror.NewValidationError("field", fmt.Sprintf("unknown update field: '%s'", field))
	}
}

func (m *Manager) updateFirstName(c *Customer, val interface{}) error {
	v, ok := val.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("firstname", "first name must be a non-empty string")
	}
	c.FirstName = strings.TrimSpace(v)
	return nil
}

func (m *Manager) updateLastName(c *Customer, val interface{}) error {
	v, ok := val.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("lastname", "last name must be a non-empty string")
	}
	c.LastName = strings.TrimSpace(v)
	return nil
}

func (m *Manager) updateIsActive(c *Customer, val interface{}) error {
	v, ok := val.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "active status must be a boolean")
	}
	c.IsActive = v
	return nil
}

func (m *Manager) Delete(caller *Customer, id int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("delete customer: caller not authorized")
	}
	cust, ok := m.customers[id]
	if !ok {
		return apperror.NewNotFoundError("customer", id)
	}
	if !cust.IsActive {
		return apperror.NewCustomerError("delete customer", fmt.Sprintf("customer with ID %d is already inactive", id))
	}
	cust.IsActive = false
	return nil
}

func (m *Manager) Reactivate(caller *Customer, id int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("reactivate customer: caller not authorized")
	}
	cust, ok := m.customers[id]
	if !ok {
		return apperror.NewNotFoundError("customer", id)
	}
	if cust.IsActive {
		return apperror.NewValidationError("reactivate", fmt.Sprintf("customer with ID %d is already active", id))
	}
	cust.IsActive = true
	return nil
}

func (m *Manager) ForceAdd(c *Customer) {
	if c == nil {
		return
	}
	m.nextID++
	c.CustomerID = m.nextID
	m.customers[c.CustomerID] = c
}

func (m *Manager) CountActive() int {
	count := 0
	for _, c := range m.customers {
		if c.IsActive {
			count++
		}
	}
	return count
}
