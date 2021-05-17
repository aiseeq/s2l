package main

import (
	log "bitbucket.org/aisee/minilog"
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/client"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func fixId(id string) string {
	id = strings.Replace(id, " ", "_", -1)
	id = strings.Replace(id, "@", "_", -1)
	for _, c := range id {
		if !unicode.IsLetter(c) {
			return "A_" + id
		}
		if unicode.IsLower(c) {
			return string(unicode.ToUpper(c)) + id[1:]
		}
		break
	}
	return id
}

func writeEnum(name string, apiType string, names []string, values map[string]uint32) {
	pkgName := strings.ToLower(name)
	fmtString := "\t%-*v api." + apiType + " = %v\n"

	maxLen, maxVal := 0, uint32(0)
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
		if val := values[name]; val > maxVal {
			maxVal = val
		}
	}

	path := filepath.Join("protocol", "enums", pkgName)
	check(os.MkdirAll(path, 0777))
	enumsFile, err := os.Create(filepath.Join(path, "enum.go"))
	check(err)
	defer enumsFile.Close()

	w := bufio.NewWriter(enumsFile)

	fmt.Fprint(w, "// Code generated by gen_ids. DO NOT EDIT.\npackage "+
		pkgName+"\n\nimport \"github.com/aiseeq/s2l/protocol/api\"\n\nconst (\n")

	for _, name := range names {
		fmt.Fprintf(w, fmtString, maxLen, name, values[name])
	}
	fmt.Fprint(w, ")\n")
	check(w.Flush())

	if !strings.HasPrefix(strings.ToLower(apiType), name) {
		return
	}

	// String() function
	fmtString2 := "\t%-*v \"%v\",\n"
	stringsFile, err := os.Create(filepath.Join(path, "strings.go"))
	check(err)
	defer stringsFile.Close()

	w2 := bufio.NewWriter(stringsFile)

	fmt.Fprint(w2, "// Code generated by gen_ids. DO NOT EDIT.\npackage "+pkgName+
		"\n\nimport \"github.com/aiseeq/s2l/protocol/api\"\n\n"+
		"func String(e api."+apiType+") string {\n\treturn strings[uint32(e)]\n}\n\nvar strings = map[uint32]string{\n")

	maxDigits := int(math.Ceil(math.Log10(float64(maxVal)))) + 1
	for _, name := range names {
		fmt.Fprintf(w2, fmtString2, maxDigits, strconv.Itoa(int(values[name]))+":", name)
	}
	fmt.Fprint(w2, "}\n")
	check(w2.Flush())
}

func dumpAbilities(abilities []*api.AbilityData) {
	// Detect base abilities of things with assigned hotkeys
	remaps := map[api.AbilityID]bool{}
	for _, ability := range abilities {
		if ability.GetAvailable() && ability.ButtonName != "" {
			if ability.RemapsToAbilityId != 0 && ability.Hotkey != "" {
				remaps[ability.RemapsToAbilityId] = true
			}
		}
	}

	// Find values to export and detect duplicate names
	byName := map[string]int{}
	for _, ability := range abilities {
		if ability.GetAvailable() && ability.ButtonName != "" {
			if ability.Hotkey != "" || remaps[ability.AbilityId] {
				byName[ability.FriendlyName] = byName[ability.FriendlyName] + 1
			}
		}
	}

	// Generate the values
	var names []string
	values := map[string]uint32{}
	for _, ability := range abilities {
		n := byName[ability.FriendlyName]
		if n == 0 {
			continue
		}

		if ability.GetAvailable() && ability.ButtonName != "" {
			if ability.Hotkey != "" || remaps[ability.AbilityId] {
				name := ability.FriendlyName
				if n > 1 {
					name = fmt.Sprintf("%v %v", name, uint32(ability.AbilityId))
				}
				name = fixId(name)

				names = append(names, name)
				values[name] = uint32(ability.AbilityId)
			}
		}
	}
	sort.Strings(names)

	values["Invalid"] = 0
	values["Smart"] = 1
	writeEnum("ability", "AbilityID", append([]string{"Invalid", "Smart"}, names...), values)
}

