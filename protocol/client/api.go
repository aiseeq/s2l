package client

import (
	"github.com/aiseeq/s2l/protocol/api"
)

func (c *Client) CreateGame(createGame api.RequestCreateGame) (*api.ResponseCreateGame, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_CreateGame{
			CreateGame: &createGame,
		},
	})
	if r != nil {
		return r.GetCreateGame(), err
	}
	return nil, err
}

func (c *Client) JoinGame(joinGame api.RequestJoinGame) (*api.ResponseJoinGame, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_JoinGame{
			JoinGame: &joinGame,
		},
	})
	if r != nil {
		return r.GetJoinGame(), err
	}
	return nil, err
}

func (c *Client) RestartGame() (*api.ResponseRestartGame, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_RestartGame{
			RestartGame: &api.RequestRestartGame{},
		},
	})
	if r != nil {
		return r.GetRestartGame(), err
	}
	return nil, err
}

func (c *Client) StartReplay(startReplay api.RequestStartReplay) (*api.ResponseStartReplay, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_StartReplay{
			StartReplay: &startReplay,
		},
	})
	if r != nil {
		return r.GetStartReplay(), err
	}
	return nil, err
}

func (c *Client) LeaveGame() error {
	_, err := c.Request(&api.Request{
		Request: &api.Request_LeaveGame{
			LeaveGame: &api.RequestLeaveGame{},
		},
	})
	return err
}

func (c *Client) QuickSave() error {
	_, err := c.Request(&api.Request{
		Request: &api.Request_QuickSave{
			QuickSave: &api.RequestQuickSave{},
		},
	})
	return err
}

func (c *Client) QuickLoad() error {
	_, err := c.Request(&api.Request{
		Request: &api.Request_QuickLoad{
			QuickLoad: &api.RequestQuickLoad{},
		},
	})
	return err
}

func (c *Client) Quit() error {
	_, err := c.Request(&api.Request{
		Request: &api.Request_Quit{
			Quit: &api.RequestQuit{},
		},
	})
	return err
}

func (c *Client) GameInfo() (*api.ResponseGameInfo, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_GameInfo{
			GameInfo: &api.RequestGameInfo{},
		},
	})
	if r != nil {
		return r.GetGameInfo(), err
	}
	return nil, err
}

func (c *Client) Observation(observation api.RequestObservation) (*api.ResponseObservation, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_Observation{
			Observation: &observation,
		},
	})
	if r != nil {
		return r.GetObservation(), err
	}
	return nil, err
}

func (c *Client) Action(action api.RequestAction) (*api.ResponseAction, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_Action{
			Action: &action,
		},
	})
	if r != nil {
		return r.GetAction(), err
	}
	return nil, err
}

func (c *Client) ObsAction(obsAction api.RequestObserverAction) error {
	_, err := c.Request(&api.Request{
		Request: &api.Request_ObsAction{
			ObsAction: &obsAction,
		},
	})
	return err
}

func (c *Client) Step(step api.RequestStep) (*api.ResponseStep, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_Step{
			Step: &step,
		},
	})
	if r != nil {
		return r.GetStep(), err
	}
	return nil, err
}

func (c *Client) Data(data api.RequestData) (*api.ResponseData, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_Data{
			Data: &data,
		},
	})
	if r != nil {
		return r.GetData(), err
	}
	return nil, err
}

func (c *Client) Query(query api.RequestQuery) (*api.ResponseQuery, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_Query{
			Query: &query,
		},
	})
	if r != nil {
		return r.GetQuery(), err
	}
	return nil, err
}

func (c *Client) SaveReplay() (*api.ResponseSaveReplay, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_SaveReplay{
			SaveReplay: &api.RequestSaveReplay{},
		},
	})
	if r != nil {
		return r.GetSaveReplay(), err
	}
	return nil, err
}

func (c *Client) MapCommand(mapCommand api.RequestMapCommand) (*api.ResponseMapCommand, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_MapCommand{
			MapCommand: &mapCommand,
		},
	})
	if r != nil {
		return r.GetMapCommand(), err
	}
	return nil, err
}

func (c *Client) ReplayInfo(replayInfo api.RequestReplayInfo) (*api.ResponseReplayInfo, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_ReplayInfo{
			ReplayInfo: &replayInfo,
		},
	})
	if r != nil {
		return r.GetReplayInfo(), err
	}
	return nil, err
}

func (c *Client) AvailableMaps() (*api.ResponseAvailableMaps, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_AvailableMaps{
			AvailableMaps: &api.RequestAvailableMaps{},
		},
	})
	if r != nil {
		return r.GetAvailableMaps(), err
	}
	return nil, err
}

func (c *Client) SaveMap(saveMap api.RequestSaveMap) (*api.ResponseSaveMap, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_SaveMap{
			SaveMap: &saveMap,
		},
	})
	if r != nil {
		return r.GetSaveMap(), err
	}
	return nil, err
}

func (c *Client) Ping() (*api.ResponsePing, error) {
	r, err := c.Request(&api.Request{
		Request: &api.Request_Ping{
			Ping: &api.RequestPing{},
		},
	})
	if r != nil {
		return r.GetPing(), err
	}
	return nil, err
}

func (c *Client) Debug(debug api.RequestDebug) error {
	_, err := c.Request(&api.Request{
		Request: &api.Request_Debug{
			Debug: &debug,
		},
	})
	return err
}
