package main

type Config struct {
	aclOfsName string
	aclOfsAddr string
	extOfsName string
	extOfsAddr string
}

func NewConfig() *Config {
	return &Config{
		aclOfsName: "crebas-acl-ofs",
		aclOfsAddr: "192.168.10.1/24",
		extOfsName: "crebas-ext-ofs",
		extOfsAddr: "192.168.20.1/24",
	}
}
