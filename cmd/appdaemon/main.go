package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/vishvananda/netlink"
)

var childCmd *exec.Cmd

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) > 1 && args[0] == "testMode" {
		testMode(args[1])
	}

	defaultRoute, err := getDefaultRoute()
	if err != nil {
		fmt.Println("Default route not found")
	} else {
		fmt.Printf("Default route: %v\n", defaultRoute.String())
		appID, err := uuid.Parse(args[1])
		if err != nil {
			fmt.Println(err)
		} else {
			url := "http://" + defaultRoute.String() + ":8080"
			appInfo, err := getAppInfo(appID, url)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(appInfo)
				configureNetwork()
			}
		}
	}

	pkgInfo, err := pkg.OpenPackageInfo(args[0])
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting appdaemon...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go signalHandler(signalChan)

	start(pkgInfo)
	exitCode := wait()
	fmt.Printf("Exiting appdaemon %v\n", exitCode)
	os.Exit(exitCode)
}

func start(pkgInfo *pkg.PackageInfo) error {

	cmds := pkgInfo.MetaInfo.CMD
	childCmd = exec.Command(cmds[0], cmds[1:]...)

	err := childCmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func stop() int {
	syscall.Kill(childCmd.Process.Pid, syscall.SIGTERM)
	return wait()
}

func wait() int {
	childCmd.Wait()
	return childCmd.ProcessState.ExitCode()
}

func signalHandler(signalChan chan os.Signal) {
	s := <-signalChan
	switch s {
	case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		syscall.Kill(childCmd.Process.Pid, syscall.SIGTERM)
	}
}

func testMode(mode string) {
	exec.Command("rm", "/tmp/appdaemon_test").Run()
	pkgInfo := pkg.CreateSkeltonPackageInfo()
	if mode == "noSIGTERM" {
		pkgInfo.MetaInfo.CMD = []string{"echo", "hello"}
	} else if mode == "SIGTERM" {
		pkgInfo.MetaInfo.CMD = []string{"sleep", "10"}
	} else {
		fmt.Printf("invalid argument %v\n", mode)
	}

	start(pkgInfo)
	exitCode := wait()
	file, err := os.Create("/tmp/appdaemon_test")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	file.WriteString(strconv.Itoa(exitCode) + "\n")
	file.Close()
	os.Exit(exitCode)
}

func configureNetwork() error {
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	dgwLinkIndex, err := getDefaultRouteLinkIndex()
	if err != nil {
		return err
	}

	var dgwLink netlink.Link = nil
	var childLink netlink.Link = nil

	for idx := range links {
		link := links[idx]
		switch link.(type) {
		case *netlink.Veth:
			if link.Attrs().Index == dgwLinkIndex {
				dgwLink = links[idx]
			} else {
				childLink = links[idx]
			}
		default:
			fmt.Println("UNKNOWN")
		}
	}

	err = configureNAT(dgwLink, childLink)
	if err != nil {
		return err
	}

	return nil
}

func getAppInfo(appID uuid.UUID, url string) (*app.AppInfo, error) {
	urlApp := url + "/app/" + appID.String()
	fmt.Println(urlApp)
	resp, err := http.Get(urlApp)
	if err != nil {
		return nil, err
	}
	fmt.Println("HOGE" + urlApp)

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	appInfo := app.AppInfo{}
	err = json.Unmarshal(respByte, &appInfo)
	if err != nil {
		return nil, err
	}

	if appID != appInfo.Id {
		return nil, fmt.Errorf("invalid appID")
	}

	return &appInfo, nil
}

func getDefaultRoute() (net.IP, error) {
	routeList, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	for idx := range routeList {
		route := routeList[idx]
		if route.Scope == netlink.SCOPE_UNIVERSE {
			return route.Gw, nil
		}
	}

	return nil, fmt.Errorf("default route not found")
}

func getDefaultRouteLinkIndex() (int, error) {
	routeList, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return 0, err
	}

	for idx := range routeList {
		route := routeList[idx]
		if route.Scope == netlink.SCOPE_UNIVERSE {
			return route.LinkIndex, nil
		}
	}

	return 0, fmt.Errorf("default route not found")
}
