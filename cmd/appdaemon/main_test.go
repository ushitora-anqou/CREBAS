package main

import (
	"testing"
	"time"

	"github.com/naoki9911/CREBAS/pkg/pkg"
)

func TestStartChildProc(t *testing.T) {
	pkgInfo := pkg.CreateSkeltonPackageInfo()
	pkgInfo.MetaInfo.CMD = []string{"echo", "hello"}

	err := start(pkgInfo)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	exitCode := stop()
	if exitCode != 0 {
		t.Fatalf("Failed exit code %v", exitCode)
	}
}

func TestStartChildProc2(t *testing.T) {
	pkgInfo := pkg.CreateSkeltonPackageInfo()
	pkgInfo.MetaInfo.CMD = []string{"sleep", "10"}

	err := start(pkgInfo)
	if err != nil {
		t.Fatalf("Failed %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	exitCode := stop()
	if exitCode == 0 {
		t.Fatalf("Failed exit code %v", exitCode)
	}
}
