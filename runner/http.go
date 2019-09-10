package runner

import (
	"frpok/config"
	"github.com/matoous/go-nanoid"
	"log"

	"github.com/go-ini/ini"
)

// HTTPRunner run http cmd
type HTTPRunner struct {
	Runner
}

const alphabet string = "abcdefghijklmnopqrstuvwxyx1234567890"

// HTTPRun run frpc with random url
func HTTPRun(cfg *config.Config, ports []string) {

	if len(ports) == 0 {
		log.Fatal("Http command should have at least one port")
	}

	var runner HTTPRunner

	// init runtime config
	runner.init(cfg)
	runtimeCfg := ini.Empty()
	copySection(runtimeCfg, cfg.Common)

	httpTips := make([]string, len(ports))

	// get subdomian host
	subdomainHost := ""
	if cfg.Common.HasKey("subdomain_host") {
		key, err := cfg.Common.GetKey("subdomain_host")
		if err != nil {
			log.Fatal(err)
		}
		subdomainHost = key.Value()
	} else {
		key, err := cfg.Common.GetKey("server_addr")
		if err != nil {
			log.Fatal(err)
		}
		subdomainHost = key.Value()
	}

	for i, p := range ports {
		randomSubDomain, err := gonanoid.Generate(alphabet, 8)
		if err != nil {
			log.Fatal(err)
		}

		section, err := runtimeCfg.NewSection(randomSubDomain)

		if err != nil {
			log.Fatal(err)
		}

		section.NewKey("type", "http")
		section.NewKey("local_port", p)

		section.NewKey("subdomain", randomSubDomain)

		httpTips[i] = "Local port: " + p + "  ->  " + "http://" + randomSubDomain + "." + subdomainHost

	}

	// write run time config
	runner.writeRuntimeConfig(runtimeCfg)

	runner.runUI(httpTips[:])
}
