package capability

import (
	"fmt"
	"sync"
)

type CapabilityCollection struct {
	mu         sync.Mutex
	collection CapabilitySlice
}

type CapabilityRequestCollection struct {
	mu         sync.Mutex
	collection CapabilityRequestSlice
}

func NewCapabilityCollection() *CapabilityCollection {
	c := CapabilityCollection{
		mu:         sync.Mutex{},
		collection: CapabilitySlice{},
	}

	return &c
}

func (c *CapabilityCollection) Add(cap *Capability) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, cap)
}

// Remove removes link from collection
func (c *CapabilityCollection) Remove(cap *Capability) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == cap {
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
func (c *CapabilityCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *CapabilityCollection) GetByIndex(index int) *Capability {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *CapabilityCollection) Where(fn func(*Capability) bool) CapabilitySlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *CapabilityCollection) GetAll() CapabilitySlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	caps := CapabilitySlice{}
	for idx := range c.collection {
		caps = append(caps, c.collection[idx])
	}

	return caps
}

func NewCapabilityRequestCollection() *CapabilityRequestCollection {
	c := CapabilityRequestCollection{
		mu:         sync.Mutex{},
		collection: CapabilityRequestSlice{},
	}

	return &c
}

func (c *CapabilityRequestCollection) Add(cap *CapabilityRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, cap)
}

// Remove removes link from collection
func (c *CapabilityRequestCollection) Remove(cap *CapabilityRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == cap {
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
func (c *CapabilityRequestCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *CapabilityRequestCollection) GetByIndex(index int) *CapabilityRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *CapabilityRequestCollection) Where(fn func(*CapabilityRequest) bool) CapabilityRequestSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *CapabilityRequestCollection) GetAll() CapabilityRequestSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	caps := CapabilityRequestSlice{}
	for idx := range c.collection {
		caps = append(caps, c.collection[idx])
	}

	return caps
}
