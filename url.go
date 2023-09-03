package main

import (
	"net/url"
	"slices"
	"strings"

	"github.com/benjamineskola/bookmarks/config"
)

func normaliseAddWWW(inputURL url.URL) url.URL {
	normalisationAddWWW := config.Config.URLNormalisations.AddWWW

	if slices.Contains(normalisationAddWWW, inputURL.Host) {
		inputURL.Host = "www." + inputURL.Host
	}

	return inputURL
}

func normaliseRemoveWWW(inputURL url.URL) url.URL {
	normalisationRemoveWWW := config.Config.URLNormalisations.RemoveWWW

	if slices.Contains(normalisationRemoveWWW, inputURL.Host) {
		inputURL.Host = strings.TrimPrefix(inputURL.Host, "www.")
	}

	return inputURL
}

func normaliseReplaceDomain(inputURL url.URL) url.URL {
	normalisationReplaceDomain := config.Config.URLNormalisations.ReplaceDomain

	if newDomain := normalisationReplaceDomain[inputURL.Host]; newDomain != "" {
		inputURL.Host = newDomain
	}

	return inputURL
}

func normaliseForceHTTPS(inputURL url.URL) url.URL {
	normalisationForceHTTPS := config.Config.URLNormalisations.ForceHTTPS

	if slices.Contains(normalisationForceHTTPS, inputURL.Host) && inputURL.Scheme != "https" {
		inputURL.Scheme = "https"
	}

	return inputURL
}

func normaliseURL(inputURL url.URL) url.URL {
	if config.Config.URLNormalisations.AddWWW == nil {
		config.LoadConfig()
	}

	inputURL = normaliseAddWWW(inputURL)
	inputURL = normaliseRemoveWWW(inputURL)
	inputURL = normaliseReplaceDomain(inputURL)

	// special case
	if inputURL.Host == "medium.com" || strings.HasSuffix(inputURL.Host, ".medium.com") {
		inputURL.Host = "scribe.rip"
	}

	inputURL = normaliseForceHTTPS(inputURL)

	return inputURL
}
