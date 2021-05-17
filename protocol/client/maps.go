package client

import (
	"math/rand"
	"path/filepath"
	"runtime"
	"time"
)

// SetMap sets the map to use (via flag)
func SetMap(name string) {
	Set("Map", name)
}

// Random1v1Map returns a random map name from the current 1v1 ladder map pool.
func Random1v1Map() string {
	currentMaps := Maps2021season1

	rand.Seed(time.Now().UnixNano())
	return currentMaps[rand.Intn(len(currentMaps))] + ".SC2Map"
}

func MapPath() string {
	// Fix linux client using maps directory instead of Maps
	if runtime.GOOS != "windows" {
		return filepath.Join(defaultSc2Path(), "Maps", MapName)
	}
	return MapName
}
