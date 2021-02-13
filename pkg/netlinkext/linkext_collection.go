package netlinkext

import (
	"fmt"
)

// LinkCollection is a concurrent collection for my link
type LinkCollection struct {
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
	c.collection = append(c.collection, link)
}

// Remove removes link from collection
func (c *LinkCollection) Remove(link *LinkExt) error {
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
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *LinkCollection) GetByIndex(index int) *LinkExt {
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *LinkCollection) Where(fn func(*LinkExt) bool) LinkExtSlice {
	return c.collection.Where(fn)
}
