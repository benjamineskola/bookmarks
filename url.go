package main

import (
	"net/url"
	"strings"
)

var normalisationAddWWW = map[string]bool{
	"theguardian.com": true,
}

var normalisationRemoveWWW = map[string]bool{
	"www.jacobin.com":      true,
	"www.jacobinmag.com":   true,
	"www.tribunemag.co.uk": true,
}

var normalisationReplaceDomain = map[string]string{
	"jacobinmag.com": "jacobin.com",
}

var normalisationForceHTTPS = map[string]bool{
	"www.theguardian.com": true,
	"jacobin.com":         true,
	"tribunemag.co.uk":    true,
	"newsocialist.org.uk": true,
}

func normaliseURL(inputURL url.URL) url.URL {
	if normalisationAddWWW[inputURL.Host] {
		inputURL.Host = "www." + inputURL.Host
	}

	if normalisationRemoveWWW[inputURL.Host] {
		inputURL.Host = strings.TrimPrefix(inputURL.Host, "www.")
	}

	if newDomain := normalisationReplaceDomain[inputURL.Host]; newDomain != "" {
		inputURL.Host = newDomain
	}

	// special case
	if inputURL.Host == "medium.com" || strings.HasSuffix(inputURL.Host, ".medium.com") {
		inputURL.Host = "scribe.rip"
	}

	if normalisationForceHTTPS[inputURL.Host] && inputURL.Scheme != "https" {
		inputURL.Scheme = "https"
	}

	return inputURL
}
