package main

import (
	"encoding/json"
	"os"
)

// Cfg contains the general configuration, normally stored in config.json
type Cfg struct {
	UserAgent         string
	MstdnServer       string
	MstdnClientID     string
	MstdnClientSecret string
	CacheDir          string
	Accounts          []Account
}

// loadConfig loads the configuration
func loadConfig(fn string) (c *Cfg, err error) {
	var cfg Cfg

	// Read the complete file (generally relatively small)
	cf, err := os.ReadFile(fn)
	if err != nil {
		return
	}

	// Unmarshall the JSON from the file
	err = json.Unmarshal(cf, &cfg)
	if err != nil {
		return
	}

	// Return a pointer to the config and no error
	return &cfg, nil
}
