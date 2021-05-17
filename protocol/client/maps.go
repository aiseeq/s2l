package client

import (
	"math/rand"
	"path/filepath"
	"runtime"
	"time"
)

var maps2021season1 = []string{
	"DeathAura506",
	"EternalEmpire506",
	"EverDream506",
	"GoldenWall506",
	"IceandChrome506",
	"PillarsofGold506",
	"Submarine506",
}
var mapName = Random1v1Map()

func init() {
	flagStr("map", &mapName, "Which map to run.")
}

// SetMap sets the default map to use.
func SetMap(name string) {
	Set("map", name)
}

// Random1v1Map returns a random map name from the current 1v1 ladder map pool.
func Random1v1Map() string {
	currentMaps := maps2021season1

	rand.Seed(time.Now().UnixNano())
	return currentMaps[rand.Intn(len(currentMaps))] + ".SC2Map"
}

func MapPath() string {
	// Fix linux client using maps directory instead of Maps
	if runtime.GOOS != "windows" {
		return filepath.Join(defaultSc2Path(), "Maps", mapName)
	}
	return mapName
}
