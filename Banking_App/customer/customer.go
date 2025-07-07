package customer

import (
	"banking-app/apperror"
	"banking-app/helper"
	"strings"
)

// Customer represents a banking customer
type Customer struct {
	CustomerID   int
	FirstName    string
	LastName     string
	TotalBalance float64
	IsActive     bool
	IsAdmin      bool
}

// CustomerManager handles customer operations
type CustomerManager struct {
	customers         map[int]*Customer
	customerIDCounter int
}

// NewCustomerManager creates a new customer manager instance
func NewCustomerManager() *CustomerManager {
	return &CustomerManager{
		customers:         make(map[int]*Customer),
		customerIDCounter: 0,
	}
}

// isAuthorized checks if caller is active admin
func isAuthorized(caller *Customer) bool {
	return caller != nil && caller.IsActive && caller.IsAdmin
}

// CreateCustomer creates a new customer (Admin-only)
func (cm *CustomerManager) CreateCustomer(caller *Customer, fname, lname string, isAdmin bool) (*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("create customer")
	}

	fname = strings.TrimSpace(fname)
	lname = strings.TrimSpace(lname)
	if fname == "" || lname == "" {
		return nil, apperror.NewValidationError("name", "first and last name required")
	}

	cm.customerIDCounter++
	newCust := &Customer{
		CustomerID:   cm.customerIDCounter,
		FirstName:    fname,
		LastName:     lname,
		TotalBalance: 0,
		IsActive:     true,
		IsAdmin:      isAdmin,
	}

	cm.customers[newCust.CustomerID] = newCust
	return newCust, nil
}

// GetCustomerByID retrieves customer by ID (Admin-only)
func (cm *CustomerManager) GetCustomerByID(caller *Customer, customerID int) (*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("get customer by ID")
	}

	cust, ok := cm.customers[customerID]
	if !ok || !cust.IsActive {
		return nil, apperror.NewNotFoundError("customer", customerID)
	}
	return cust, nil
}

// GetAllCustomersPaginated returns paginated list of active customers (Admin-only)
func (cm *CustomerManager) GetAllCustomersPaginated(caller *Customer, page, pageSize int) ([]*Customer, error) {
	if !isAuthorized(caller) {
		return nil, apperror.NewAuthError("get customer list")
	}

	var activeCustomers []*Customer
	for _, cust := range cm.customers {
		if cust.IsActive {
			activeCustomers = append(activeCustomers, cust)
		}
	}

	start, end := helper.PaginationBounds(page, pageSize, len(activeCustomers))
	if start > len(activeCustomers) {
		return []*Customer{}, nil
	}
	if end > len(activeCustomers) {
		end = len(activeCustomers)
	}

	return activeCustomers[start:end], nil
}

// UpdateCustomer updates customer details (Admin-only)
func (cm *CustomerManager) UpdateCustomer(caller *Customer, updatedCust *Customer) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("update customer")
	}

	if updatedCust == nil {
		return apperror.NewValidationError("customer", "customer cannot be nil")
	}

	cust, ok := cm.customers[updatedCust.CustomerID]
	if !ok || !cust.IsActive {
		return apperror.NewNotFoundError("customer", updatedCust.CustomerID)
	}

	updatedCust.FirstName = strings.TrimSpace(updatedCust.FirstName)
	updatedCust.LastName = strings.TrimSpace(updatedCust.LastName)
	if updatedCust.FirstName == "" || updatedCust.LastName == "" {
		return apperror.NewValidationError("name", "first and last name cannot be empty")
	}

	cust.FirstName = updatedCust.FirstName
	cust.LastName = updatedCust.LastName
	return nil
}

// updateFirstName helper for updating first name
func (cm *CustomerManager) updateFirstName(cust *Customer, value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("firstname", "must be a non-empty string")
	}
	cust.FirstName = strings.TrimSpace(v)
	return nil
}

// updateLastName helper for updating last name
func (cm *CustomerManager) updateLastName(cust *Customer, value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("lastname", "must be a non-empty string")
	}
	cust.LastName = strings.TrimSpace(v)
	return nil
}

// updateIsActive helper for updating active status
func (cm *CustomerManager) updateIsActive(cust *Customer, value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	cust.IsActive = v
	return nil
}

// UpdateCustomerField updates specific customer field (Admin-only)
func (cm *CustomerManager) UpdateCustomerField(caller *Customer, id int, field string, value interface{}) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("update customer field")
	}

	cust, ok := cm.customers[id]
	if !ok || !cust.IsActive {
		return apperror.NewNotFoundError("customer", id)
	}

	switch strings.ToLower(field) {
	case "firstname":
		return cm.updateFirstName(cust, value)
	case "lastname":
		return cm.updateLastName(cust, value)
	case "isactive":
		return cm.updateIsActive(cust, value)
	default:
		return apperror.NewValidationError("field", "unknown customer field")
	}
}

// DeleteCustomer deactivates customer (Admin-only)
func (cm *CustomerManager) DeleteCustomer(caller *Customer, id int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("delete customer")
	}

	cust, ok := cm.customers[id]
	if !ok || !cust.IsActive {
		return apperror.NewNotFoundError("customer", id)
	}

	cust.IsActive = false
	return nil
}

// ReactivateCustomer reactivates deactivated customer (Admin-only)
func (cm *CustomerManager) ReactivateCustomer(caller *Customer, id int) error {
	if !isAuthorized(caller) {
		return apperror.NewAuthError("reactivate customer")
	}

	cust, ok := cm.customers[id]
	if !ok {
		return apperror.NewNotFoundError("customer", id)
	}
	if cust.IsActive {
		return apperror.NewValidationError("reactivate", "customer is already active")
	}

	cust.IsActive = true
	return nil
}

// ForceAddCustomer bypasses auth for system initialization
func (cm *CustomerManager) ForceAddCustomer(cust *Customer) {
	if cust == nil {
		return
	}

	cm.customerIDCounter++
	cust.CustomerID = cm.customerIDCounter
	cm.customers[cust.CustomerID] = cust
}

// GetActiveCustomerCount returns count of active customers (for testing)
func (cm *CustomerManager) GetActiveCustomerCount() int {
	count := 0
	for _, cust := range cm.customers {
		if cust.IsActive {
			count++
		}
	}
	return count
}
