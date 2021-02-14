package app

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// LinuxProcess is a application running on linux process
type LinuxProcess struct {
	id        uuid.UUID
	pid       int
	links     netlinkext.LinkCollection
	namespace string
	cmd       []string
}

// NewLinuxProcess creates linux process application
func NewLinuxProcess() *LinuxProcess {
	proc := new(LinuxProcess)
	proc.id, _ = uuid.NewRandom()

	return proc
}

// Create creates process
func (p *LinuxProcess) Create() error {
	uuidStr := p.id.String()[0:8]

	p.namespace = "netns-" + uuidStr
	handle, err := netlinkext.CreateNetns(p.namespace)
	if err != nil {
		return err
	}
	defer handle.Close()

	return nil
}

// Delete deletes process
func (p *LinuxProcess) Delete() error {
	return netns.DeleteNamed(p.namespace)
}

// Start process
func (p *LinuxProcess) Start() error {
	return p.execCmdWithNetns()
}

// Stop process
func (p *LinuxProcess) Stop() error {
	proc, err := os.FindProcess(p.pid)
	if err != nil {
		return err
	}
	err = syscall.Kill(-p.pid, syscall.SIGKILL)
	if err != nil {
		return err
	}
	_, err = proc.Wait()
	if err != nil {
		return nil
	}
	return nil
}

func (p *LinuxProcess) addLink(name string) error {

	return nil
}

func (p *LinuxProcess) execCmdWithNetns() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rootNetns, err := netns.Get()
	if err != nil {
		return err
	}
	defer rootNetns.Close()

	handle, err := netns.GetFromName(p.namespace)
	if err != nil {
		return err
	}
	defer handle.Close()

	err = netns.Set(handle)
	if err != nil {
		return err
	}

	//EXEC
	childPID, err := syscall.ForkExec(p.cmd[0], p.cmd,
		&syscall.ProcAttr{
			Env: os.Environ(),
			Sys: &syscall.SysProcAttr{
				Setsid: true,
			},
			Files: []uintptr{0, 1, 2}, // print message to the same pty
		})

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	p.pid = childPID

	err = netns.Set(rootNetns)
	if err != nil {
		return err
	}

	return nil
}

func createVethPeer(linkName string, peerName string) error {

	vethLink := &netlink.Veth{}
	vethLink.LinkAttrs = netlink.NewLinkAttrs()
	vethLink.Name = linkName
	vethLink.PeerName = peerName

	return netlink.LinkAdd(vethLink)
}
