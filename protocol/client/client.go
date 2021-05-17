package client

import (
	"bitbucket.org/aisee/minilog"
	"fmt"
	"time"

	"github.com/aiseeq/s2l/protocol/api"
)

// Client ...
type Client struct {
	Connection
	realtime bool

	playerID    api.PlayerID
	gameInfo    *api.ResponseGameInfo
	replayInfo  *api.ResponseReplayInfo
	data        *api.ResponseData
	observation *api.ResponseObservation
	upgrades    map[api.UpgradeID]struct{}
	newUpgrades []api.UpgradeID

	beforeStep []func()
	subStep    []func()
	afterStep  []func()

	debugDraw chan struct{}

	perfInterval uint32
	lastDraw     []*api.DebugCommand

	perfStart       time.Time
	perfStartFrame  uint32
	beforeStepTime  time.Duration
	stepTime        time.Duration
	observationTime time.Duration
	afterStepTime   time.Duration

	actions          int
	maxActions       int
	actionsCompleted int
	observerActions  int
	debugCommands    int
}

// Connect ...
func (c *Client) Connect(address string, port int, timeout time.Duration) error {
	attempts := int(timeout.Seconds() + 1.5)
	if attempts < 1 {
		attempts = 1
	}

	connected := false
	for i := 0; i < attempts; i++ {
		if err := c.Connection.Connect(address, port); err == nil {
			connected = true
			break
		}
		time.Sleep(time.Second)

		if i == 0 {
			fmt.Print("Waiting for connection")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()

	if !connected {
		return fmt.Errorf("unable to connect to game")
	}

	log.Infof("Connected to %v:%v", address, port)
	return nil
}

// TryConnect ...
func (c *Client) TryConnect(address string, port int) error {
	if err := c.Connection.Connect(address, port); err != nil {
		return err
	}

	log.Infof("Connected to %v:%v", address, port)
	return nil
}

// CreateGame ...
func (c *Client) CreateGame(mapPath string, players []*api.PlayerSetup, realtime bool) error {
	r, err := c.Connection.CreateGame(api.RequestCreateGame{
		Map: &api.RequestCreateGame_LocalMap{
			LocalMap: &api.LocalMap{
				MapPath: mapPath,
			},
		},
		PlayerSetup: players,
		Realtime:    realtime,
	})
	if err != nil {
		return err
	}
	c.realtime = realtime

	if r.Error != api.ResponseCreateGame_nil {
		return fmt.Errorf("%v: %v", r.Error, r.GetErrorDetails())
	}

	return nil
}

// RequestJoinGame ...
func (c *Client) RequestJoinGame(setup *api.PlayerSetup, options *api.InterfaceOptions, ports Ports) error {
	req := api.RequestJoinGame{
		Participation: &api.RequestJoinGame_Race{
			Race: setup.Race,
		},
		Options: options,
	}
	if ports.isValid() {
		req.SharedPort = ports.SharedPort
		req.ServerPorts = ports.ServerPorts
		req.ClientPorts = ports.ClientPorts
	}
	r, err := c.Connection.JoinGame(req)
	if err != nil {
		return err
	}

	if r.Error != api.ResponseJoinGame_nil {
		return fmt.Errorf("%v: %v", r.Error.String(), r.GetErrorDetails())
	}

	c.playerID = r.GetPlayerId()
	return nil
}
