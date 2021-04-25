package capability

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type UserGrantPolicyCollection struct {
	mu         sync.Mutex
	collection UserGrantPolicySlice
}

func NewUserGrantPolicyCollection() *UserGrantPolicyCollection {
	c := UserGrantPolicyCollection{
		mu:         sync.Mutex{},
		collection: UserGrantPolicySlice{},
	}

	return &c
}

func (c *UserGrantPolicyCollection) Add(u *UserGrantPolicy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, u)
}

// Remove removes link from collection
func (c *UserGrantPolicyCollection) Remove(u *UserGrantPolicy) error {
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
func (c *UserGrantPolicyCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *UserGrantPolicyCollection) GetByIndex(index int) *UserGrantPolicy {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *UserGrantPolicyCollection) Where(fn func(*UserGrantPolicy) bool) UserGrantPolicySlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *UserGrantPolicyCollection) GetAll() UserGrantPolicySlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	caps := UserGrantPolicySlice{}
	for idx := range c.collection {
		caps = append(caps, c.collection[idx])
	}

	return caps
}

func (c *UserGrantPolicyCollection) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.collection = UserGrantPolicySlice{}
	return nil
}

func (c *UserGrantPolicyCollection) Contains(u *UserGrantPolicy) bool {
	selectedCaps := c.Where(func(u2 *UserGrantPolicy) bool {
		return u.UserGrantPolicyID == u2.UserGrantPolicyID
	})

	return len(selectedCaps) != 0
}

func (c *UserGrantPolicyCollection) GetByID(uID uuid.UUID) *UserGrantPolicy {
	selectedPolicies := c.Where(func(u2 *UserGrantPolicy) bool {
		return uID == u2.UserGrantPolicyID
	})

	if len(selectedPolicies) != 0 {
		return selectedPolicies[0]
	} else {
		return nil
	}
}

func (c *UserGrantPolicyCollection) IsGranted(cap *Capability, capReq *CapabilityRequest) (bool, bool, error) {
	selectedPolicies := c.Where(func(u *UserGrantPolicy) bool {
		return u.CapabilityID == cap.CapabilityID && u.RequesterID == capReq.RequesterID
	})

	if len(selectedPolicies) > 1 {
		return false, true, fmt.Errorf("unexpected number policies %v", len(selectedPolicies))
	}

	if len(selectedPolicies) == 0 {
		return false, false, nil
	}

	return selectedPolicies[0].Grant, true, nil
}
