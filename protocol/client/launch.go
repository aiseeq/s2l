package client

import (
	log "bitbucket.org/aisee/minilog"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	launchBaseBuild        = uint32(0)
	launchDataVersion      = ""
	LaunchPortStart        = 8168
	launchExtraCommandArgs = []string(nil)
)

// SetGameVersion specifies a specific base game and data version to use when launching.
func SetGameVersion(baseBuild uint32, dataVersion string) {
	launchBaseBuild = baseBuild
	launchDataVersion = dataVersion
}

func StartProcess(path string, args []string) int {
	cmd := exec.Command(path, args...)

	// Set the working directory on windows
	if runtime.GOOS == "windows" {
		_, exe := filepath.Split(path)
		dir := sc2Path(path)
		if strings.Contains(exe, "_x64") {
			dir = filepath.Join(dir, "Support64")
		} else {
			dir = filepath.Join(dir, "Support")
		}
		cmd.Dir = dir
	}

	if err := cmd.Start(); err != nil {
		log.Error(err)
		return 0
	}

	return cmd.Process.Pid
}

func (config *GameConfig) LaunchAndAttach(path string, c *Client) ProcessInfo {
	pi := ProcessInfo{}
	pi.Port = LaunchPortStart

	// See if we can connect to an old instance real quick before launching
	if err := c.TryConnect(config.netAddress, pi.Port); err != nil {
		args := []string{
			"-listen", config.netAddress,
			"-port", strconv.Itoa(pi.Port),
			// DirectX will fail if multiple games try to launch in fullscreen mode. Force them into windowed mode.
			"-displayMode", "0",
		}

		if len(launchDataVersion) > 0 {
			args = append(args, "-dataVersion", launchDataVersion)
		}
		args = append(args, launchExtraCommandArgs...)

		pi.Path = path
		pi.PID = StartProcess(pi.Path, args)
		if pi.PID == 0 {
			log.Error("Unable to start sc2 executable with path: ", pi.Path)
		} else {
			log.Infof("Launched SC2 (%v), PID: %v", pi.Path, pi.PID)
		}

		// Attach
		if err := c.Connect(config.netAddress, pi.Port, processConnectTimeout); err != nil {
			log.Fatal("Failed to connect")
		}
	}

	return pi
}

func ProcessPathForBuild(build uint32) string {
	path := processPath
	if build != 0 {
		// Get the exe name and then back out to the Versions directory
		_, exe := filepath.Split(path)
		root := sc2Path(path)
		if root == "" {
			log.Errorf("Can't find game dir: %v", path)
		}
		dir := filepath.Join(sc2Path(path), "Versions")

		// Get the path of the correct version and make sure the exe exists
		path = filepath.Join(dir, fmt.Sprintf("Base%v", build), exe)
		if _, err := os.Stat(path); err != nil {
			log.Errorf("Base version not found: %v", err)
		}
	}
	return path
}

func (config *GameConfig) LaunchProcess(client *Client) ProcessInfo {
	// Make sure we have a valid executable path
	path := ProcessPathForBuild(launchBaseBuild)
	if _, err := os.Stat(path); err != nil {
		log.Error("Executable path can't be found, try running the StarCraft II executable first.")
		if len(path) > 0 {
			log.Errorf("%v does not exist on your filesystem.", path)
		}
	}

	return config.LaunchAndAttach(path, client)
}

func (config *GameConfig) LaunchStarcraft() {
	config.processInfo = config.LaunchProcess(config.Client)
	config.started = true
	config.lastPort = LaunchPortStart
}
