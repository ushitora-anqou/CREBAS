package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
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

func getAppInfo(c *gin.Context) {
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
	c.JSON(http.StatusOK, app.GetAppInfo())
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

	return r
}
