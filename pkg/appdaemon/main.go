package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/naoki9911/CREBAS/pkg/pkg"
)

var childCmd *exec.Cmd
var childExitCode int

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go signalHandler(signalChan)

	flag.Parse()
	args := flag.Args()
	if len(args) > 1 && args[0] == "testMode" {
		testMode(args[1])
	}
	exec.Command("/usr/bin/ls", "-l", "/tmp/apppackager").Run()
	pkgInfo, err := pkg.OpenPackageInfo(args[0])
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting appdaemon...")
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
