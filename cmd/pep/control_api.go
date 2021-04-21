package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/pkg"
	"github.com/vishvananda/netlink"
)

func getAllApps(c *gin.Context) {
	c.JSON(http.StatusOK, apps.GetAllAppInfos())
}

func getAllPkgs(c *gin.Context) {
	c.JSON(http.StatusOK, pkgs.GetAll())
}

func startAppFromPkg(c *gin.Context) {
	id := c.Param("id")
	pkgID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error: invalid id %v", id)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	selectedPkgs := pkgs.Where(func(a *pkg.PackageInfo) bool {
		return a.MetaInfo.PkgID == pkgID
	})

	if len(selectedPkgs) != 1 {
		log.Printf("error: Unknown package ID %v", pkgID.String())
		c.JSON(http.StatusNotFound, nil)
		return
	}

	startPkg := selectedPkgs[0]
	startPkg, err = pkg.UnpackPkg(startPkg.PkgPath)
	if err != nil {
		log.Printf("error: Failed to unpack pkg %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	proc, err := app.NewLinuxProcessFromPkgInfo(startPkg)
	if err != nil {
		log.Printf("error: Failed to create process %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	procAddr, err := appAddrPool.Lease()
	if err != nil {
		log.Printf("error: Failed to lease ip address %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	link, err := proc.AddLinkWithAddr(aclOfs, netlinkext.ACLOFSwitch, procAddr)
	if err != nil {
		log.Printf("error: Failed to add link %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err = aclOfs.AddHostRestrictedFlow(link)
	if err != nil {
		log.Printf("error: Failed to add flow %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	extAddr, err := netlink.ParseAddr("192.168.20.1/24")
	if err != nil {
		log.Printf("error: Failed to parse %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	_, err = proc.AddLinkWithAddr(extOfs, netlinkext.ExternalOFSwitch, extAddr)
	if err != nil {
		log.Printf("error: Failed to parse %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	apps.Add(proc)

	c.JSON(http.StatusOK, proc.GetAppInfo())
}

func stopApp(c *gin.Context) {
	id := c.Param("id")
	appID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error: invalid id %v", id)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	selectedApp := apps.Where(func(a app.AppInterface) bool {
		return a.ID() == appID
	})

	if len(selectedApp) != 1 {
		log.Printf("error: invalid app ID %v", appID)
		c.JSON(http.StatusNotFound, nil)
		return
	}

	app := selectedApp[0]
	err = app.Stop()
	if err != nil {
		log.Printf("error: Failed to stop app(%v) %v", appID, err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err = apps.Remove(app)
	if err != nil {
		log.Printf("error: Failed to remove app(%v) %v", appID, err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, app.ID())
}

func getAppFromID(appID uuid.UUID) app.AppInterface {
	selectedApp := apps.Where(func(a app.AppInterface) bool {
		return a.ID() == appID
	})

	if len(selectedApp) != 1 {
		log.Printf("error: invalid app ID %v", appID)
		return nil
	}

	return selectedApp[0]
}

func getAppInfo(c *gin.Context) {
	id := c.Param("id")
	appID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error: invalid id %v", id)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	app := getAppFromID(appID)
	if app == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}
	c.JSON(http.StatusOK, app.GetAppInfo())
}

func setDevice(c *gin.Context) {
	id := c.Param("id")
	appID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error: invalid id %v", id)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var req app.Device
	app := getAppFromID(appID)
	if app == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}
	c.JSON(http.StatusOK, app.GetAppInfo())
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app.SetDevice(&req)
	c.JSON(http.StatusOK, req)
}

func getDevice(c *gin.Context) {
	id := c.Param("id")
	appID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error: invalid id %v", id)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	app := getAppFromID(appID)
	if app == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	fmt.Println(app.GetDevice())
	c.JSON(http.StatusOK, app.GetDevice())
}

func postAppCap(c *gin.Context) {
	id := c.Param("id")
	appID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("error: invalid id %v", id)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var req capability.Capability
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app := getAppFromID(appID)
	if app == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	exist := app.Capabilities().Where(func(cap *capability.Capability) bool {
		return cap.CapabilityID == req.CapabilityID
	})

	if len(exist) == 0 {
		app.Capabilities().Add(&req)
	}

	err = enforceCapability(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func StartAPIServer() error {
	return setupRouter().Run("0.0.0.0:8080")
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	r.GET("/pkgs", getAllPkgs)
	r.POST("/pkg/:id/start", startAppFromPkg)
	r.GET("/apps", getAllApps)
	r.GET("/app/:id", getAppInfo)
	r.POST("/app/:id/stop", stopApp)
	r.POST("/app/:id/device", setDevice)
	r.GET("/app/:id/device", getDevice)
	r.POST("/app/:id/cap", postAppCap)

	return r
}

func enforceCapability(cap *capability.Capability) error {
	log.Printf("info: Enforcing cap %v", cap)

	clientApp := getAppFromID(cap.AssigneeID)
	if clientApp == nil {
		log.Printf("error: clientApp %v not found", cap.AssigneeID)
		return fmt.Errorf("error: clientApp %v not found", cap.AssigneeID)
	}
	log.Printf("info: clientApp %v found", cap.AssigneeID)
	clientProc := clientApp.(*app.LinuxProcess)

	serverApp := getAppFromID(cap.AppID)
	if serverApp == nil {
		log.Printf("error: serverApp %v not found", cap.AppID)
		return fmt.Errorf("error: serverApp %v not found", cap.AppID)
	}
	log.Printf("info: serverApp %v found", cap.AppID)
	serverProc := serverApp.(*app.LinuxProcess)

	err := extOfs.AddAppsARPFlow(serverProc.GetDevice(), serverProc.ACLLink, clientProc.GetDevice(), clientProc.ACLLink)
	if err != nil {
		return err
	}
	err = extOfs.AddAppsICMPFlow(serverProc.GetDevice(), serverProc.ACLLink, clientProc.GetDevice(), clientProc.ACLLink)
	if err != nil {
		return err
	}

	if cap.CapabilityName == capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION {
		err = extOfs.AddAppsBroadcastUDPDstFlow(clientProc.GetDevice(), clientProc.ACLLink, serverProc.GetDevice(), serverProc.ACLLink, 8000)
		if err != nil {
			return err
		}
	}

	if cap.CapabilityName == capability.CAPABILITY_NAME_TEMPERATURE || cap.CapabilityName == capability.CAPABILITY_NAME_HUMIDITY {
		err = extOfs.AddAppsUnicastUDPDstFlow(clientProc.GetDevice(), clientProc.ACLLink, serverProc.GetDevice(), serverProc.ACLLink, 8000)
		if err != nil {
			return err
		}
	}

	log.Printf("info: Successfully enforced cap")
	return nil
}
