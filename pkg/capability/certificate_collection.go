package capability

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type AppCertificateCollection struct {
	mu         sync.Mutex
	collection AppCertificateSlice
}

func NewAppCertificateCollection() *AppCertificateCollection {
	c := AppCertificateCollection{
		mu:         sync.Mutex{},
		collection: AppCertificateSlice{},
	}

	return &c
}

func (c *AppCertificateCollection) Add(u *AppCertificate) {
	if c.Contains(u) {
		c.Remove(u)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, u)
}

// Remove removes link from collection
func (c *AppCertificateCollection) Remove(u *AppCertificate) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == u {
			removeIndex = idx
			break
		}
	}

	if removeIndex < 0 {
		return fmt.Errorf("element not found in collection")
	}
	c.collection = append(c.collection[:removeIndex], c.collection[removeIndex+1:]...)
	return nil
}

// Count returns length of collection
func (c *AppCertificateCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *AppCertificateCollection) GetByIndex(index int) *AppCertificate {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *AppCertificateCollection) Where(fn func(*AppCertificate) bool) AppCertificateSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *AppCertificateCollection) GetAll() AppCertificateSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	caps := AppCertificateSlice{}
	for idx := range c.collection {
		caps = append(caps, c.collection[idx])
	}

	return caps
}

func (c *AppCertificateCollection) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.collection = AppCertificateSlice{}
	return nil
}

func (c *AppCertificateCollection) Contains(u *AppCertificate) bool {
	selectedCaps := c.Where(func(u2 *AppCertificate) bool {
		return u.AppID == u2.AppID
	})

	return len(selectedCaps) != 0
}

func (c *AppCertificateCollection) GetByID(uID uuid.UUID) *AppCertificate {
	selectedPolicies := c.Where(func(u2 *AppCertificate) bool {
		return uID == u2.AppID
	})

	if len(selectedPolicies) != 0 {
		return selectedPolicies[0]
	} else {
		return nil
	}
}
