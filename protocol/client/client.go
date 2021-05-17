package client

import (
	"bitbucket.org/aisee/minilog"
	"fmt"
	"time"

	"github.com/aiseeq/s2l/protocol/api"
)

type Client struct {
	api.ResponsePing

	Status api.Status

	counter  uint32
	requests chan<- request
}

func (c *Client) Connect(address string, port int, timeout time.Duration) error {
	attempts := int(timeout.Seconds() + 1.5)
	if attempts < 1 {
		attempts = 1
	}

	connected := false
	for i := 0; i < attempts; i++ {
		if err := c.Dial(address, port); err == nil {
			connected = true
			break
		}
		time.Sleep(time.Second)

		if i == 0 {
			log.Info("Waiting for connection")
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

func (c *Client) TryConnect(address string, port int) error {
	if err := c.Dial(address, port); err != nil {
		return err
	}

	log.Infof("Connected to %v:%v", address, port)
	return nil
}

func (c *Client) RequestCreateGame(mapPath string, players []*api.PlayerSetup, realtime bool) error {
	r, err := c.CreateGame(api.RequestCreateGame{
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

	if r.Error != api.ResponseCreateGame_nil {
		return fmt.Errorf("%v: %v", r.Error, r.GetErrorDetails())
	}

	return nil
}

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
	r, err := c.JoinGame(req)
	if err != nil {
		return err
	}

	if r.Error != api.ResponseJoinGame_nil {
		return fmt.Errorf("%v: %v", r.Error.String(), r.GetErrorDetails())
	}

	// c.playerID = r.GetPlayerId()
	return nil
}
