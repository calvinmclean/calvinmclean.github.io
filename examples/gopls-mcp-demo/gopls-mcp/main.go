package main

import (
	"net/http"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/storage/kv"
)

// Customer represents a CRM customer and supports end-dating
// Implements babyapi.Resource and babyapi.EndDateable
type Customer struct {
	babyapi.DefaultRenderer
	ID      babyapi.ID `json:"id"`
	Name    string     `json:"name"`
	Email   string     `json:"email"`
	Phone   string     `json:"phone"`
	EndDate *time.Time `json:"end_date,omitempty"`
}

func (c *Customer) GetID() string              { return c.ID.String() }
func (c *Customer) ParentID() string           { return "" }
func (c *Customer) Bind(r *http.Request) error { return nil } // Required for interface
func (c *Customer) EndDated() bool             { return c.EndDate != nil && c.EndDate.Before(time.Now()) }
func (c *Customer) SetEndDate(t time.Time)     { c.EndDate = &t }

func main() {
	db, err := kv.NewFileDB(kv.Config{Filename: "customers.json"})
	if err != nil {
		panic(err)
	}
	store := babyapi.NewKVStorage[*Customer](db, "customer")
	api := babyapi.NewAPI[*Customer]("customer", "/customers", func() *Customer {
		return &Customer{ID: babyapi.NewID()}
	})
	api.SetStorage(store)
	api.RunCLI()
}
