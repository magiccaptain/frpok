package config

import (
	"github.com/go-ini/ini"
)

// Config struct of frpok config
type Config struct {
	// frp common config
	Common *ini.Section
	// frpok common config
	Frpok *ini.Section
}
