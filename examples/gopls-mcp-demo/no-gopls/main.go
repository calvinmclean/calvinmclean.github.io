package main

import (
	"log"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/storage/kv"
	"github.com/tarmac-project/hord/hashmap"
)

// Customer represents a CRM customer resource.
type Customer struct {
	babyapi.DefaultResource

	Name    string     `json:"name"`
	Email   string     `json:"email"`
	Phone   string     `json:"phone"`
	EndDate *time.Time `json:"end_date,omitempty"`
}

// EndDated implements babyapi.EndDateable for soft-delete support.
func (c *Customer) EndDated() bool {
	return c.EndDate != nil
}

// SetEndDate sets the end date for soft-delete.
func (c *Customer) SetEndDate(t time.Time) {
	c.EndDate = &t
}

func main() {
	// Create the API for customers
	api := babyapi.NewAPI(
		"Customers", "/customers",
		func() *Customer { return &Customer{} },
	)

	// Use file-based storage (JSON file)
	db, err := kv.NewFileDB(hashmap.Config{
		Filename: "customers.json",
	})
	if err != nil {
		log.Fatalf("failed to create file DB: %v", err)
	}
	api.SetStorage(babyapi.NewKVStorage[*Customer](db, "Customer"))

	// Enable MCP CRUD tools (optional, but recommended for full babyapi features)
	api.EnableMCP(babyapi.MCPPermCRUD)

	// Run CLI (enables both server and client commands)
	api.RunCLI()
}
