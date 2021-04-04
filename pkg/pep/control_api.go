package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/pkg"
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
		c.JSON(http.StatusNotFound, 0)
		return
	}

	pkg := selectedPkgs[0]
	proc := app.NewLinuxProcessFromPkgInfo(pkg)
	proc.Create()
	proc.Start()
	apps.Add(proc)

	c.JSON(http.StatusOK, proc.GetAppInfo())
}

func StartAPIServer() error {
	return setupRouter().Run()
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	r.GET("/pkgs", getAllPkgs)
	r.POST("/pkg/:id/start", startAppFromPkg)
	r.GET("/apps", getAllApps)

	return r
}
