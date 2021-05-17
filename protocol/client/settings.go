package client

import (
	"github.com/aiseeq/s2l/protocol/api"
)

// ProcessInfo ...
type ProcessInfo struct {
	Path string
	PID  int
	Port int
}

// PlayerSetup ...
type PlayerSetup struct {
	*api.PlayerSetup
	Agent
}

// NewParticipant ...
func NewParticipant(race api.Race, agent Agent, name string) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Participant,
			Race:       race,
			PlayerName: name,
		},
		Agent: agent,
	}
}

// NewComputer ...
func NewComputer(race api.Race, difficulty api.Difficulty, build api.AIBuild) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Computer,
			Race:       race,
			Difficulty: difficulty,
			AiBuild:    build,
		},
	}
}

// NewObserver ...
func NewObserver(agent Agent, name string) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Observer,
			PlayerName: name,
		},
		Agent: agent,
	}
}

// Ports ...
type Ports struct {
	ServerPorts *api.PortSet
	ClientPorts []*api.PortSet
	SharedPort  int32
}

func newPorts() Ports {
	return Ports{&api.PortSet{GamePort: -1, BasePort: -1}, []*api.PortSet{}, -1}
}

func (p Ports) isValid() bool {
	if p.SharedPort < 1 || !portSetIsValid(p.ServerPorts) || len(p.ClientPorts) < 1 {
		return false
	}

	for _, ps := range p.ClientPorts {
		if !portSetIsValid(ps) {
			return false
		}
	}

	return true
}

func portSetIsValid(ps *api.PortSet) bool {
	return ps != nil && ps.GamePort > 0 && ps.BasePort > 0
}
