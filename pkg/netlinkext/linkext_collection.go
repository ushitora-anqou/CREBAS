package netlinkext

import (
	"fmt"
	"sync"
)

// LinkCollection is a concurrent collection for my link
type LinkCollection struct {
	mu         sync.Mutex
	collection LinkExtSlice
}

// NewLinkCollection creates collection for link
func NewLinkCollection() *LinkCollection {
	collection := new(LinkCollection)

	collection.collection = LinkExtSlice{}

	return collection
}

// Add adds link to collection
func (c *LinkCollection) Add(link *LinkExt) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, link)
}

// Remove removes link from collection
func (c *LinkCollection) Remove(link *LinkExt) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == link {
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
func (c *LinkCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *LinkCollection) GetByIndex(index int) *LinkExt {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *LinkCollection) Where(fn func(*LinkExt) bool) LinkExtSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}
