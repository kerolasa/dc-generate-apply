package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func readTemplate(file string) (Template, error) {
	var template Template

	templateFD, err := os.Open(flag.Arg(0))
	defer templateFD.Close()
	if err != nil {
		return template, fmt.Errorf("cannot open %s: %s", file, err)
	}

	templateBytes, err := io.ReadAll(templateFD)
	if err != nil {
		return template, fmt.Errorf("cannot read %s: %s", file, err)
	}

	err = json.Unmarshal(templateBytes, &template)
	if err != nil {
		return template, fmt.Errorf("cannot unmarshal json %s: %s", file, err)
	}

	return template, err
}

func getSettings(fqdn string) (Settings, error) {
	var settings Settings

	txtrecords, err := net.LookupTXT("_domainconnect." + fqdn)
	if err != nil {
		return settings, fmt.Errorf("lookup failed %s: %s", fqdn, err)
	}

	if len(txtrecords) != 1 {
		return settings, fmt.Errorf("unexpected number of txt records %s: %d", fqdn, len(txtrecords))
	}

	settingsBytes, err := webGet("https://" + txtrecords[0] + "/v2/" + fqdn + "/settings")

	err = json.Unmarshal(settingsBytes, &settings)
	if err != nil {
		return settings, fmt.Errorf("cannot unmarshal json %s: %s", fqdn, err)
	}

	return settings, nil
}

func webGet(url string) ([]byte, error) {
	var webClient = &http.Client{Timeout: 2 * time.Second}

	response, err := webClient.Get(url)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	webBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("webGet %s: %s", url, err)
	}

	return webBytes, nil
}

func compareVersions(template Template, settings Settings) (bool, error) {
	versionBytes, err := webGet(settings.URLAPI + "/v2/domainTemplates/providers/" + template.ProviderID + "/services/" + template.ServiceID)
	if err != nil {
		return false, err
	}

	var version DNSProviderVersion
	err = json.Unmarshal(versionBytes, &version)
	if err != nil {
		return false, err
	}

	if version.Version != template.Version {
		fmt.Printf("version check: %d != %d\n", version.Version, template.Version)
		return false, nil
	}

	return true, nil
}
