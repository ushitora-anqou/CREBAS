package app

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// LinuxProcess is a application running on linux process
type LinuxProcess struct {
	pkgInfo      *pkg.PackageInfo
	id           uuid.UUID
	pid          int
	defaultRoute net.IP
	links        *netlinkext.LinkCollection
	namespace    string
	cmd          []string
	exitCode     int
	exitChan     chan bool
	device       *Device
	capabilities *capability.CapabilityCollection

	DhcpConfigPath string
	ACLLink        *netlinkext.LinkExt
}

// NewLinuxProcess creates linux process application
func NewLinuxProcess() (*LinuxProcess, error) {
	proc := new(LinuxProcess)
	proc.id, _ = uuid.NewRandom()
	proc.pid = -1
	proc.exitCode = -1
	proc.exitChan = make(chan bool, 1)
	proc.links = netlinkext.NewLinkCollection()
	proc.capabilities = capability.NewCapabilityCollection()

	uuidStr := proc.id.String()[0:8]
	proc.namespace = "netns-" + uuidStr
	handle, err := netlinkext.CreateNetns(proc.namespace)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	return proc, nil
}

func NewLinuxProcessFromPkgInfo(pkgInfo *pkg.PackageInfo) (*LinuxProcess, error) {
	proc, err := NewLinuxProcess()
	proc.pkgInfo = pkgInfo
	proc.cmd = pkgInfo.MetaInfo.CMD

	return proc, err
}

// Start process
func (p *LinuxProcess) Start() error {
	return p.execCmdWithNetns()
}

// Stop process
func (p *LinuxProcess) Stop() error {
	if p.pid > 0 {
		err := p.killProc()
		if err != nil {
			return err
		}
	}

	err := p.delete()
	if err != nil {
		return err
	}

	return nil
}

