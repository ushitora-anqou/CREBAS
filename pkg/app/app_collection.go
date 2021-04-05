package app

import (
	"fmt"
	"sync"
)

// AppCollection is a collection for app
type AppCollection struct {
	mu         sync.Mutex
	collection AppSlice
}

func NewAppCollection() *AppCollection {
	ac := new(AppCollection)
	ac.collection = AppSlice{}

	return ac
}

// Add adds app to collection
func (c *AppCollection) Add(app AppInterface) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, app)
}

// Remove removes link from collection
func (c *AppCollection) Remove(app AppInterface) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == app {
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
func (c *AppCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *AppCollection) GetByIndex(index int) AppInterface {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *AppCollection) Where(fn func(AppInterface) bool) AppSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *AppCollection) GetAll() []AppInterface {
	c.mu.Lock()
	defer c.mu.Unlock()
	apps := []AppInterface{}
	for idx := range c.collection {
		apps = append(apps, c.collection[idx])
	}

	return apps
}

// GetAllAppInfos returns all apps info
func (c *AppCollection) GetAllAppInfos() []*AppInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	apps := []*AppInfo{}
	for idx := range c.collection {
		apps = append(apps, c.collection[idx].GetAppInfo())
	}

	return apps
}

func (c *AppCollection) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.collection = AppSlice{}

	return nil
}

func (c *AppCollection) ClearNotRunningApp() error {
	apps := c.GetAll()
	for idx := range apps {
		app := apps[idx]
		if app.IsRunning() {
			continue
		}
		err := app.Stop()
		if err != nil {
			return err
		}
		err = c.Remove(app)
		if err != nil {
			return err
		}
	}

	return nil
}