func dumpBuffs(buffs []*api.BuffData) {
	var names []string
	values := map[string]uint32{}
	for _, buff := range buffs {
		if name := fixId(buff.GetName()); name != "" {
			names = append(names, name)
			values[name] = uint32(buff.BuffId)
		}
	}
	//sort.Strings(names)

	values["Invalid"] = 0
	writeEnum("buff", "BuffID", append([]string{"Invalid"}, names...), values)
}

func dumpEffects(effects []*api.EffectData) {
	var names []string
	values := map[string]uint32{}
	for _, effect := range effects {
		if name := fixId(effect.GetFriendlyName()); name != "" {
			names = append(names, name)
			values[name] = uint32(effect.EffectId)
		}
	}

	values["Invalid"] = 0
	writeEnum("effect", "EffectID", append([]string{"Invalid"}, names...), values)
}

func dumpUnits(units []*api.UnitTypeData) {
	var names []string
	values := map[string]uint32{}
	namesByRace := map[string][]string{}
	valuesByRace := map[string]map[string]uint32{}
	for _, unit := range units {
		if unit.GetAvailable() && unit.Name != "" {
			race := unit.Race.String()
			if race == "NoRace" {
				race = "Neutral"
			}
			name := fixId(race + "_" + unit.Name)

			names = append(names, name)
			values[name] = uint32(unit.UnitId)

			namesByRace[race] = append(namesByRace[race], fixId(unit.Name))
			if valuesByRace[race] == nil {
				valuesByRace[race] = make(map[string]uint32)
			}
			valuesByRace[race][unit.Name] = uint32(unit.UnitId)
		}
	}
	sort.Strings(names)

	values["Invalid"] = 0
	writeEnum("unit", "UnitTypeID", append([]string{"Invalid"}, names...), values)

	for race, names := range namesByRace {
		sort.Strings(names)

		writeEnum(strings.ToLower(race), "UnitTypeID", names, valuesByRace[race])
	}
}

func dumpUpgrades(upgrades []*api.UpgradeData) {
	var names []string
	values := map[string]uint32{}
	for _, upgrade := range upgrades {
		if name := fixId(upgrade.GetName()); name != "" {
			names = append(names, name)
			values[name] = uint32(upgrade.UpgradeId)
		}
	}

	values["Invalid"] = 0
	writeEnum("upgrade", "UpgradeID", append([]string{"Invalid"}, names...), values)
}

func dumpVersion(ping api.ResponsePing) {
	check(os.MkdirAll(filepath.Join("protocol", "version"), 0777))
	file, err := os.Create(filepath.Join("protocol", "version", "version.go"))
	check(err)
	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprint(w, "// Code generated by gen_ids. DO NOT EDIT.\npackage version\n\nconst (\n")
	fmt.Fprintf(w, "\tGameVersion = %#v\n", ping.GameVersion)
	fmt.Fprintf(w, "\tDataVersion = %#v\n", ping.DataVersion)
	fmt.Fprintf(w, "\tDataBuild   = %v\n", ping.DataBuild)
	fmt.Fprintf(w, "\tBaseBuild   = %v\n", ping.BaseBuild)
	fmt.Fprint(w, ")\n")
	check(w.Flush())
}

func generate(c *client.Client) {
	// c.gameInfo, infoErr = c.GameInfo()
	// c.observation, obsErr = c.Observation(api.RequestObservation{})
	data, err := c.Data(api.RequestData{
		AbilityId:  true,
		UnitTypeId: true,
		UpgradeId:  true,
		BuffId:     true,
		EffectId:   true,
	})
	if err != nil {
		log.Fatal(err)
	}
	dumpAbilities(data.GetAbilities())
	dumpBuffs(data.GetBuffs())
	dumpEffects(data.GetEffects())
	dumpUnits(data.GetUnits())
	dumpUpgrades(data.GetUpgrades())

	log.Infof("%v", c.ResponsePing)
	dumpVersion(c.ResponsePing)
}

func main() {
	bot := client.NewParticipant(api.Race_Random, "NilBot")
	config := client.NewGameConfig(bot)
	config.LaunchStarcraft()
	config.StartGame(client.MapPath())

	c := config.Client
	generate(c)
}
