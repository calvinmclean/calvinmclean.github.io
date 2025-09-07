package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/storage/kv"
	"github.com/tarmac-project/hord/drivers/hashmap"
)

// Customer represents a CRM customer and supports end-dating (soft delete).
type Customer struct {
	babyapi.DefaultResource

	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone,omitempty"`
	Company   string     `json:"company,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// Bind implements render.Binder for input validation and initialization.
func (c *Customer) Bind(r *http.Request) error {
	if c.Name == "" {
		return babyapi.ErrInvalidRequest(errors.New("name is required"))
	}
	if c.Email == "" {
		return babyapi.ErrInvalidRequest(errors.New("email is required"))
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	return nil
}

// EndDated implements babyapi.EndDateable.
func (c *Customer) EndDated() bool {
	return c.EndDate != nil && !c.EndDate.IsZero()
}

// SetEndDate implements babyapi.EndDateable.
func (c *Customer) SetEndDate(t time.Time) {
	c.EndDate = &t
}

func main() {
	// Use babyapi's kv storage with a JSON file backend
	db, err := kv.NewFileDB(hashmap.Config{
		Filename: "customers.json",
	})
	if err != nil {
		panic(err)
	}

	// Set up babyapi storage with end-dating support
	storage := babyapi.NewKVStorage[*Customer](db, "customer")

	// Create the API for Customer
	api := babyapi.NewAPI[*Customer](
		"customer",
		"/customers",
		func() *Customer { return &Customer{} },
	).
		SetStorage(storage).
		EnableMCP(babyapi.MCPPermCRUD)

	// Run as CLI (server or client) using babyapi's built-in CLI support
	api.RunCLI()
}
