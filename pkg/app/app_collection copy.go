package app

import (
	"fmt"
	"sync"
)

// DeviceCollection is a collection for device
type DeviceCollection struct {
	mu         sync.Mutex
	collection DeviceSlice
}

func NewDeviceCollection() *DeviceCollection {
	ac := new(DeviceCollection)
	ac.collection = DeviceSlice{}

	return ac
}

// Add adds app to collection
func (c *DeviceCollection) Add(device *Device) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, device)
}

// Remove removes link from collection
func (c *DeviceCollection) Remove(device *Device) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == device {
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
func (c *DeviceCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *DeviceCollection) GetByIndex(index int) *Device {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *DeviceCollection) Where(fn func(*Device) bool) DeviceSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *DeviceCollection) GetAll() DeviceSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	devices := DeviceSlice{}
	for idx := range c.collection {
		devices = append(devices, c.collection[idx])
	}

	return devices
}

func (c *DeviceCollection) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.collection = DeviceSlice{}

	return nil
}