func (p *LinuxProcess) IsRunning() bool {
	if p.pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(p.pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func (p *LinuxProcess) killProc() error {
	if !p.IsRunning() {
		return nil
	}

	err := syscall.Kill(p.pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	exit := <-p.exitChan
	fmt.Printf("EXIT %v", exit)
	return nil
}

// Delete deletes process
func (p *LinuxProcess) delete() error {
	links := p.links.Where(func(link *netlinkext.LinkExt) bool { return true })
	for _, link := range links {
		err := link.Delete()
		if err != nil {
			return err
		}
	}

	if p.pkgInfo != nil && p.pkgInfo.UnpackedPkgPath != "" {
		exec.Command("rm", "-rf", p.pkgInfo.UnpackedPkgPath).Run()
	}

	return netns.DeleteNamed(p.namespace)
}

// ID returns app's id
func (p *LinuxProcess) ID() uuid.UUID {
	return p.id
}

// GetAppInfo returns appInfo
func (p *LinuxProcess) GetAppInfo() *AppInfo {
	appInfo := AppInfo{
		Id: p.id,
	}

	return &appInfo
}

// AddLink adds link
func (p *LinuxProcess) AddLink(ofs *ofswitch.OFSwitch, ofType netlinkext.OFType) (*netlinkext.LinkExt, error) {
	linkUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	linkName := linkUUID.String()[0:8]
	peerName := linkName + "-p"

	link := netlinkext.NewLinkExtVeth(linkName, peerName)
	err = link.Create()
	if err != nil {
		return nil, err
	}

	err = link.SetNsByName(p.namespace)
	if err != nil {
		return nil, err
	}

	err = ofs.AttachLink(link, ofType)
	if err != nil {
		return nil, err
	}

	err = link.SetLinkPeerUp()
	if err != nil {
		return nil, err
	}

	err = link.SetLinkUp()
	if err != nil {
		return nil, err
	}

	p.links.Add(link)

	return link, nil
}

func (p *LinuxProcess) AddLinkWithAddr(ofs *ofswitch.OFSwitch, ofType netlinkext.OFType, addr *netlink.Addr) (*netlinkext.LinkExt, error) {
	link, err := p.AddLink(ofs, ofType)
	if err != nil {
		return nil, err
	}

	err = link.SetAddr(addr)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (p *LinuxProcess) AddLinkWithReservedAddr(ofs *ofswitch.OFSwitch, ofType netlinkext.OFType, addr *netlink.Addr) (*netlinkext.LinkExt, error) {
	link, err := p.AddLink(ofs, ofType)
	if err != nil {
		return nil, err
	}

	link.Addr = addr

	return link, nil
}

func (p *LinuxProcess) SetDefaultRoute(addr net.IP) error {
	err := exec.Command("ip", "netns", "exec", p.namespace, "ip", "route", "add", "default", "via", addr.String()).Run()
	if err != nil {
		return err
	}

	p.defaultRoute = addr
	return nil
}

func (p *LinuxProcess) GetDefaultRoute() net.IP {
	return p.defaultRoute
}

func (p *LinuxProcess) GetExitCode() int {
	return p.exitCode
}

func (p *LinuxProcess) SetDevice(device *Device) error {
	p.device = device
	return nil
}

func (p *LinuxProcess) GetDevice() *Device {
	return p.device
}

func (p *LinuxProcess) Capabilities() *capability.CapabilityCollection {
	return p.capabilities
}

func (p *LinuxProcess) Links() *netlinkext.LinkCollection {
	return p.links
}

func (p *LinuxProcess) NameSpace() string {
	return p.namespace
}

func (p *LinuxProcess) SetDNSServer(addr net.IP) error {
	netnsPath := "/etc/netns/" + p.namespace
	cmd := exec.Command("mkdir", "-p", netnsPath)
	if err := cmd.Run(); err != nil {
		log.Printf("error: Failed to create directory %v", netnsPath)
		return err
	}
	fmt.Println(netnsPath)
	file, err := os.Create(netnsPath + "/resolv.conf")
	if err != nil {
		log.Printf("error: Failed to create %v", netnsPath)
		return err
	}
	defer file.Close()
	_, err = file.WriteString("nameserver " + addr.String() + "\n")
	if err != nil {
		log.Printf("error: Failed to write to %v", netnsPath+"/resolv.conf")
		return err
	}

	return nil
}

func (p *LinuxProcess) SetDHCPClientIdentifier(clientIdentifier string) error {
	netnsPath := "/etc/netns/" + p.namespace
	cmd := exec.Command("mkdir", "-p", netnsPath)
	if err := cmd.Run(); err != nil {
		log.Printf("error: Failed to create directory %v", netnsPath)
		return err
	}
	configPath := netnsPath + "/dhclient.conf"
	file, err := os.Create(configPath)
	if err != nil {
		log.Printf("error: Failed to create %v", configPath)
		return err
	}
	defer file.Close()
	p.DhcpConfigPath = configPath

	file.WriteString("option rfc3442-classless-static-routes code 121 = array of unsigned integer 8;\n")
	file.WriteString("send host-name = " + clientIdentifier + ";\n")
	file.WriteString("request subnet-mask, broadcast-address, time-offset, routers,\n")
	file.WriteString("domain-name, domain-name-servers, domain-search, host-name,\n")
	file.WriteString("dhcp6.name-servers, dhcp6.domain-search, dhcp6.fqdn, dhcp6.sntp-servers,\n")
	file.WriteString("netbios-name-servers, netbios-scope, interface-mtu,\n")
	file.WriteString("rfc3442-classless-static-routes, ntp-servers;\n")
	file.WriteString("send dhcp-client-identifier = 1:0:a0:24:ab:fb:9c;\n")
	file.WriteString("timeout 300;\n")

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
	cmd := p.cmd
	if p.pkgInfo != nil && p.pkgInfo.UnpackedPkgPath != "" {
		cmd = []string{"/tmp/appdaemon", filepath.Join(p.pkgInfo.UnpackedPkgPath, "pkgInfo.json"), p.id.String()}
	}

	netnsCmd := []string{"netns", "exec", p.namespace}
	netnsCmd = append(netnsCmd, cmd...)
	cmdExec := exec.Command("ip", netnsCmd...)

	stdout, _ := cmdExec.StdoutPipe()

	buff := make([]byte, 1024)

	err = cmdExec.Start()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	p.pid = cmdExec.Process.Pid

	err = netns.Set(rootNetns)
	if err != nil {
		return err
	}

	go func() {
		proc, err := os.FindProcess(p.pid)
		if err != nil {
			return
		}
		procState, err := proc.Wait()
		if err != nil {
			return
		}
		p.exitCode = procState.ExitCode()
		p.exitChan <- true
	}()

	go func() {
		n, err := stdout.Read(buff)

		for err == nil || err != io.EOF {
			if n > 0 {
				fmt.Print(string(buff[:n]))
			}

			n, err = stdout.Read(buff)
		}
	}()

	return nil
}

func createVethPeer(linkName string, peerName string) error {

	vethLink := &netlink.Veth{}
	vethLink.LinkAttrs = netlink.NewLinkAttrs()
	vethLink.Name = linkName
	vethLink.PeerName = peerName

	return netlink.LinkAdd(vethLink)
}
