package app

import "fmt"

// AppCollection is a collection for app
type AppCollection struct {
	collection AppSlice
}

func NewAppCollection() *AppCollection {
	ac := new(AppCollection)
	ac.collection = AppSlice{}

	return ac
}

// Add adds app to collection
func (c *AppCollection) Add(app AppInterface) {
	c.collection = append(c.collection, app)
}

// Remove removes link from collection
func (c *AppCollection) Remove(app AppInterface) error {
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
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *AppCollection) GetByIndex(index int) AppInterface {
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *AppCollection) Where(fn func(AppInterface) bool) AppSlice {
	return c.collection.Where(fn)
}

// GetAll returns all apps
func (c *AppCollection) GetAll() []AppInterface {
	apps := []AppInterface{}
	for idx := range c.collection {
		apps = append(apps, c.collection[idx])
	}

	return apps
}

// GetAllAppInfos returns all apps info
func (c *AppCollection) GetAllAppInfos() []*AppInfo {
	apps := []*AppInfo{}
	for idx := range c.collection {
		apps = append(apps, c.collection[idx].GetAppInfo())
	}

	return apps
}
