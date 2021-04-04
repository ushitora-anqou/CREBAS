package main

import (
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/pkg"
)

var apps = app.AppCollection{}
var pkgs = pkg.PkgCollection{}

func main() {
	StartWebServer()
}
