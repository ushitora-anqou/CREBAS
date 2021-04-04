package pkg

import (
	"os/exec"
	"path/filepath"
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

func TestLoadPackages(t *testing.T) {
	testPkgsDir := "/tmp/pkg_collection_test"
	exec.Command("rm", "-rf", testPkgsDir).Run()
	exec.Command("mkdir", "-p", testPkgsDir).Run()

	pkgInfo := CreateSkeltonPackageInfo()
	err := CreatePackage(pkgInfo, testPkgsDir)
	if err != nil {
		panic(err)
	}

	pkgCollection := NewPkgCollection()
	err = pkgCollection.LoadPkgs(testPkgsDir)
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	if pkgCollection.Count() != 1 {
		t.Fatalf("pkgCollection count is not 1")
	}
	pkgInfoTest := pkgCollection.GetByIndex(0)

	packageName := pkgInfo.MetaInfo.Name + ".tar.gz"
	packagePath := filepath.Join(testPkgsDir, packageName)
	if pkgInfoTest.PkgPath != packagePath {
		t.Fatalf("Unequal PkgPath expected:%v actual:%v", packagePath, pkgInfoTest.PkgPath)
	}

	if pkgInfoTest.MetaInfo.Name != pkgInfo.MetaInfo.Name {
		t.Fatalf("Unequal pkg name expected:%v actual:%v", pkgInfo.MetaInfo.Name, pkgInfoTest.MetaInfo.Name)
	}

	pkgInfoTest2 := pkgCollection.GetAll()[0]
	if pkgInfoTest2.MetaInfo.PkgID != pkgInfo.MetaInfo.PkgID {
		t.Fatalf("Unexpected elements at GetAll()")
	}

	exec.Command("rm", "-rf", testPkgsDir).Run()
}
