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
	apps.Add(app.NewLinuxProcess())

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
	exec.Command("rm", "-rf", testPkgsDir).Run()
	exec.Command("mkdir", "-p", testPkgsDir).Run()
	pkgs.Clear()
	apps.Clear()

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

func TestStartAppFromPkg(t *testing.T) {
	testPkgsDir := "/tmp/pep_test"
	exec.Command("rm", "-rf", testPkgsDir).Run()
	exec.Command("mkdir", "-p", testPkgsDir).Run()
	pkgs.Clear()
	apps.Clear()

	pkgInfo := pkg.CreateSkeltonPackageInfo()
	err := pkg.CreatePackage(pkgInfo, testPkgsDir)
	if err != nil {
		panic(err)
	}

	err = pkgs.LoadPkgs(testPkgsDir)
	if err != nil {
		t.Fatalf("Failed error:%v", err)
	}

	// Get test pkg
	req := httptest.NewRequest("GET", "/pkgs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	var pkgInfos []pkg.PackageInfo
	json.Unmarshal(resbody, &pkgInfos)

	assert.Equal(t, pkgInfos[0].MetaInfo.PkgID, pkgInfo.MetaInfo.PkgID, "unmatched ID")

	// Start test pkg
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/pkg/"+pkgInfo.MetaInfo.PkgID.String()+"/start", nil)
	router.ServeHTTP(w, req)

	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	var startedApp app.AppInfo
	json.Unmarshal(resbody, &startedApp)

	// Get started app and check
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/apps", nil)

	router.ServeHTTP(w, req)
	resp = w.Result()
	resbody, _ = ioutil.ReadAll(resp.Body)
	var appInfos []app.AppInfo
	json.Unmarshal(resbody, &appInfos)

	assert.Equal(t, startedApp.Id, appInfos[0].Id)

	exec.Command("rm", "-rf", testPkgsDir).Run()
}
