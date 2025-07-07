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

	// Internals for managing customers (map & counter)
	customers         map[int]*Customer
	customerIDCounter int
}

// Constructor initializes the manager with internal map and counter
func NewCustomerManager() *Customer {
	return &Customer{
		customers:         make(map[int]*Customer),
		customerIDCounter: 0,
	}
}

// CreateCustomer creates and stores a new customer
func (c *Customer) CreateCustomer(fname, lname string) (*Customer, error) {
	fname = strings.TrimSpace(fname)
	lname = strings.TrimSpace(lname)
	if fname == "" || lname == "" {
		return nil, apperror.NewValidationError("name", "first and last name required")
	}

	c.customerIDCounter++
	newCust := &Customer{
		CustomerID:   c.customerIDCounter,
		FirstName:    fname,
		LastName:     lname,
		TotalBalance: 0,
		IsActive:     true,
	}

	c.customers[newCust.CustomerID] = newCust
	return newCust, nil
}

// GetCustomerByID fetches an active customer by ID
func (c *Customer) GetCustomerByID(customerID int) (*Customer, error) {
	cust, ok := c.customers[customerID]
	if !ok || !cust.IsActive {
		return nil, apperror.NewNotFoundError("customer", customerID)
	}
	return cust, nil
}

// GetAllCustomersPaginated returns paginated active customers
func (c *Customer) GetAllCustomersPaginated(page, pageSize int) ([]*Customer, error) {
	var active []*Customer
	for _, cust := range c.customers {
		if cust.IsActive {
			active = append(active, cust)
		}
	}
	start, end := helper.PaginationBounds(page, pageSize, len(active))
	if start > end || start < 0 || end > len(active) {
		return nil, apperror.NewValidationError("pagination", "invalid page or pageSize")
	}
	return active[start:end], nil
}

// UpdateCustomerName updates the first and last name of a customer by ID
func (c *Customer) UpdateCustomerName(id int, newFName, newLName string) error {
	cust, err := c.GetCustomerByID(id)
	if err != nil {
		return err
	}
	newFName = strings.TrimSpace(newFName)
	newLName = strings.TrimSpace(newLName)
	if newFName == "" || newLName == "" {
		return apperror.NewValidationError("name", "names cannot be empty")
	}
	cust.FirstName = newFName
	cust.LastName = newLName
	return nil
}

// UpdateCustomer updates the customer details from the given customer struct
func (c *Customer) UpdateCustomer(updatedCust *Customer) error {
	if updatedCust == nil {
		return apperror.NewValidationError("customer", "updated customer cannot be nil")
	}
	cust, err := c.GetCustomerByID(updatedCust.CustomerID)
	if err != nil {
		return err
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

// UpdateCustomerField updates a specific field of a customer
func (c *Customer) UpdateCustomerField(cust *Customer, field string, value interface{}) error {
	switch strings.ToLower(field) {
	case "firstname":
		return c.updateFirstName(cust, value)
	case "lastname":
		return c.updateLastName(cust, value)
	case "isactive":
		return c.updateIsActive(cust, value)
	default:
		return apperror.NewValidationError("field", "unknown customer field")
	}
}

func (c *Customer) updateFirstName(cust *Customer, value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("firstname", "must be a non-empty string")
	}
	cust.FirstName = strings.TrimSpace(v)
	return nil
}

func (c *Customer) updateLastName(cust *Customer, value interface{}) error {
	v, ok := value.(string)
	if !ok || strings.TrimSpace(v) == "" {
		return apperror.NewValidationError("lastname", "must be a non-empty string")
	}
	cust.LastName = strings.TrimSpace(v)
	return nil
}

func (c *Customer) updateIsActive(cust *Customer, value interface{}) error {
	v, ok := value.(bool)
	if !ok {
		return apperror.NewValidationError("isactive", "must be a boolean")
	}
	cust.IsActive = v
	return nil
}

// Soft Delete a customer by setting IsActive to false
func (c *Customer) DeleteCustomer(id int) error {
	cust, err := c.GetCustomerByID(id)
	if err != nil {
		return err
	}
	cust.IsActive = false
	return nil
}

// Reactivate a soft-deleted customer
func (c *Customer) ReactivateCustomer(id int) error {
	cust, err := c.GetCustomerByID(id)
	if err != nil {
		return err
	}
	cust.IsActive = true
	return nil
}
