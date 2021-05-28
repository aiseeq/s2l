package scl

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/protocol/api"
)

func TimeToLoop(minutes, seconds int) int {
	return int(float64(minutes*60+seconds) * FPS)
}

func (b *Bot) MyRace() api.Race {
	return B.Info.PlayerInfo[B.Obs.PlayerCommon.PlayerId-1].RaceActual
}

func (b *Bot) CanBuy(ability api.AbilityID) bool {
	cost, ok := b.U.AbilityCost[ability]
	if !ok {
		log.Warning("no cost for ability: ", ability)
	}
	return (cost.Minerals == 0 || b.Minerals >= cost.Minerals) &&
		(cost.Vespene == 0 || b.Vespene >= cost.Vespene) &&
		(cost.Food <= 0 || b.FoodLeft >= cost.Food)
}

func (b *Bot) DeductResources(aid api.AbilityID) {
	cost := b.U.AbilityCost[aid]
	b.Minerals -= cost.Minerals
	b.Vespene -= cost.Vespene
	if cost.Food > 0 {
		b.FoodUsed += cost.Food
		b.FoodLeft -= cost.Food
	}
}

func (b *Bot) Pending(aid api.AbilityID) int {
	return b.Units.My[b.U.AbilityUnit[aid]].Len() + b.Orders[aid]
}

func (b *Bot) PendingAliases(aid api.AbilityID) int {
	return b.Units.My.OfType(b.U.UnitAliases.For(b.U.AbilityUnit[aid])...).Len() + b.Orders[aid]
}

func (b *Bot) CanBuild(aid api.AbilityID, limit, active int) bool {
	return b.CanBuy(aid) && b.Pending(aid) < limit && b.Orders[aid] < active
}
