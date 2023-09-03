package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type URLNormalisations struct {
	AddWWW        []string          `toml:"add-www"`
	RemoveWWW     []string          `toml:"remove-www"`
	ReplaceDomain map[string]string `toml:"replace-domain"`
	ForceHTTPS    []string          `toml:"force-https"`
}

type ConfigType struct { //nolint:revive
	URLNormalisations URLNormalisations `toml:"UrlNormalisations"`
}

var Config ConfigType //nolint:gochecknoglobals

func LoadConfig() {
	configPath := "config.toml"

	configFile, err := os.Open(configPath)
	if err != nil {
		return
	}

	if _, err := toml.NewDecoder(configFile).Decode(&Config); err != nil {
		log.Printf("error decoding TOML: %s", err)

		return
	}
}
