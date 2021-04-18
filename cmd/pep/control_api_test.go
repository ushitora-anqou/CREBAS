package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

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

func TestStartAppFromPkg(t *testing.T) {
	testPkgsDir := "/tmp/pep_test"
	exec.Command("mkdir", "-p", testPkgsDir).Run()
	defer exec.Command("rm", "-rf", testPkgsDir).Run()
	defer pkgs.Clear()
	defer apps.Clear()

	startOFController()
	defer controller.Stop()

	err := prepareNetwork()
	if err != nil {
		panic(err)
	}
	defer clearNetwork()

	pkgInfo := pkg.CreateSkeltonPackageInfo()
	pkgInfo.MetaInfo.CMD = []string{"/bin/bash", "-c", "sleep 3"}
	err = pkg.CreatePackage(pkgInfo, testPkgsDir)
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

	// Check network
	time.Sleep(1 * time.Second)
	cmd := exec.Command("ping", "192.168.10.2", "-c", "1")
	out, err := cmd.Output()
	fmt.Println(string(out))
	if err != nil {
		t.Fatalf("Failed %v %v", err, string(out))
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		t.Fatalf("Failed exit code:%v", exitCode)
	}

	// Check process exit
	time.Sleep(4 * time.Second)
	apps.ClearNotRunningApp()

	assert.Equal(t, 0, apps.Count(), "app is not cleared")
}
