package scl

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/enums/ability"
	"github.com/aiseeq/s2l/protocol/enums/protoss"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"github.com/aiseeq/s2l/protocol/enums/zerg"
)

func (b *Bot) InitCCMinerals(cc *Unit) {
	mfs := B.Units.Minerals.All().CloserThan(ResourceSpreadDistance, cc)
	dist := float64(mfs.First().Radius + 0.2)
	for _, mf := range mfs {
		target := mf.Towards(cc, dist)
		if mf2 := mfs.CloserThan(dist, target).ClosestTo(target); mf2 != nil && mf.Tag != mf2.Tag {
			targetMineral := point.NewCircle(float64(mf.Pos.X), float64(mf.Pos.Y), dist)
			closeMineral := point.NewCircle(float64(mf2.Pos.X), float64(mf2.Pos.Y), dist)
			pts := point.Intersect(targetMineral, closeMineral)
			if len(pts) == 2 {
				target = pts.ClosestTo(target)
			}
		}
		b.Miners.TargetForMineral[mf.Tag] = target
	}
	log.Debugf("Minerals inited for %v", cc) // Check in future, is it called correctly without repeats
}

func (b *Bot) InitMining() {
	b.Miners.CCForMiner = map[api.UnitTag]api.UnitTag{}
	b.Miners.GasForMiner = map[api.UnitTag]api.UnitTag{}
	b.Miners.MineralForMiner = map[api.UnitTag]api.UnitTag{}
	b.Miners.TargetForMineral = map[api.UnitTag]point.Point{}
	b.Miners.LastSeen = map[api.UnitTag]int{}

	cc := b.Units.My.OfType(terran.CommandCenter, terran.OrbitalCommand, terran.PlanetaryFortress,
		zerg.Hatchery, zerg.Lair, zerg.Hive, protoss.Nexus).Filter(Ready).First()
	b.InitCCMinerals(cc)
	miners := b.Units.My.OfType(terran.SCV, zerg.Drone, protoss.Probe)
	mfs := b.Units.Minerals.All().CloserThan(ResourceSpreadDistance, cc)
	for _, mf := range mfs {
		miner := miners.ClosestTo(mf)
		b.addMinerToMineral(miner, mf, cc)
		miners.Remove(miner)
		if miners.Empty() {
			break
		}
	}
	for _, miner := range miners {
		mf := mfs.ClosestTo(miner)
		b.addMinerToMineral(miner, mf, cc)
		mfs.Remove(mf)
		if mfs.Empty() {
			break
		}
	}
}

func (b *Bot) addMinerToMineral(miner, mf, cc *Unit) {
	miner.CommandTag(ability.Smart, mf.Tag)
	b.Miners.MineralForMiner[miner.Tag] = mf.Tag
	b.Miners.CCForMiner[miner.Tag] = cc.Tag
}

func (b *Bot) GetMineralsSaturation(mfs Units) map[api.UnitTag]int {
	saturation := map[api.UnitTag]int{}
	for _, mf := range mfs {
		saturation[mf.Tag] = 0
	}
	for _, mfTag := range b.Miners.MineralForMiner {
		saturation[mfTag]++
	}
	return saturation
}

func (b *Bot) FillMineralsUpTo2(miners *Units, ccs, allMfs Units) {
	if miners.Empty() {
		return
	}
	// Calculate how many workers are already on each crystal
	saturation := b.GetMineralsSaturation(allMfs)
nextMiner:
	for _, miner := range *miners {
		ccs.OrderByDistanceTo(miner, false)
		for _, cc := range ccs {
			mfs := allMfs.CloserThan(ResourceSpreadDistance, cc)
			bestMfs := Units{}
			for _, mf := range mfs {
				if saturation[mf.Tag] == 0 {
					// We found free crystal, use it
					b.addMinerToMineral(miner, mf, cc)
					miners.Remove(miner)
					continue nextMiner
				}
				if saturation[mf.Tag] == 1 {
					// Non-saturated crystal
					bestMfs.Add(mf)
				}
			}
			if bestMfs.Exists() {
				// Send to closest mineral
				mf := bestMfs.ClosestTo(cc)
				b.addMinerToMineral(miner, mf, cc)
				miners.Remove(miner)
			}
		}
	}
}

func (b *Bot) FillMineralsUpTo3(miners *Units, ccs, allMfs Units) {
	if miners.Empty() {
		return
	}
	// Calculate how many workers are already on each crystal
	saturation := b.GetMineralsSaturation(allMfs)
	for _, miner := range *miners {
		ccs.OrderByDistanceTo(miner, false)
		for _, cc := range ccs {
			mfs := allMfs.CloserThan(ResourceSpreadDistance, cc)
			bestMfs := Units{}
			for _, mf := range mfs {
				if saturation[mf.Tag] == 2 {
					// Non-oversaturated crystal
					bestMfs.Add(mf)
				}
			}
			if bestMfs.Exists() {
				// Send to farest mineral
				mf := bestMfs.FurthestTo(cc)
				b.addMinerToMineral(miner, mf, cc)
				miners.Remove(miner)
			}
		}
	}
}

