package customer

import (
	"banking-app/apperror"
	"banking-app/helper"
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
		return nil, apperror.NewAuthError("create customer")
	}
	fname = strings.TrimSpace(fname)
	lname = strings.TrimSpace(lname)
	if fname == "" || lname == "" {
		return nil, apperror.NewValidationError("name", "first and last name required")
	}
	m.nextID++
	cust := &Customer{
		CustomerID:   m.nextID,
		FirstName:    fname,
		LastName:     lname,
		IsActive:     true,
		IsAdmin:      isAdmin,
		TotalBalance: 0,
	}
	m.customers[cust.CustomerID] = cust
	return cust, nil
}

func (m *Manager) GetByID(caller *Customer, id int) (*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("get customer")
	}
	cust, ok := m.customers[id]
	if !ok || !cust.IsActive {
		return nil, apperror.NewNotFoundError("customer", id)
	}
	return cust, nil
}

func (m *Manager) GetAllPaginated(caller *Customer, page, pageSize int) ([]*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("list customers")
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
		return apperror.NewAuthError("update customer")
	}
	if updated == nil {
		return apperror.NewValidationError("customer", "cannot be nil")
	}
	cust, ok := m.customers[updated.CustomerID]
	if !ok || !cust.IsActive {
		return apperror.NewNotFoundError("customer", updated.CustomerID)
	}
	updated.FirstName = strings.TrimSpace(updated.FirstName)
	updated.LastName = strings.TrimSpace(updated.LastName)
	if updated.FirstName == "" || updated.LastName == "" {
		return apperror.NewValidationError("name", "first and last name required")
	}
	cust.FirstName = updated.FirstName
	cust.LastName = updated.LastName
	return nil
}

func (m *Manager) UpdateField(caller *Customer, id int, field string, value interface{}) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("update field")
	}
	cust, ok := m.customers[id]
	if !ok || !cust.IsActive {
		return apperror.NewNotFoundError("customer", id)
	}

	switch strings.ToLower(field) {
	case "firstname":
		return m.updateFirstName(cust, value)
	case "lastname":
		return m.updateLastName(cust, value)
	case "isactive":
		return m.updateIsActive(cust, value)
	default:
		return apperror.NewValidationError("field", "unknown field")
	}
}

// --- Field update helper methods ---

func (m *Manager) updateFirstName(c *Customer, val interface{}) error {
	v, ok := val.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("firstname", "must be a non-empty string")
	}
	c.FirstName = strings.TrimSpace(v)
	return nil
}

func (m *Manager) updateLastName(c *Customer, val interface{}) error {
	v, ok := val.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("lastname", "must be a non-empty string")
	}
	c.LastName = strings.TrimSpace(v)
	return nil
}

func (m *Manager) updateIsActive(c *Customer, val interface{}) error {
	v, ok := val.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	c.IsActive = v
	return nil
}

// --- Lifecycle methods ---

func (m *Manager) Delete(caller *Customer, id int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("delete customer")
	}
	cust, ok := m.customers[id]
	if !ok || !cust.IsActive {
		return apperror.NewNotFoundError("customer", id)
	}
	cust.IsActive = false
	return nil
}

func (m *Manager) Reactivate(caller *Customer, id int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("reactivate customer")
	}
	cust, ok := m.customers[id]
	if !ok {
		return apperror.NewNotFoundError("customer", id)
	}
	if cust.IsActive {
		return apperror.NewValidationError("reactivate", "already active")
	}
	cust.IsActive = true
	return nil
}

// --- Utilities ---

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
