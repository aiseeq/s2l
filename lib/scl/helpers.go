package scl

import "github.com/aiseeq/s2l/protocol/api"

func TimeToLoop(minutes, seconds int) int {
	return int(float64(minutes*60+seconds) * FPS)
}

func (b *Bot) MyRace() api.Race {
	return B.Info.PlayerInfo[B.Obs.PlayerCommon.PlayerId-1].RaceActual
}
