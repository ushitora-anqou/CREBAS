package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/stretchr/testify/assert"
)

var router = setupRouter()

func TestGetAllAppInfos(t *testing.T) {
	ap, err := app.NewLinuxProcess()
	if err != nil {
		t.Fatalf("Failed %v", err)
	}
	defer ap.Stop()
	apps.Add(ap)
	defer apps.Clear()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/apps", nil)

	router.ServeHTTP(w, req)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	var appInfos []app.AppInfo
	json.Unmarshal(resbody, &appInfos)

	var expectedAppInfos = apps.GetAllAppInfos()
	for id, appInfo := range appInfos {
		assert.Equal(t, appInfo.Id, expectedAppInfos[id].Id, "unmatched ID")
	}
}

func TestGetAllPkgs(t *testing.T) {
	testPkgsDir := "/tmp/pep_test"
	exec.Command("mkdir", "-p", testPkgsDir).Run()
	defer exec.Command("rm", "-rf", testPkgsDir).Run()
	defer pkgs.Clear()
	defer apps.Clear()

	pkgInfo := pkg.CreateSkeltonPackageInfo()
	err := pkg.CreatePackage(pkgInfo, testPkgsDir)
	if err != nil {
		panic(err)
	}

	err = pkgs.LoadPkgs(testPkgsDir)
	if err != nil {
		t.Fatalf("Failed error:%v", err)
	}

	req := httptest.NewRequest("GET", "/pkgs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	var pkgInfos []pkg.PackageInfo
	json.Unmarshal(resbody, &pkgInfos)

	assert.Equal(t, pkgInfos[0].MetaInfo.PkgID, pkgInfo.MetaInfo.PkgID, "unmatched ID")
	exec.Command("rm", "-rf", testPkgsDir).Run()

	exec.Command("rm", "-rf", testPkgsDir).Run()
}
