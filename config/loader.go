package config

import (
	"log"
	"os"
	"path"

	"github.com/go-ini/ini"
)

// LoadIni load config file from filepath
func LoadIni(filepath string) *Config {

	var finalPath string

	if path.IsAbs(filepath) {
		finalPath = filepath
	} else {
		pwd, _ := os.Getwd()
		finalPath = path.Join(pwd, filepath)
	}

	cfg, err := ini.Load(finalPath)

	if err != nil {
		log.Fatal(err)
	}

	common, err := cfg.GetSection("common")

	if err != nil {
		log.Fatal(err)
	}

	frpok, err := cfg.GetSection("frpok")

	if err != nil {
		log.Fatal(err)
	}

	ret := Config{
		Common: common,
		Frpok:  frpok,
	}

	return &ret
}
