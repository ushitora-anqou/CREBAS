package pkg

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
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
	exec.Command("rm", "-rf", "/tmp/pkg_collection_test").Run()

	uuid, _ := uuid.NewRandom()
	tmpPkgDir := "/tmp/pkg_collection_test/" + uuid.String()
	err := exec.Command("mkdir", "-p", tmpPkgDir).Run()
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	prevDir, _ := filepath.Abs(".")
	os.Chdir("/tmp/pkg_collection_test")

	pkgInfo := CreateSkeltonPackage(tmpPkgDir)

	packageName := pkgInfo.MetaInfo.Name + ".tar.gz"
	log.Printf("info: Packing application to %v", packageName)
	err = exec.Command("tar", "-zcvf", packageName, "-C", tmpPkgDir+"/", ".").Run()
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	pkgCollection := NewPkgCollection()
	err = pkgCollection.LoadPkgs("/tmp/pkg_collection_test")
	if err != nil {
		t.Fatalf("Failed error:#%v", err)
	}

	if pkgCollection.Count() != 1 {
		t.Fatalf("pkgCollection count is not 1")
	}
	pkgInfoTest := pkgCollection.GetByIndex(0)

	packagePath := filepath.Join("/tmp/pkg_collection_test", packageName)
	if pkgInfoTest.PkgPath != packagePath {
		t.Fatalf("Unequal PkgPath expected:%v actual:%v", packagePath, pkgInfoTest.PkgPath)
	}

	if pkgInfoTest.MetaInfo.Name != pkgInfo.MetaInfo.Name {
		t.Fatalf("Unequal pkg name expected:%v actual:%v", pkgInfo.MetaInfo.Name, pkgInfoTest.MetaInfo.Name)
	}

	os.Chdir(prevDir)

	exec.Command("rm", "-rf", "/tmp/pkg_collection_test").Run()
}
