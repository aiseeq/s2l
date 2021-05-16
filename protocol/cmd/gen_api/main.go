package main

import (
	"bitbucket.org/aisee/minilog"
	"bufio"
	"fmt"
	"github.com/aiseeq/helpers/pkg/file"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// If doesn't work: go get -u github.com/gogo/protobuf

const (
	importPrefix   = "import \"s2clientprotocol/"
	optionalPrefix = "optional "
	enumPrefix     = "enum "
)

var ApiDir = filepath.Join("protocol", "api")
var SourceDir = filepath.Join("protocol", "s2client-proto")
var OutputDir = filepath.Join("protocol", "output")
var typeMap = map[string]string{ // Make the API more type-safe
	// common.proto
	"AvailableAbility.ability_id": "AbilityID",
	// data.proto
	"AbilityData.ability_id":           "AbilityID",
	"AbilityData.remaps_to_ability_id": "AbilityID",
	"UnitTypeData.unit_id":             "UnitTypeID",
	"UnitTypeData.ability_id":          "AbilityID",
	"UnitTypeData.tech_alias":          "UnitTypeID",
	"UnitTypeData.unit_alias":          "UnitTypeID",
	"UnitTypeData.tech_requirement":    "UnitTypeID",
	"UpgradeData.upgrade_id":           "UpgradeID",
	"UpgradeData.ability_id":           "AbilityID",
	"BuffData.buff_id":                 "BuffID",
	"EffectData.effect_id":             "EffectID",
	// debug.proto
	"DebugCreateUnit.unit_type":  "UnitTypeID",
	"DebugCreateUnit.owner":      "PlayerID",
	"DebugKillUnit.tag":          "UnitTag",
	"DebugSetUnitValue.unit_tag": "UnitTag",
	// error.proto
	// query.proto
	"RequestQueryPathing.start.unit_tag":             "UnitTag",
	"RequestQueryAvailableAbilities.unit_tag":        "UnitTag",
	"ResponseQueryAvailableAbilities.unit_tag":       "UnitTag",
	"ResponseQueryAvailableAbilities.unit_type_id":   "UnitTypeID",
	"RequestQueryBuildingPlacement.ability_id":       "AbilityID",
	"RequestQueryBuildingPlacement.placing_unit_tag": "UnitTag",
	// raw.proto
	"PowerSource.tag":                             "UnitTag",
	"PlayerRaw.upgrade_ids":                       "UpgradeID",
	"UnitOrder.ability_id":                        "AbilityID",
	"UnitOrder.target.target_unit_tag":            "UnitTag",
	"PassengerUnit.tag":                           "UnitTag",
	"PassengerUnit.unit_type":                     "UnitTypeID",
	"Unit.tag":                                    "UnitTag",
	"Unit.unit_type":                              "UnitTypeID",
	"Unit.owner":                                  "PlayerID",
	"Unit.add_on_tag":                             "UnitTag",
	"Unit.buff_ids":                               "BuffID",
	"Unit.engaged_target_tag":                     "UnitTag",
	"Event.dead_units":                            "UnitTag",
	"Effect.effect_id":                            "EffectID",
	"ActionRawUnitCommand.ability_id":             "AbilityID",
	"ActionRawUnitCommand.target.target_unit_tag": "UnitTag",
	"ActionRawUnitCommand.unit_tags":              "UnitTag",
	"ActionRawToggleAutocast.ability_id":          "AbilityID",
	"ActionRawToggleAutocast.unit_tags":           "UnitTag",
	// sc2api.proto
	"RequestJoinGame.participation.observed_player_id": "PlayerID",
	"ResponseJoinGame.player_id":                       "PlayerID",
	"RequestStartReplay.observed_player_id":            "PlayerID",
	"ChatReceived.player_id":                           "PlayerID",
	"PlayerInfo.player_id":                             "PlayerID",
	"PlayerCommon.player_id":                           "PlayerID",
	"ActionError.unit_tag":                             "UnitTag",
	"ActionError.ability_id":                           "AbilityID",
	"ActionObserverPlayerPerspective.player_id":        "PlayerID",
	"ActionObserverCameraFollowPlayer.player_id":       "PlayerID",
	"ActionObserverCameraFollowUnits.unit_tags":        "UnitTag",
	"PlayerResult.player_id":                           "PlayerID",
	// score.proto
	// spatial.proto
	// ui.proto
	"ControlGroup.leader_unit_type":   "UnitTypeID",
	"UnitInfo.unit_type":              "UnitTypeID",
	"BuildItem.ability_id":            "AbilityID",
	"ActionToggleAutocast.ability_id": "AbilityID",
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func execAndCheck(dir, binary string, args ...string) {
	cmd := exec.Command(binary, args...)
	if dir != "" {
		cmd.Dir = SourceDir
	}
	out, err := cmd.CombinedOutput()
	log.Info(string(out))
	check(err)
}

func mapTypes(path []string, line string) string {
	parts := strings.Split(line, " ")
	if len(parts) < 4 {
		return line // need at least "<type> <name> = <num>;"
	}

	key := strings.Join(path, ".") + "." + parts[len(parts)-3]
	if value, ok := typeMap[key]; ok {
		// Add the casttype option
		opt := fmt.Sprintf("[(gogoproto.casttype) = \"%v\"];", value)

		last := parts[len(parts)-1]
		parts = append(parts[:len(parts)-1], last[:len(last)-1], opt)
		delete(typeMap, key) // track which ones have been processed
		return strings.Join(parts, " ")
	}

	return line
}

func updateProto(path string) []string {
	f, err := os.Open(path)
	check(err)
	defer f.Close()

	var propPath, lines []string

	// Read line by line, making modifications as needed
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Get the line and trim comments and whitespace to make matching easier
		line := scanner.Text()
		if comment := strings.Index(line, "//"); comment > 0 {
			line = line[:comment]
		}
		line = strings.TrimSpace(line)

		switch {
		// Upgrade to proto3 and set the go package name
		case line == "syntax = \"proto2\";":
			lines = append(lines, "syntax = \"proto3\";", "option go_package = \"./;api\";\nimport \"gogo.proto\";")

		// Remove subdirectory of the import so the output path isn't nested
		case strings.HasPrefix(line, importPrefix):
			lines = append(lines, "import \""+line[len(importPrefix):])

		// Remove "optional" prefixes (they are implicit in proto3)
		case strings.HasPrefix(line, optionalPrefix):
			lines = append(lines, mapTypes(propPath, line[len(optionalPrefix):]))

		// Track where we are in the path
		case strings.HasSuffix(line, " {"):
			id := strings.Split(line, " ")[1] // "<type> Identifier {"
			propPath = append(propPath, id)

			lines = append(lines, line)

			// Enums must have a zero value in proto3 (and unfortunately they must be unique due to C++ scoping rules)
			if strings.HasPrefix(line, enumPrefix) && line != "enum Race {" && line != "enum CloakState {" {
				lines = append(lines, line[len(enumPrefix):len(line)-2]+"_nil = 0 [(gogoproto.enumvalue_customname) = \"nil\"];")
			}

		// Pop the last path element
		case line == "}":
			if propPath[len(propPath)-1] == "Unit" {
				lines = append(lines,
					"repeated AvailableAbility actions = 100;",
				)
			}
			propPath = propPath[:len(propPath)-1]
			lines = append(lines, line)

		// Everything else just gets copied to the output
		default:
			lines = append(lines, mapTypes(propPath, line))
		}
	}

	return lines
}

func writeFile(path string, lines []string) {
	f, err := os.Create(path)
	check(err)
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	check(writer.Flush())
}

func main() {
	if !file.Exists(SourceDir) {
		execAndCheck("", "git", "clone", "git@github.com:Blizzard/s2client-proto.git", SourceDir)
	} else {
		execAndCheck(SourceDir, "git", "fetch", "--all")
		execAndCheck(SourceDir, "git", "reset", "--hard", "origin/master")
	}

	// Get all the .proto files
	protoDir := filepath.Join(SourceDir, "s2clientprotocol")
	files, err := ioutil.ReadDir(protoDir)
	check(err)

	protocArgs := []string{
		"-I=" + filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "gogo", "protobuf", "gogoproto"),
		"-I=" + filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "gogo", "protobuf", "protobuf"),
		"--proto_path=" + OutputDir,
		"--gogofaster_out=" + ApiDir,
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".proto" {
			continue
		}
		sourcePath := filepath.Join(protoDir, f.Name())
		newPath := filepath.Join(OutputDir, f.Name())

		// Upgrade the file to proto3 and fix the package name
		writeFile(newPath, updateProto(sourcePath))

		// Add the file to the list of command line args for protoc
		protocArgs = append(protocArgs, newPath)
	}

	// Make sure we mapped all the expected types
	if len(typeMap) != 0 {
		log.Warning("Not all types were mapped, missing:")
		for key := range typeMap {
			log.Warning(key)
		}
	}

	// Generate go code from the .proto files
	execAndCheck("", "protoc", protocArgs...)
}
