package client

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/protocol/api"
)

type GameConfig struct {
	netAddress  string
	processInfo []ProcessInfo
	playerSetup []*api.PlayerSetup
	ports       Ports

	Clients  []*Client
	started  bool
	lastPort int
}

func NewGameConfig(participants ...PlayerSetup) *GameConfig {
	config := &GameConfig{
		"127.0.0.1",
		nil,
		nil,
		Ports{},
		nil,
		false,
		0,
	}

	for _, p := range participants {
		if p.Type == api.PlayerType_Participant {
			config.Clients = append(config.Clients, &Client{})
		}
		config.playerSetup = append(config.playerSetup, p.PlayerSetup)
	}
	return config
}

func (config *GameConfig) StartGame(mapPath string) {
	if !config.CreateGame(mapPath) {
		log.Fatal("Failed to create game.")
	}
	config.JoinGame()
}

func (config *GameConfig) CreateGame(mapPath string) bool {
	if !config.started {
		log.Fatal("Game not started")
	}

	// Create with the first client
	err := config.Clients[0].CreateGame(mapPath, config.playerSetup, processRealtime)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (config *GameConfig) JoinGame() bool {
	// TODO: Make this parallel and get rid of the WaitJoinGame method
	for i, client := range config.Clients {
		if err := client.RequestJoinGame(config.playerSetup[i], processInterfaceOptions, config.ports); err != nil {
			log.Fatalf("Unable to join game: %v", err)
		}
	}

	return true
}

func (config *GameConfig) Connect(port int) {
	pi := ProcessInfo{Path: "", PID: 0, Port: port}

	// Set process info for each bot
	for range config.Clients {
		config.processInfo = append(config.processInfo, pi)
	}

	// Since connect is blocking do it after the processes are launched.
	for i, client := range config.Clients {
		pi := config.processInfo[i]

		if err := client.Connect(config.netAddress, pi.Port, processConnectTimeout); err != nil {
			log.Fatal("Failed to connect")
		}
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
