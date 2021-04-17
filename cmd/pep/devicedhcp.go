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
	"github.com/insomniacslk/dhcp/dhcpv6"

	"github.com/coredhcp/coredhcp/plugins"
	pl_leasetime "github.com/coredhcp/coredhcp/plugins/leasetime"

	"github.com/sirupsen/logrus"
)

var (
	flagLogFile     = flag.String("logfile", "", "Name of the log file to append to. Default: stdout/stderr only")
	flagLogNoStdout = flag.Bool("nostdout", false, "Disable logging to stdout/stderr")
	flagLogLevel    = flag.String("loglevel", "info", fmt.Sprintf("Log level. One of %v", getLogLevels()))
	flagConfig      = flag.String("conf", "", "Use this configuration file instead of the default location")
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
		Zone: "crebas-ext-ofs",
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
	srv, err := server.Start(conf)
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Wait(); err != nil {
		log.Print(err)
	}
	time.Sleep(time.Second)
}

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name:   "externaldhcp",
	Setup6: setup6,
	Setup4: setup4,
}

// Handler6 TODO: Implement IPv6
func handler6(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
	return resp, false
}

// Handler4 handles DHCPv4 packets for the fcvm plugin.
func handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	fmt.Println("HANDLE")
	log := logger.GetLogger("dhcpserver")
	routerIP := "192.168.20.1/24"
	ovsIP, ovsSubnet, err := net.ParseCIDR(routerIP)
	if err != nil {
		log.Errorf("Failed to parse : %v", routerIP)
		return resp, true
	}

	// Update DefaultRoute, SubnetMask
	resp.Options.Update(dhcpv4.OptRouter([]net.IP{ovsIP}...))
	resp.Options.Update(dhcpv4.OptSubnetMask(ovsSubnet.Mask))

	// Update ServerIdentifier
	resp.ServerIPAddr = make(net.IP, net.IPv4len)
	copy(resp.ServerIPAddr[:], ovsIP)
	resp.UpdateOption(dhcpv4.OptServerIdentifier(ovsIP))

	// Update DNS information
	dnsIP := net.ParseIP("192.168.10.1")
	if req.IsOptionRequested(dhcpv4.OptionDomainNameServer) {
		resp.Options.Update(dhcpv4.OptDNS([]net.IP{dnsIP}...))
	}

	//clientIdentifierBytes := req.Options.Get(dhcpv4.OptionClientIdentifier)
	//if clientIdentifierBytes == nil {
	//	return resp, true
	//}
	//clientIdentifier := string(clientIdentifierBytes)
	//log.Infof("ClientIdentifier :%v", clientIdentifier)

	deviceIP := net.ParseIP("192.168.20.2")
	resp.YourIPAddr = deviceIP
	log.Printf("found IP address %s for MAC %s", resp.YourIPAddr, req.ClientHWAddr.String())
	return resp, false
}

func setup4(args ...string) (handler.Handler4, error) {
	return handler4, nil
}

func setup6(args ...string) (handler.Handler6, error) {
	return handler6, nil
}
