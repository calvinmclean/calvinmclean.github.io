package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/calvinmclean/babyapi"
)

// Customer represents a CRM customer and supports end-dating (soft delete)
type Customer struct {
	babyapi.DefaultRenderer

	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// --- babyapi.Resource implementation ---

func (c *Customer) GetID() string {
	return c.ID
}

func (c *Customer) ParentID() string {
	return ""
}

// --- babyapi.RendererBinder implementation ---

func (c *Customer) Bind(_ *http.Request) error {
	if c.ID == "" {
		c.ID = babyapi.NewID().String()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	return nil
}

func (c *Customer) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// --- babyapi.EndDateable implementation ---

func (c *Customer) EndDated() bool {
	return c.EndDate != nil && !c.EndDate.IsZero()
}

func (c *Customer) SetEndDate(t time.Time) {
	c.EndDate = &t
}

// --- JSON file storage implementation ---

type CustomerFileStorage struct {
	mu        sync.Mutex
	filePath  string
	customers map[string]*Customer
}

func NewCustomerFileStorage(filePath string) (*CustomerFileStorage, error) {
	s := &CustomerFileStorage{
		filePath:  filePath,
		customers: make(map[string]*Customer),
	}
	if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return s, nil
}

func (s *CustomerFileStorage) load() error {
	f, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	return dec.Decode(&s.customers)
}

func (s *CustomerFileStorage) save() error {
	f, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s.customers)
}

func (s *CustomerFileStorage) Get(_ context.Context, id string) (*Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.customers[id]
	if !ok {
		return nil, babyapi.ErrNotFound
	}
	return c, nil
}

func (s *CustomerFileStorage) Search(_ context.Context, _ string, query url.Values) ([]*Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*Customer
	includeEndDated := false
	if v, ok := query["end_dated"]; ok && len(v) > 0 && v[0] == "true" {
		includeEndDated = true
	}
	for _, c := range s.customers {
		if !includeEndDated && c.EndDated() {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *CustomerFileStorage) Set(_ context.Context, c *Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customers[c.ID] = c
	return s.save()
}

func (s *CustomerFileStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.customers[id]
	if !ok {
		return babyapi.ErrNotFound
	}
	// Soft-delete if possible
	if ed, ok := interface{}(c).(babyapi.EndDateable); ok && !ed.EndDated() {
		ed.SetEndDate(time.Now().UTC())
		return s.save()
	}
	// Hard-delete if already end-dated
	delete(s.customers, id)
	return s.save()
}

// --- Main ---

func main() {
	storage, err := NewCustomerFileStorage("customers.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load storage: %v\n", err)
		os.Exit(1)
	}

	api := babyapi.NewAPI[*Customer](
		"customer",
		"/customers",
		func() *Customer { return &Customer{} },
	)

	api.SetStorage(storage)
	api.EnableMCP(babyapi.MCPPermCRUD)
	api.RunCLI()
}
