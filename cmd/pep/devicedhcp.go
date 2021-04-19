package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/coredhcp/coredhcp/config"
	"github.com/coredhcp/coredhcp/handler"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/server"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/vishvananda/netlink"

	"github.com/coredhcp/coredhcp/plugins"
	pl_leasetime "github.com/coredhcp/coredhcp/plugins/leasetime"

	"github.com/sirupsen/logrus"
)

var (
	flagLogFile     = flag.String("logfile", "", "Name of the log file to append to. Default: stdout/stderr only")
	flagLogNoStdout = flag.Bool("nostdout", false, "Disable logging to stdout/stderr")
	flagLogLevel    = flag.String("loglevel", "info", fmt.Sprintf("Log level. One of %v", getLogLevels()))
	flagPlugins     = flag.Bool("plugins", false, "list plugins")
)

var logLevels = map[string]func(*logrus.Logger){
	"none":    func(l *logrus.Logger) { l.SetOutput(ioutil.Discard) },
	"debug":   func(l *logrus.Logger) { l.SetLevel(logrus.DebugLevel) },
	"info":    func(l *logrus.Logger) { l.SetLevel(logrus.InfoLevel) },
	"warning": func(l *logrus.Logger) { l.SetLevel(logrus.WarnLevel) },
	"error":   func(l *logrus.Logger) { l.SetLevel(logrus.ErrorLevel) },
	"fatal":   func(l *logrus.Logger) { l.SetLevel(logrus.FatalLevel) },
}

func getLogLevels() []string {
	var levels []string
	for k := range logLevels {
		levels = append(levels, k)
	}
	return levels
}

var desiredPlugins = []*plugins.Plugin{
	&pl_leasetime.Plugin,
	&Plugin,
}

var srv *server.Servers

func StartDHCPServer() {
	flag.Parse()

	if *flagPlugins {
		for _, p := range desiredPlugins {
			fmt.Println(p.Name)
		}
		os.Exit(0)
	}

	log := logger.GetLogger("dhcpserver")
	fn, ok := logLevels[*flagLogLevel]
	if !ok {
		log.Fatalf("Invalid log level '%s'. Valid log levels are %v", *flagLogLevel, getLogLevels())
	}
	fn(log.Logger)
	log.Infof("Setting log level to '%s'", *flagLogLevel)
	if *flagLogFile != "" {
		log.Infof("Logging to file %s", *flagLogFile)
		logger.WithFile(log, *flagLogFile)
	}
	if *flagLogNoStdout {
		log.Infof("Disabling logging to stdout/stderr")
		logger.WithNoStdOutErr(log)
	}
	// register plugins
	for _, plugin := range desiredPlugins {
		if err := plugins.RegisterPlugin(plugin); err != nil {
			log.Fatalf("Failed to register plugin '%s': %v", plugin.Name, err)
		}
	}
	conf := config.New()
	server4Config := config.ServerConfig{}

	listener := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: dhcpv4.ServerPort,
		Zone: pepConfig.extOfsName,
	}

	server4Config.Addresses = []net.UDPAddr{listener}
	server4Config.Plugins = []config.PluginConfig{
		{
			Name: "lease_time",
			Args: []string{"60s"},
		},
		{
			Name: "externaldhcp",
			Args: []string{"60s"},
		},
	}
	conf.Server4 = &server4Config

	// start server
	var err error
	srv, err = server.Start(conf)
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Wait(); err != nil {
		log.Print(err)
	}
	time.Sleep(time.Second)
}

var Plugin = plugins.Plugin{
	Name:   "externaldhcp",
	Setup6: nil,
	Setup4: setup4,
}

func handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	fmt.Println("HANDLE")
	log := logger.GetLogger("dhcpserver")
	ovsIP, ovsSubnet, err := net.ParseCIDR(pepConfig.extOfsAddr)
	if err != nil {
		log.Errorf("Failed to parse : %v", pepConfig.extOfsAddr)
		return resp, true
	}

	resp.Options.Update(dhcpv4.OptRouter([]net.IP{ovsIP}...))
	resp.Options.Update(dhcpv4.OptSubnetMask(ovsSubnet.Mask))

	resp.ServerIPAddr = make(net.IP, net.IPv4len)
	copy(resp.ServerIPAddr[:], ovsIP)
	resp.UpdateOption(dhcpv4.OptServerIdentifier(ovsIP))

	aclIP, _, err := net.ParseCIDR(pepConfig.aclOfsAddr)
	if err != nil {
		log.Errorf("Failed to parse : %v", pepConfig.aclOfsAddr)
	}
	if req.IsOptionRequested(dhcpv4.OptionDomainNameServer) {
		resp.Options.Update(dhcpv4.OptDNS([]net.IP{aclIP}...))
	}

	clientIdentifierBytes := req.Options.Get(dhcpv4.OptionClientIdentifier)
	if clientIdentifierBytes != nil {
		log.Infof("ClientIdentifier :%v", string(clientIdentifierBytes))
	}

	selectedDevices := devices.Where(func(d *app.Device) bool {
		return d.HWAddress.String() == req.ClientHWAddr.String()
	})

	if len(selectedDevices) == 0 {
		deviceIP, err := extAddrPool.Lease()
		if err != nil {
			log.Infof("Failed to Lease Addr for %v", req.ClientHWAddr.String())
			return resp, true
		}
		device := app.Device{
			HWAddress: req.ClientHWAddr,
			IPAddress: deviceIP,
		}
		devices.Add(&device)

		resp.YourIPAddr = deviceIP.IP
		log.Infof("Assigned IP %v for %v", deviceIP.IP.String(), req.ClientHWAddr.String())
	} else {
		device := selectedDevices[0]
		resp.YourIPAddr = device.IPAddress.IP
		log.Infof("found IP address %s for MAC %s", resp.YourIPAddr, req.ClientHWAddr.String())

		if device.App != nil {
			if !device.App.IsRunning() {
				err = startAppWithDevice(device)
				if err != nil {
					log.Errorf("failed to start app %v", err)
					return resp, true
				}
				log.Infof("Starting corresponding app %v", device.App.ID())
			}
		}
	}

	return resp, false
}

func setup4(args ...string) (handler.Handler4, error) {
	return handler4, nil
}

func startAppWithDevice(device *app.Device) error {
	proc := device.App.(*app.LinuxProcess)
	procAddr, err := netlink.ParseAddr(pepConfig.extOfsAppAddr)
	if err != nil {
		return err
	}
	procLink, err := proc.AddLinkWithAddr(extOfs, netlinkext.ExternalOFSwitch, procAddr)
	if err != nil {
		return err
	}

	if device.GetViaWlan() {
		err = extOfs.AddDeviceAppARPFlow(device, procLink)
		if err != nil {
			return err
		}
		err = extOfs.AddDeviceAppIPFlow(device, procLink)
		if err != nil {
			return err
		}
	} else {
		err = extOfs.AddDeviceTunnelFlow(device, procLink)
		if err != nil {
			return err
		}

		err = extOfs.DeleteHostARPFlow(device)
		if err != nil {
			return err
		}
	}

	err = device.App.Start()
	if err != nil {
		return err
	}
	fmt.Println("APP NAMESPACE:" + proc.NameSpace())
	return nil
}
