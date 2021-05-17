package client

import (
	"github.com/aiseeq/s2l/protocol/api"
)

type ProcessInfo struct {
	Path string
	PID  int
	Port int
}

type PlayerSetup struct {
	*api.PlayerSetup
}

func NewParticipant(race api.Race, name string) PlayerSetup {
	return PlayerSetup{
		PlayerSetup: &api.PlayerSetup{
			Type:       api.PlayerType_Participant,
			Race:       race,
			PlayerName: name,
		},
	}
}

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

type Ports struct {
	ServerPorts *api.PortSet
	ClientPorts []*api.PortSet
	SharedPort  int32
}

func portSetIsValid(ps *api.PortSet) bool {
	return ps != nil && ps.GamePort > 0 && ps.BasePort > 0
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
