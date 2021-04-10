package pkg

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type PkgCollection struct {
	mu         sync.Mutex
	collection PkgSlice
}

func NewPkgCollection() *PkgCollection {
	pc := new(PkgCollection)
	pc.collection = PkgSlice{}

	return pc
}

// Add adds app to collection
func (c *PkgCollection) Add(pkg *PackageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, pkg)
}

// Remove removes link from collection
func (c *PkgCollection) Remove(pkg *PackageInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	removeIndex := -1
	for idx, l := range c.collection {
		if l == pkg {
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
func (c *PkgCollection) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.collection)
}

// GetByIndex returns index's element
func (c *PkgCollection) GetByIndex(index int) *PackageInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	link := c.collection[index]
	return link
}

// Where returns a first Link which returns true for func
func (c *PkgCollection) Where(fn func(*PackageInfo) bool) PkgSlice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection.Where(fn)
}

func (c *PkgCollection) LoadPkgs(loadDirPath string) error {
	files, err := ioutil.ReadDir(loadDirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "tar.gz") {
			pkgPathAbs, err := filepath.Abs(filepath.Join(loadDirPath, file.Name()))
			if err != nil {
				return err
			}
			pkgInfo, err := UnpackPkg(pkgPathAbs)
			if err != nil {
				return err
			}
			pkgInfo.UnpackedPkgPath = ""
			c.Add(pkgInfo)
			err = exec.Command("rm", "-rf", pkgInfo.UnpackedPkgPath).Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *PkgCollection) GetAll() []*PackageInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	pkgs := []*PackageInfo{}
	for idx := range c.collection {
		pkgs = append(pkgs, c.collection[idx])
	}

	return pkgs
}

func (c *PkgCollection) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.collection = PkgSlice{}
	return nil
}
