package scl

import (
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/enums/ability"
	"github.com/aiseeq/s2l/protocol/enums/protoss"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"github.com/aiseeq/s2l/protocol/enums/zerg"
)

func (b *Bot) InitMining() {
	cc := b.Units.My.OfType(terran.CommandCenter, terran.OrbitalCommand, terran.PlanetaryFortress,
		zerg.Hatchery, zerg.Lair, zerg.Hive, protoss.Nexus).Filter(Ready).First()
	miners := b.Units.My.OfType(terran.SCV, zerg.Drone, protoss.Probe)
	mfs := b.Units.Minerals.All().CloserThan(ResourceSpreadDistance, cc)
	for _, mf := range mfs {
		miner := miners.ClosestTo(mf)
		miner.CommandTag(ability.Smart, mf.Tag)
		miners.Remove(miner)
		if miners.Empty() {
			break
		}
	}
}

func (b *Bot) FillMinerals(miners *Units, ccs Units, ignoreSurplus bool) {
	for _, miner := range *miners {
		ccs.OrderByDistanceTo(miner, false)
		for _, cc := range ccs {
			surplus := cc.AssignedHarvesters - cc.IdealHarvesters
			if surplus < 0 || ignoreSurplus {
				mf := b.Units.Minerals.All().CloserThan(ResourceSpreadDistance, cc).First()
				if mf != nil {
					miner.CommandTag(ability.Smart, mf.Tag)
					miners.Remove(miner)
					cc.AssignedHarvesters++
				}
			}
		}
	}
}

func (b *Bot) FillGas(miners *Units, gases Units) {
	for _, miner := range *miners {
		gases.OrderByDistanceTo(miner, false)
		for _, gas := range gases {
			surplus := gas.AssignedHarvesters - gas.IdealHarvesters
			if surplus < 0 {
				miner.CommandTag(ability.Smart, gas.Tag)
				miners.Remove(miner)
				gas.AssignedHarvesters++
			}
		}
	}
}

// balance - minerals to gas gather ratio, ex: 2 => gather more vespene if it is less than minerals * 2
// todo: exclude dangerous zones (liberator circle, attacked bases), don't choose dangerous routes
func (b *Bot) HandleMiners(miners Units, ccs Units, balance float64) {
	pool := miners.Filter(func(unit *Unit) bool {
		return unit.IsIdle() || (!unit.IsGathering() && !unit.IsReturning())
	})
	mfs := b.Units.Minerals.All() //.Filter(Visible)
	gases := b.Units.My.OfType(terran.Refinery, zerg.Extractor, protoss.Assimilator).Filter(Ready, func(unit *Unit) bool {
		return unit.VespeneContents > 0
	})
	if ccs.Empty() || miners.Empty() || (mfs.Empty() && gases.Empty()) {
		return
	}

	for _, cc := range ccs {
		cmfs := mfs.CloserThan(ResourceSpreadDistance, cc)
		mineralMiners := miners. /*.CloserThan(ResourceSpreadDistance, cc)*/ Filter(func(unit *Unit) bool {
			if unit.IsGathering() {
				// log.Info(unit.Orders[0].Progress) - mining progress is unknown =(
				return cmfs.ByTag(unit.TargetTag()) != nil // target is one of known minerals
			}
			return !unit.IsReturning() // Don't bother returning miners
		})
		mineralMiners.OrderByDistanceTo(cc, true)
		if surplus := cc.AssignedHarvesters - cc.IdealHarvesters; surplus > 0 {
			for _, miner := range mineralMiners {
				pool.Add(miner)
				if surplus -= 1; surplus == 0 {
					break
				}
			}
		} else {
			// Balance miners by minerals
			minersTargets := map[api.UnitTag]Units{}
			// First, add all close minerals
			ccmfs := mfs.CloserThan(ResourceSpreadDistance, cc)
			for _, mf := range ccmfs {
				minersTargets[mf.Tag] = Units{}
			}
			// Count miners for each mineral
			for _, miner := range mineralMiners {
				mineralTag := miner.TargetTag()
				mt := minersTargets[mineralTag]
				mt.Add(miner)
				minersTargets[mineralTag] = mt
			}
			// Find minerals mined by 3 (or more) workers
			disbalancedMiners := Units{}
			for mineralTag, miners := range minersTargets {
				if miners.Len() > 2 {
					// Add furthest worker to redistribution list
					mf := mfs.ByTag(mineralTag)
					if mf != nil {
						disbalancedMiners.Add(miners.FurthestTo(mf))
					}
				}
			}
			if disbalancedMiners.Exists() {
				// Find minerals mined by 1 (or less) workers
				for mineralTag, miners := range minersTargets {
					if miners.Len() < 2 {
						miner := disbalancedMiners.Pop()
						miner.CommandTag(ability.Smart, mineralTag)
						if disbalancedMiners.Empty() {
							break
						}
					}
				}
			}
		}
	}

	for _, gas := range gases {
		if surplus := gas.AssignedHarvesters - gas.IdealHarvesters; surplus > 0 {
			gasMiners := miners.CloserThan(float64(gas.Radius)+3, gas).Filter(func(unit *Unit) bool {
				if unit.IsGathering() {
					return gas.Tag == unit.TargetTag()
				}
				return !unit.IsReturning() // Don't bother returning miners
			})
			for _, miner := range gasMiners {
				pool.Add(miner)
				if surplus -= 1; surplus == 0 {
					break
				}
			}
		}
	}

	// todo: rebalance miners
	if b.MineralsPerFrame*balance < b.VespenePerFrame {
		b.FillMinerals(&pool, ccs, false)
		if balance > 0 {
			b.FillGas(&pool, gases)
		}
	} else {
		b.FillGas(&pool, gases)
		b.FillMinerals(&pool, ccs, false)
	}

	// Move excess to minerals nearby
	idlePool := pool.Filter(func(unit *Unit) bool {
		return !unit.IsReturning() && !unit.IsGathering()
	})
	if idlePool.Exists() {
		b.FillMinerals(&pool, ccs, true)
	}
}
