package main

import (
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
)

type Config struct {
	aclOfsName    string
	aclOfsAddr    string
	extOfsName    string
	extOfsAddr    string
	extOfsAppAddr string
	wifiLink      *netlinkext.LinkExt
}

func NewConfig() *Config {
	return &Config{
		aclOfsName:    "crebas-acl-ofs",
		aclOfsAddr:    "192.168.10.1/24",
		extOfsName:    "crebas-ext-ofs",
		extOfsAddr:    "192.168.20.254/24",
		extOfsAppAddr: "192.168.20.1/24",
	}
}
