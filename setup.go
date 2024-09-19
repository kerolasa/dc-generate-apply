package main

import (
	"encoding/json"
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

func readTemplate(file string) Template {
	log.Debug().Str("template", file).Msg("reading template")

	var template Template

	templateFD, err := os.Open(flag.Arg(0))
	defer templateFD.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open file")
	}

	templateBytes, err := io.ReadAll(templateFD)
	if err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("cannot read file")
	}

	err = json.Unmarshal(templateBytes, &template)
	if err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("cannot unmarshal json")
	}

	return template
}

func getSettings(fqdn string) Settings {
	log.Debug().Str("fqdn", fqdn).Msg("getting settings")

	var settings Settings

	txtrecords, err := net.LookupTXT("_domainconnect." + fqdn)
	if err != nil {
		log.Fatal().Err(err).Str("fqdn", fqdn).Msg("dns lookup failed")
	}

	if len(txtrecords) != 1 {
		log.Fatal().Str("fqdn", fqdn).Int("number", len(txtrecords)).Msg("unexpected number of txt records")
	}

	settingsBytes := webGet("https://" + txtrecords[0] + "/v2/" + fqdn + "/settings")

	err = json.Unmarshal(settingsBytes, &settings)
	if err != nil {
		log.Fatal().Err(err).Str("fqdn", fqdn).Msg("cannot unmarshal json")
	}

	return settings
}

func webGet(url string) []byte {
	log.Debug().Str("url", url).Msg("web get url")

	var webClient = &http.Client{Timeout: 2 * time.Second}

	response, err := webClient.Get(url)
	defer response.Body.Close()
	if err != nil {
		log.Fatal().Err(err).Str("url", url).Msg("web get failed")
	}

	webBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal().Err(err).Str("url", url).Msg("cannot read response body")
	}

	return webBytes
}

func compareVersions(template Template, settings Settings) {
	log.Debug().Msg("comparing versions")

	url := settings.URLAPI + "/v2/domainTemplates/providers/" + template.ProviderID + "/services/" + template.ServiceID
	versionBytes := webGet(url)

	var version DNSProviderVersion
	if err := json.Unmarshal(versionBytes, &version); err != nil {
		log.Fatal().Err(err).Str("url", url).Msg("cannot unmarshal json")
	}

	if version.Version != template.Version {
		log.Fatal().Str("url", url).Uint("template", template.Version).Uint("upstream", version.Version).Msg("template and upstream versions disagree")
	}
}
