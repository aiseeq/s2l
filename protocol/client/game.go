package client

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/protocol/api"
)

type GameConfig struct {
	netAddress  string
	processInfo ProcessInfo
	playerSetup []*api.PlayerSetup
	ports       Ports

	Client   *Client
	started  bool
	lastPort int
}

func NewGameConfig(participants ...*api.PlayerSetup) *GameConfig {
	config := &GameConfig{"127.0.0.1", ProcessInfo{}, nil, Ports{}, nil, false, 0}

	for _, p := range participants {
		if p.Type == api.PlayerType_Participant {
			config.Client = &Client{}
		}
		config.playerSetup = append(config.playerSetup, p)
	}
	return config
}

func (config *GameConfig) Connect(port int) {
	// Set process info for bot
	config.processInfo = ProcessInfo{Path: "", PID: 0, Port: port}

	// Since connect is blocking do it after the processes are launched.
	if err := config.Client.Connect(config.netAddress, config.processInfo.Port, processConnectTimeout); err != nil {
		log.Fatal("Failed to connect")
	}

	// Assume starcraft has started after succesfully attaching to a server
	config.started = true
}

func (config *GameConfig) SetupPorts(startPort int) {
	var ports = config.ports
	ports.SharedPort = int32(startPort + 1)
	ports.ServerPorts = &api.PortSet{
		GamePort: int32(startPort + 2),
		BasePort: int32(startPort + 3),
	}

	for i := 0; i < 2; i++ {
		var base = int32(startPort + 4 + i*2)
		ports.ClientPorts = append(ports.ClientPorts, &api.PortSet{GamePort: base, BasePort: base + 1})
	}
	config.ports = ports
}

func (config *GameConfig) CreateGame(mapPath string) bool {
	if !config.started {
		log.Fatal("Game not started")
	}

	err := config.Client.RequestCreateGame(mapPath, config.playerSetup, processRealtime)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (config *GameConfig) JoinGame() bool {
	if err := config.Client.RequestJoinGame(config.playerSetup[0], processInterfaceOptions, config.ports); err != nil {
		log.Fatalf("Unable to join game: %v", err)
	}

	return true
}

func (config *GameConfig) StartGame(mapPath string) {
	if !config.CreateGame(mapPath) {
		log.Fatal("Failed to create game.")
	}
	config.JoinGame()
}

func LaunchAndJoin(bot, cpu *api.PlayerSetup) *GameConfig {
	if !LoadSettings() {
		log.Fatal("Can't load settings")
	}
	var config *GameConfig
	if LadderGamePort > 0 {
		// Game against other bot or human via Ladder Manager
		config = NewGameConfig(bot)
		log.Info("Connecting to port ", LadderGamePort)
		config.Connect(LadderGamePort)
		config.SetupPorts(LadderStartPort)
		config.JoinGame()
		log.Info("Successfully joined game")
	} else {
		// Local game versus cpu
		config = NewGameConfig(bot, cpu)
		config.LaunchStarcraft()
		config.StartGame(MapPath())
	}

	return config
}