func (b *Bot) GetGasSaturation(gases Units) map[api.UnitTag]int {
	saturation := map[api.UnitTag]int{}
	for _, gas := range gases {
		saturation[gas.Tag] = 0
	}
	for _, gasTag := range b.Miners.GasForMiner {
		saturation[gasTag]++
	}
	return saturation
}

func (b *Bot) addMinerToGas(miner, gas, cc *Unit) {
	miner.CommandTag(ability.Smart, gas.Tag)
	b.Miners.GasForMiner[miner.Tag] = gas.Tag
	b.Miners.CCForMiner[miner.Tag] = cc.Tag
}

func (b *Bot) FillGases(miners *Units, ccs, gases Units) {
	if miners.Empty() {
		return
	}
	// Calculate how many workers are already on each gas
	saturation := b.GetGasSaturation(gases)
nextMiner: // First, fill gases up to 2 workers
	for _, miner := range *miners {
		ccs.OrderByDistanceTo(miner, false)
		for _, cc := range ccs {
			localGases := gases.CloserThan(ResourceSpreadDistance, cc)
			for _, gas := range localGases {
				if saturation[gas.Tag] <= 1 {
					// We found free gas, use it
					b.addMinerToGas(miner, gas, cc)
					saturation[gas.Tag]++
					miners.Remove(miner)
					continue nextMiner
				}
			}
		}
	}
nextMiner2: // If someone is left, fill up to 3
	for _, miner := range *miners {
		ccs.OrderByDistanceTo(miner, false)
		for _, cc := range ccs {
			localGases := gases.CloserThan(ResourceSpreadDistance, cc)
			for _, gas := range localGases {
				if saturation[gas.Tag] == 2 {
					// We found free gas, use it
					b.addMinerToGas(miner, gas, cc)
					saturation[gas.Tag]++
					miners.Remove(miner)
					continue nextMiner2
				}
			}
		}
	}
}

func (b *Bot) MicroMinerals(miners, ccs Units) {
	for _, miner := range miners {
		mfTag := b.Miners.MineralForMiner[miner.Tag]
		if mfTag == 0 {
			continue
		}
		cc := ccs.ByTag(b.Miners.CCForMiner[miner.Tag])
		target := b.Miners.TargetForMineral[mfTag]
		if target == 0 {
			// Minerals are not inited for this CC yet
			b.InitCCMinerals(cc)
		}
		if !miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(0.75, target) && miner.IsCloserThan(2, target) {
			miner.CommandPos(ability.Move_Move, target)
			miner.CommandTagQueue(ability.Smart, mfTag)
		}
		target = cc.Towards(miner, float64(cc.Radius+miner.Radius))
		if miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(1, target) && miner.IsCloserThan(2, target) {
			miner.CommandPos(ability.Move_Move, target)
			miner.CommandTagQueue(ability.Smart, cc.Tag)
		}
	}
}

func (b *Bot) MicroGas(miners, gases, ccs Units) {
	if gases.Empty() {
		return
	}
	for _, miner := range miners {
		gasTag := b.Miners.GasForMiner[miner.Tag]
		if gasTag == 0 {
			continue
		}
		target := gases.ByTag(gasTag).Towards(miner, float64(gases.First().Radius+miner.Radius))
		if !miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(1, target) && miner.IsCloserThan(2, target) {
			miner.CommandPos(ability.Move_Move, target)
			miner.CommandTagQueue(ability.Smart, gasTag)
		}
		cc := ccs.ByTag(b.Miners.CCForMiner[miner.Tag])
		target = cc.Towards(miner, float64(cc.Radius+miner.Radius))
		if miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(1, target) && miner.IsCloserThan(2, target) {
			miner.CommandPos(ability.Move_Move, target)
			miner.CommandTagQueue(ability.Smart, cc.Tag)
		}
	}
}

