package client

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	LadderGamePort   = 0
	LadderStartPort  = 0
	LadderServer     = ""
	LadderOpponentID = ""
)

var Maps2021season1 = []string{
	"DeathAura506",
	"EternalEmpire506",
	"EverDream506",
	"GoldenWall506",
	"IceandChrome506",
	"PillarsofGold506",
	"Submarine506",
}
var MapName = Random1v1Map()

func init() {
	// Ladder Flags
	flagInt("GamePort", &LadderGamePort, "Port of client to connect to")
	flagInt("StartPort", &LadderStartPort, "Starting server port")
	flagStr("LadderServer", &LadderServer, "Ladder server address")
	flagStr("OpponentId", &LadderOpponentID, "Ladder ID of the opponent (for learning bots)")
	flagStr("Map", &MapName, "Which map to run.")
}

// Set changes the default value of a command line flag.
func Set(name, value string) {
	if err := flag.Set(name, value); err != nil {
		log.Print(err)
	}
}

var hasLoaded = false

func LoadSettings() bool {
	if flag.Parsed() {
		return hasLoaded
	}

	// Parse the command line arguments
	showHelp := flag.Bool("help", false, "Prints help message")
	flag.Parse()
	if *showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !hasProcessPath() {
		log.Println("Can't find executable path, hope that it's ok. If not, " +
			"please run StarCraft II first or use the --executable <path> arg")
	}

	hasLoaded = true
	return true
}

func flagStr(name string, value *string, usage string) {
	flag.StringVar(value, name, *value, usage)
}

func flagInt(name string, value *int, usage string) {
	flag.IntVar(value, name, *value, usage)
}

func flagBool(name string, value *bool, usage string) {
	flag.BoolVar(value, name, *value, usage)
}

func flagDur(name string, value *time.Duration, usage string) {
	flag.DurationVar(value, name, *value, usage)
}
