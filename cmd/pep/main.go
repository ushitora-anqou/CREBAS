package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/naoki9911/gofc"
	"github.com/vishvananda/netlink"
)

var apps = app.AppCollection{}
var pkgs = pkg.PkgCollection{}
var devices = app.DeviceCollection{}
var aclOfs = &ofswitch.OFSwitch{}
var extOfs = &ofswitch.OFSwitch{}
var appAddrPool = &ofswitch.IP4AddrPool{}
var extAddrPool = &ofswitch.IP4AddrPool{}
var controller = gofc.NewOFController()
var dnsServer = "8.8.8.8:53"
var pepConfig = NewConfig()
var pepID uuid.UUID
var certificate *x509.Certificate
var privateKey *rsa.PrivateKey

var cpCert *capability.AppCertificate
var userCert *capability.AppCertificate

const cpUrl = "http://localhost:8081"

func main() {
	configureCredentials()
	startOFController()
	err := prepareNetwork()
	if err != nil {
		panic(err)
	}
	defer clearNetwork()
	err = setupWiFi()
	if err != nil {
		panic(err)
	}
	err = prepareTestPkg()
	if err != nil {
		panic(err)
	}
	go startDNSServer(aclOfs)
	go StartDHCPServer()
	StartAPIServer()
}

func configureCredentials() {
	pepID, _ = uuid.NewRandom()

	certBytes, err := capability.ReadCertificateWithoutDecode("/home/naoki/CREBAS/test/keys/pep/test-pep.crt")
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	certificate, err = capability.DecodeCertificate(certBytes)
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	privateKey, err = capability.ReadPrivateKey("/home/naoki/CREBAS/test/keys/pep/test-pep.key")
	if err != nil {
		fmt.Printf("Failed %v\n", err)
		panic(err)
	}

	certBase64 := base64.StdEncoding.EncodeToString(certBytes)
	appCert := capability.AppCertificate{
		AppID:             pepID,
		CertificateString: certBase64,
	}
	_, err = capability.SendContentsToCP(cpUrl+"/app/cert", appCert)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	cpCert, err = getCertificate(cpUrl + "/app/cpCert")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	userCert, err = getCertificate(cpUrl + "/app/userCert")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func getCertificate(url string) (*capability.AppCertificate, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	appCert := capability.AppCertificate{}
	err = json.Unmarshal(respByte, &appCert)
	if err != nil {
		return nil, err
	}

	err = appCert.Decode()
	if err != nil {
		return nil, err
	}

	return &appCert, nil
}

func startOFController() {
	log.Printf("Starting OpenFlow Controller...")
	go controller.ServerLoop(gofc.DEFAULT_PORT)
	log.Printf("Started OpenFlow Controller")
}

func appendOFSwitchToController(c *ofswitch.OFSwitch) {
	gofc.GetAppManager().RegistApplication(c)
}

func waitOFSwitchConnectedToController(c *ofswitch.OFSwitch) {
	for {
		if c.IsConnectedToController() {
			fmt.Println(c.Name + " is Connected!")
			break
		}
		time.Sleep(1 * time.Second)
	}

}

func prepareNetwork() error {
	aclOfs = ofswitch.NewOFSwitch(pepConfig.aclOfsName)
	aclOfs.Delete()
	err := aclOfs.Create()
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(pepConfig.aclOfsAddr)
	if err != nil {
		return err
	}
	err = aclOfs.SetAddr(addr)
	if err != nil {
		return err
	}
	appAddrPool = ofswitch.NewIP4AddrPool(addr)
	err = appAddrPool.LeaseWithAddr(addr)
	if err != nil {
		return err
	}

	extOfs = ofswitch.NewOFSwitch(pepConfig.extOfsName)
	extOfs.Delete()
	err = extOfs.Create()
	if err != nil {
		return err
	}

	appendOFSwitchToController(extOfs)

	addr, err = netlink.ParseAddr(pepConfig.extOfsAddr)
	if err != nil {
		return err
	}
	err = extOfs.SetAddr(addr)
	if err != nil {
		return err
	}
	extAddrPool = ofswitch.NewIP4AddrPool(addr)
	err = extAddrPool.LeaseWithAddr(addr)
	if err != nil {
		return err
	}

	extAppAddr, err := netlink.ParseAddr(pepConfig.extOfsAppAddr)
	if err != nil {
		return err
	}
	err = extAddrPool.LeaseWithAddr(extAppAddr)
	if err != nil {
		return err
	}

	err = extOfs.SetController("tcp:127.0.0.1:6653")
	if err != nil {
		return err
	}

	waitOFSwitchConnectedToController(extOfs)

	return nil
}

func clearNetwork() error {
	err := aclOfs.Delete()
	if err != nil {
		return err
	}
	err = extOfs.Delete()
	if err != nil {
		return err
	}

	return nil
}

func prepareTestPkg() error {
	pkgDir := "/tmp/pep_test"

	pkg1 := pkg.CreateSkeltonPackageInfo()
	pkg1.MetaInfo.CMD = []string{"/bin/bash", "-c", "while true; do sleep 1; done"}
	pkg1.Server = false
	capReqND1 := capability.NewCreateSkeltonCapabilityRequest()
	capReqND1.RequestCapabilityName = capability.CAPABILITY_NAME_NEIGHBOR_DISCOVERY
	capReqND2 := capability.NewCreateSkeltonCapabilityRequest()
	capReqND2.RequestCapabilityName = capability.CAPABILITY_NAME_TEMPERATURE
	capReqND3 := capability.NewCreateSkeltonCapabilityRequest()
	capReqND3.RequestCapabilityName = capability.CAPABILITY_NAME_HUMIDITY
	pkg1.CapabilityRequests = append(pkg1.CapabilityRequests, capReqND1)
	pkg1.CapabilityRequests = append(pkg1.CapabilityRequests, capReqND2)
	pkg1.CapabilityRequests = append(pkg1.CapabilityRequests, capReqND3)
	pkg1.PrivateKeyPath = "/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.key"
	pkg1.CertificatePath = "/home/naoki/CREBAS/test/keys/virt-dev-1/test-virt-dev-1.crt"
	proc1, err := app.NewLinuxProcessFromPkgInfo(pkg1)
	if err != nil {
		return err
	}
	err = pkg.CreateUnpackedPackage(pkg1, pkgDir)
	if err != nil {
		return err
	}

	deviceIP, err := extAddrPool.Lease()
	if err != nil {
		return err
	}
	hwAddr, err := net.ParseMAC("58:cb:52:56:73:21")
	if err != nil {
		return err
	}
	device := &app.Device{
		HWAddress: hwAddr,
		IPAddress: deviceIP,
		App:       proc1,
		OfPort:    pepConfig.wifiLink.GetOfPort(),
		ViaWlan:   true,
	}
	proc1.SetDevice(device)

	devices.Add(device)

	pkg2 := pkg.CreateSkeltonPackageInfo()
	pkg2.MetaInfo.CMD = []string{"/bin/bash", "-c", "while true; do sleep 1; done"}
	pkg2.Server = true
	capND := capability.NewCreateSkeltonCapability()
	capND.CapabilityName = capability.CAPABILITY_NAME_NEIGHBOR_DISCOVERY
	capND.CapabilityValue = "8000/udp"
	capND.GrantCondition = "always"
	capTemp := capability.NewCreateSkeltonCapability()
	capTemp.CapabilityName = capability.CAPABILITY_NAME_TEMPERATURE
	capTemp.CapabilityValue = "8000/udp"
	capHumid := capability.NewCreateSkeltonCapability()
	capHumid.CapabilityName = capability.CAPABILITY_NAME_HUMIDITY
	capHumid.CapabilityValue = "8000/udp"

	pkg2.Capabilities = append(pkg2.Capabilities, capND)
	pkg2.Capabilities = append(pkg2.Capabilities, capTemp)
	pkg2.Capabilities = append(pkg2.Capabilities, capHumid)
	pkg1.PrivateKeyPath = "/home/naoki/CREBAS/test/keys/virt-dev-2/test-virt-dev-2.key"
	pkg1.CertificatePath = "/home/naoki/CREBAS/test/keys/virt-dev-2/test-virt-dev-2.crt"

	proc2, err := app.NewLinuxProcessFromPkgInfo(pkg2)
	if err != nil {
		return err
	}
	err = pkg.CreateUnpackedPackage(pkg2, pkgDir)
	if err != nil {
		return err
	}

	device2IP, err := extAddrPool.Lease()
	if err != nil {
		return err
	}
	hwAddr2, err := net.ParseMAC("80:7d:3a:c8:2b:5c")
	if err != nil {
		return err
	}
	device2 := &app.Device{
		HWAddress: hwAddr2,
		IPAddress: device2IP,
		App:       proc2,
		OfPort:    pepConfig.wifiLink.GetOfPort(),
		ViaWlan:   true,
	}
	proc2.SetDevice(device2)

	devices.Add(device2)

	return nil
}

func setupWiFi() error {
	wifiLinkName := "wlp4s0"
	link, err := netlink.LinkByName(wifiLinkName)
	if err != nil {
		return err
	}
	linkExt := &netlinkext.LinkExt{}
	linkExt.SetLink(link)

	for {
		ofport, err := ofswitch.GetOFPortByLinkName(wifiLinkName)
		if err == nil {
			log.Printf("Found OfPort %v %v", wifiLinkName, ofport)
			linkExt.Ofport = ofport
			break
		}
		log.Printf("Not Found OfPort for %v", wifiLinkName)
		time.Sleep(1 * time.Second)
	}
	err = extOfs.ResetController()
	if err != nil {
		return err
	}
	log.Printf("Reset Controller")
	waitOFSwitchConnectedToController(extOfs)

	pepConfig.wifiLink = linkExt

	err = extOfs.AddHostEAPoLFlow(linkExt)
	if err != nil {
		return err
	}
	err = extOfs.AddHostAggregatedARPFlow(linkExt)
	if err != nil {
		return err
	}
	err = extOfs.AddHostAggregatedDHCPFlow(linkExt)
	if err != nil {
		return err
	}
	return nil
}