func (b *Bot) HandleOversaturation(ccs, allMfs Units) {
	undersaturated := 0
	oversaturated := map[api.UnitTag]bool{}
	for _, cc := range ccs {
		mfs := allMfs.CloserThan(ResourceSpreadDistance, cc)
		for mfTag, qty := range b.GetMineralsSaturation(mfs) {
			if qty < 2 {
				undersaturated++
			} else if qty > 2 {
				oversaturated[mfTag] = true
			}
		}
	}
	if undersaturated == 0 || len(oversaturated) == 0 {
		// There is nothing we can do
		return
	}
	log.Debugf("Undersaturated: %d, oversaturated: %d", undersaturated, len(oversaturated))
	for minerTag, mfTag := range b.Miners.MineralForMiner {
		if !oversaturated[mfTag] {
			continue
		}
		// Free miner
		delete(b.Miners.MineralForMiner, minerTag)
		delete(b.Miners.CCForMiner, minerTag)
		oversaturated[mfTag] = false
		undersaturated--
		if undersaturated == 0 {
			return
		}
	}
}

// todo: minerals targets for new cc
// todo: redistribution from oversaturated on new cc
// todo: exclude dangerous zones (liberator circle, attacked bases), don't choose dangerous routes
// balance - minerals to gas gather ratio, ex: 2 => gather more vespene if it is less than minerals * 2
func (b *Bot) HandleMiners(miners Units, ccs Units, balance float64) {
	if miners.Empty() || ccs.Empty() {
		return
	}

	// Some CCs may be abandoned because of enemy attacks. Workers should leave them
	ccTags := ccs.TagsMap()
	// Some minerals may be gone
	mfs := b.Units.Minerals.All()
	mfsTags := mfs.TagsMap()
	// Gases too
	gases := b.Units.My.OfType(terran.Refinery, terran.RefineryRich, zerg.Extractor, zerg.ExtractorRich,
		protoss.Assimilator, protoss.AssimilatorRich).Filter(Ready, func(unit *Unit) bool {
		return unit.VespeneContents > 0
	})
	gasesTags := gases.TagsMap()
	if mfs.Empty() && gases.Empty() {
		return
	}

	b.HandleOversaturation(ccs, mfs) // Call it less often if it will overload cpu

	// Free workers pool
	pool := miners.Filter(func(unit *Unit) bool {
		if unit.IsIdle() || !ccTags[b.Miners.CCForMiner[unit.Tag]] {
			delete(b.Miners.MineralForMiner, unit.Tag)
			delete(b.Miners.GasForMiner, unit.Tag)
			return true
		}
		// Assigned mineral doesn't exist anymore
		if b.Miners.MineralForMiner[unit.Tag] != 0 && !mfsTags[b.Miners.MineralForMiner[unit.Tag]] {
			delete(b.Miners.MineralForMiner, unit.Tag)
			delete(b.Miners.CCForMiner, unit.Tag)
			return true
		}
		// Assigned refinery empty or doesn't exist anymore
		if b.Miners.GasForMiner[unit.Tag] != 0 && !gasesTags[b.Miners.GasForMiner[unit.Tag]] {
			delete(b.Miners.GasForMiner, unit.Tag)
			delete(b.Miners.CCForMiner, unit.Tag)
			return true
		}
		return false
	})
	// Update memory list
	for _, miner := range miners {
		b.Miners.LastSeen[miner.Tag] = b.Loop
	}
	// Check if some units were excluded from miners group
	for minerTag, loop := range b.Miners.LastSeen {
		if loop+45 > b.Loop {
			continue // Probably, still alive but maybe somewhere in refinery
		}
		// Remove unit from all lists, so his vacant place can be occupied again
		delete(b.Miners.CCForMiner, minerTag)
		delete(b.Miners.GasForMiner, minerTag)
		delete(b.Miners.MineralForMiner, minerTag)
	}

	b.MicroMinerals(miners, ccs)
	b.MicroGas(miners, gases, ccs)

	if pool.Empty() {
		return
	}

	if b.MineralsPerFrame*balance < b.VespenePerFrame {
		b.FillMineralsUpTo2(&pool, ccs, mfs)
		if balance > 0 {
			b.FillGases(&pool, ccs, gases)
		}
	} else {
		b.FillGases(&pool, ccs, gases)
		b.FillMineralsUpTo2(&pool, ccs, mfs)
	}
	b.FillMineralsUpTo3(&pool, ccs, mfs)

	// todo: something with the rest?
}

func (b *Bot) RedistributeWorkersToRefineryIfNeeded(ref *Unit, miners Units, limit int) {
	miners.OrderByDistanceTo(ref, false)
	freed := 0
	for _, miner := range miners {
		// Should mine mineral and not carrying anything
		if b.Miners.MineralForMiner[miner.Tag] != 0 && len(miner.BuffIds) == 0 {
			// Move to free pool
			delete(b.Miners.CCForMiner, miner.Tag)
			delete(b.Miners.MineralForMiner, miner.Tag)
			freed++
			if freed > limit { // todo: why we need +1?
				return
			}
		}
	}
}
