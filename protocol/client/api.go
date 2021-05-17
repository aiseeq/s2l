package client

import (
	"github.com/aiseeq/s2l/protocol/api"
)

func (c *Connection) CreateGame(createGame api.RequestCreateGame) (*api.ResponseCreateGame, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_CreateGame{
			CreateGame: &createGame,
		},
	})
	if r != nil {
		return r.GetCreateGame(), err
	}
	return nil, err
}

func (c *Connection) JoinGame(joinGame api.RequestJoinGame) (*api.ResponseJoinGame, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_JoinGame{
			JoinGame: &joinGame,
		},
	})
	if r != nil {
		return r.GetJoinGame(), err
	}
	return nil, err
}

func (c *Connection) RestartGame() (*api.ResponseRestartGame, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_RestartGame{
			RestartGame: &api.RequestRestartGame{},
		},
	})
	if r != nil {
		return r.GetRestartGame(), err
	}
	return nil, err
}

func (c *Connection) StartReplay(startReplay api.RequestStartReplay) (*api.ResponseStartReplay, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_StartReplay{
			StartReplay: &startReplay,
		},
	})
	if r != nil {
		return r.GetStartReplay(), err
	}
	return nil, err
}

func (c *Connection) LeaveGame() error {
	_, err := c.request(&api.Request{
		Request: &api.Request_LeaveGame{
			LeaveGame: &api.RequestLeaveGame{},
		},
	})
	return err
}

func (c *Connection) QuickSave() error {
	_, err := c.request(&api.Request{
		Request: &api.Request_QuickSave{
			QuickSave: &api.RequestQuickSave{},
		},
	})
	return err
}

func (c *Connection) QuickLoad() error {
	_, err := c.request(&api.Request{
		Request: &api.Request_QuickLoad{
			QuickLoad: &api.RequestQuickLoad{},
		},
	})
	return err
}

func (c *Connection) Quit() error {
	_, err := c.request(&api.Request{
		Request: &api.Request_Quit{
			Quit: &api.RequestQuit{},
		},
	})
	return err
}

func (c *Connection) GameInfo() (*api.ResponseGameInfo, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_GameInfo{
			GameInfo: &api.RequestGameInfo{},
		},
	})
	if r != nil {
		return r.GetGameInfo(), err
	}
	return nil, err
}

func (c *Connection) Observation(observation api.RequestObservation) (*api.ResponseObservation, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_Observation{
			Observation: &observation,
		},
	})
	if r != nil {
		return r.GetObservation(), err
	}
	return nil, err
}

func (c *Connection) Action(action api.RequestAction) (*api.ResponseAction, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_Action{
			Action: &action,
		},
	})
	if r != nil {
		return r.GetAction(), err
	}
	return nil, err
}

func (c *Connection) ObsAction(obsAction api.RequestObserverAction) error {
	_, err := c.request(&api.Request{
		Request: &api.Request_ObsAction{
			ObsAction: &obsAction,
		},
	})
	return err
}

func (c *Connection) Step(step api.RequestStep) (*api.ResponseStep, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_Step{
			Step: &step,
		},
	})
	if r != nil {
		return r.GetStep(), err
	}
	return nil, err
}

func (c *Connection) Data(data api.RequestData) (*api.ResponseData, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_Data{
			Data: &data,
		},
	})
	if r != nil {
		return r.GetData(), err
	}
	return nil, err
}

func (c *Connection) Query(query api.RequestQuery) (*api.ResponseQuery, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_Query{
			Query: &query,
		},
	})
	if r != nil {
		return r.GetQuery(), err
	}
	return nil, err
}

func (c *Connection) SaveReplay() (*api.ResponseSaveReplay, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_SaveReplay{
			SaveReplay: &api.RequestSaveReplay{},
		},
	})
	if r != nil {
		return r.GetSaveReplay(), err
	}
	return nil, err
}

func (c *Connection) MapCommand(mapCommand api.RequestMapCommand) (*api.ResponseMapCommand, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_MapCommand{
			MapCommand: &mapCommand,
		},
	})
	if r != nil {
		return r.GetMapCommand(), err
	}
	return nil, err
}

func (c *Connection) ReplayInfo(replayInfo api.RequestReplayInfo) (*api.ResponseReplayInfo, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_ReplayInfo{
			ReplayInfo: &replayInfo,
		},
	})
	if r != nil {
		return r.GetReplayInfo(), err
	}
	return nil, err
}

func (c *Connection) AvailableMaps() (*api.ResponseAvailableMaps, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_AvailableMaps{
			AvailableMaps: &api.RequestAvailableMaps{},
		},
	})
	if r != nil {
		return r.GetAvailableMaps(), err
	}
	return nil, err
}

func (c *Connection) SaveMap(saveMap api.RequestSaveMap) (*api.ResponseSaveMap, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_SaveMap{
			SaveMap: &saveMap,
		},
	})
	if r != nil {
		return r.GetSaveMap(), err
	}
	return nil, err
}

func (c *Connection) Ping() (*api.ResponsePing, error) {
	r, err := c.request(&api.Request{
		Request: &api.Request_Ping{
			Ping: &api.RequestPing{},
		},
	})
	if r != nil {
		return r.GetPing(), err
	}
	return nil, err
}

func (c *Connection) Debug(debug api.RequestDebug) error {
	_, err := c.request(&api.Request{
		Request: &api.Request_Debug{
			Debug: &debug,
		},
	})
	return err
}
