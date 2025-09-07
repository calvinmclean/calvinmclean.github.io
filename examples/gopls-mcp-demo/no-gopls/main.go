package main

import (
	"net/http"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/storage/kv"
	"github.com/tarmac-project/hord/drivers/hashmap"
)

// Customer represents a CRM customer and supports end-dating (soft-delete)
type Customer struct {
	babyapi.DefaultResource

	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Company   string    `json:"company,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// EndDate is used for soft-delete (end-dating)
	EndDate *time.Time `json:"end_date,omitempty"`
}

// EndDated returns true if the resource has been end-dated (soft-deleted)
func (c *Customer) EndDated() bool {
	return c.EndDate != nil
}

// SetEndDate sets the end date for soft-deleting the resource
func (c *Customer) SetEndDate(t time.Time) {
	c.EndDate = &t
}

// Bind implements input validation and sets timestamps
func (c *Customer) Bind(_ *http.Request) error {
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	return nil
}

func main() {
	// Set up file-based storage for customers
	db, err := kv.NewFileDB(hashmap.Config{
		Filename: "customers.json",
	})
	if err != nil {
		panic(err)
	}

	api := babyapi.NewAPI(
		"Customers", "/customers",
		func() *Customer { return &Customer{} },
	)
	api.SetStorage(babyapi.NewKVStorage[*Customer](db, "Customer"))

	// Enable CLI (server/client) and end-dating support
	api.RunCLI()
}
