package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Cache defines a cache entry
// it is used to track the IDs of messages that have been posted already
// thus avoiding reposting messages that have already be seen before
type Cache struct {
	GUIDs map[string]bool
}

// newCache creates a new cache
func newCache() (c *Cache) {
	c = &Cache{}
	c.GUIDs = make(map[string]bool)
	return
}

// loadCache loads a previously saveCache'd file
func loadCache(fn string) (c *Cache, err error) {
	c = newCache()

	cf, err := os.ReadFile(fn)
	if os.IsNotExist(err) {
		// Acceptable, causes an empty cache to have been loaded
		err = nil
		return
	}
	if err != nil {
		fmt.Errorf("Loading Cache from %s failed: %s", fn, err)
		return
	}

	// Decode the JSON fields we have
	err = json.Unmarshal(cf, c)
	if err != nil {
		fmt.Errorf("Unmarshalling Cache from %s failed: %s", fn, err)
		return
	}

	return
}

// saveCache saves a cache to the given file
func saveCache(fn string, cache *Cache) (err error) {
	// Store the cache so we can use it again
	cf, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Errorf("Opening Cache for saving to %s failed: %s", fn, err)
		return
	}

	defer cf.Close()

	// Encode and write
	je := json.NewEncoder(cf)
	err = je.Encode(*cache)
	if err != nil {
		fmt.Errorf("Encoding Cache for saving to %s failed: %s", fn, err)
		return
	}

	return
}
