package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/babyapi/storage/kv"
	"github.com/go-chi/render"
	"github.com/tarmac-project/hord/drivers/hashmap"
)

// msgResponse is a simple response type for sending a message
type msgResponse struct {
	Message string `json:"message"`
}

// Render implements render.Renderer
func (m *msgResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Customer represents a CRM customer and supports end-dating (soft delete)
type Customer struct {
	babyapi.DefaultResource
	Name    string     `json:"name"`
	Email   string     `json:"email"`
	Phone   string     `json:"phone,omitempty"`
	EndDate *time.Time `json:"end_date,omitempty"`
}

// EndDated implements babyapi.EndDateable for soft delete
func (c *Customer) EndDated() bool {
	return c.EndDate != nil && c.EndDate.Before(time.Now())
}

func (c *Customer) SetEndDate(t time.Time) {
	c.EndDate = &t
}

func main() {
	db, err := kv.NewFileDB(hashmap.Config{
		Filename: "customers.json",
	})
	if err != nil {
		panic(err)
	}
	api := babyapi.NewAPI("Customers", "/customers", func() *Customer { return &Customer{} })
	api.SetStorage(babyapi.NewKVStorage[*Customer](db, "customer"))
	api.EnableMCP(babyapi.MCPPermCRUD)
	api.AddCustomIDRoute("POST", "/send-email", babyapi.Handler(func(w http.ResponseWriter, r *http.Request) render.Renderer {
		customer, errResp := api.GetRequestedResource(r)
		if errResp != nil {
			return errResp
		}
		// Stub sending email
		// In a real implementation, you would send an email here
		return &msgResponse{Message: fmt.Sprintf("Stub: Sent email to %s (%s)", customer.Name, customer.Email)}
	}))

	api.RunCLI()
}
