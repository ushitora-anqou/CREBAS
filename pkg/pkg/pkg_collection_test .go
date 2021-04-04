package pkg

import (
	"testing"
)

func TestAdd(t *testing.T) {
	pkgCollection := NewPkgCollection()
	pkg := &PackageInfo{}

	pkgCollection.Add(pkg)
	pkgTest := pkgCollection.GetByIndex(0)
	if pkg != pkgTest {
		t.Fatalf("Failed")
	}
}

func TestRemove(t *testing.T) {
	pkg := &PackageInfo{}
	pkg2 := &PackageInfo{}

	pkgCollection := NewPkgCollection()
	pkgCollection.Add(pkg)
	pkgCollection.Add(pkg2)

	if count := pkgCollection.Count(); count != 2 {
		t.Fatalf("Failed expected:2 actual:#%v", count)
	}

	err := pkgCollection.Remove(pkg)
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	linkTest := pkgCollection.GetByIndex(0)

	if linkTest != pkg2 {
		t.Fatalf("Failed")
	}

	if linkTest == pkg {
		t.Fatalf("Failed")
	}
}
