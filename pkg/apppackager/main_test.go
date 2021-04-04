package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/pkg"
)

func TestPackAndUnpack(t *testing.T) {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	uuidStr := uuidObj.String()
	tmpDir := "/tmp/" + uuidStr
	err = exec.Command("mkdir", tmpDir).Run()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	pkg.CreateSkeltonPackage(tmpDir)

	os.Chdir("/tmp")
	packPkg(tmpDir)
	err = exec.Command("rm", "-rf", tmpDir).Run()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	pkgPath := "/tmp/test-pkg.tar.gz"
	_, err = os.Stat(pkgPath)
	if err != nil {
		t.Fatalf("Package %v does not exist", pkgPath)
	}

	pkgInfo := showPkgInfo(pkgPath)
	if pkgInfo.MetaInfo.Name != "test-pkg" {
		t.Fatalf("Unexpected package name expected:%v actual:%v", "test-pkg", pkgInfo.MetaInfo.Name)
	}
}
