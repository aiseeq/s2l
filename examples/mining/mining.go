package main

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/lib/scl"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/client"
	"github.com/aiseeq/s2l/protocol/enums/ability"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"github.com/aiseeq/s2l/protocol/enums/zerg"
	"github.com/gonum/floats"
	"github.com/google/gxui/math"
)

var B *scl.Bot

// var CCForMiner map[api.UnitTag]api.UnitTag
var GasForMiner map[api.UnitTag]api.UnitTag
var MineralForMiner map[api.UnitTag]api.UnitTag
var TargetForMineral map[api.UnitTag]point.Point

func SplitMinerals() {
	cc := B.Units.My[terran.CommandCenter].First()
	miners := B.Units.My[terran.SCV]
	mfs := B.Units.Minerals.All().CloserThan(scl.ResourceSpreadDistance, cc)
	for _, mf := range append(mfs, mfs...) {
		miner := miners.ClosestTo(mf)
		miner.CommandTag(ability.Smart, mf.Tag)
		MineralForMiner[miner.Tag] = mf.Tag
		miners.Remove(miner)
		if miners.Empty() {
			break
		}
	}
}

func BuildSCVs() {
	cc := B.Units.My[terran.CommandCenter].First()
	if cc.IsIdle() && B.Units.My[terran.SCV].Len() < workersLimit {
		cc.Command(ability.Train_SCV)
	}
}

func CheckTime() {
	if B.Loop >= scl.TimeToLoop(2, 0) {
		log.Infof("Minerals total: %f, rate: %f, final: %d",
			B.Obs.Score.ScoreDetails.CollectedMinerals,
			B.Obs.Score.ScoreDetails.CollectionRateMinerals,
			B.Minerals)
		log.Infof("Vespene total: %f, rate: %f, final: %d",
			B.Obs.Score.ScoreDetails.CollectedVespene,
			B.Obs.Score.ScoreDetails.CollectionRateVespene,
			B.Vespene)
		if err := B.Client.LeaveGame(); err != nil {
			log.Error(err)
		}
	}
}

func InitMinerals() {
	TargetForMineral = map[api.UnitTag]point.Point{}
	cc := B.Units.My[terran.CommandCenter].First()
	mfs := B.Units.Minerals.All().CloserThan(scl.ResourceSpreadDistance, cc)
	dist := float64(mfs.First().Radius + 0.2) // + miner.Radius
	// log.Info(mfs.Len())
	for _, mf := range mfs {
		target := mf.Towards(cc, dist)
		if mf2 := mfs.CloserThan(dist, target).ClosestTo(target); mf2 != nil && mf.Tag != mf2.Tag {
			// There could be only one mineral so close
			targetMineral := point.NewCircle(float64(mf.Pos.X), float64(mf.Pos.Y), dist)
			closeMineral := point.NewCircle(float64(mf2.Pos.X), float64(mf2.Pos.Y), dist)
			// log.Info(targetMineral, closeMineral)
			pts := point.Intersect(targetMineral, closeMineral)
			if len(pts) == 2 {
				target = pts.ClosestTo(target)
				// B.DebugCircles(*targetMineral, *closeMineral)
				// B.DebugPoints(pts...)
			}
		}
		TargetForMineral[mf.Tag] = target
		// B.DebugPoints(target)
	}
	// B.DebugSend()
}

func addMinerToMineral(miner, mf *scl.Unit) {
	miner.CommandTag(ability.Smart, mf.Tag)
	MineralForMiner[miner.Tag] = mf.Tag
}

func ManageNewMiner() {
	cc := B.Units.My[terran.CommandCenter].First()
	miners := B.Units.My[terran.SCV]

	var miner *scl.Unit
	for _, scv := range miners {
		if MineralForMiner[scv.Tag] == 0 {
			miner = scv // New SCV found
			break
		}
	}
	if miner == nil {
		return
	}

	bestMfs := scl.Units{}
	mfs := B.Units.Minerals.All().CloserThan(scl.ResourceSpreadDistance, cc)
	for _, mf := range mfs {
		minersOnCrystal := 0
		for _, scv := range miners {
			if MineralForMiner[scv.Tag] == mf.Tag {
				minersOnCrystal++
			}
		}
		if minersOnCrystal == 0 {
			// We found free crystal, use it
			addMinerToMineral(miner, mf)
			return
		}
		if minersOnCrystal == 1 {
			// Non-saturated crystal
			bestMfs.Add(mf)
		}
	}
	if bestMfs.Exists() {
		// Send to closest mineral
		mf := bestMfs.ClosestTo(cc)
		addMinerToMineral(miner, mf)
		return
	}
	// All minerals are saturated
	for _, mf := range mfs {
		minersOnCrystal := 0
		for _, scv := range miners {
			if MineralForMiner[scv.Tag] == mf.Tag {
				minersOnCrystal++
			}
		}
		if minersOnCrystal == 2 {
			bestMfs.Add(mf)
		}
	}
	if bestMfs.Exists() {
		// Send to farest mineral
		mf := bestMfs.FurthestTo(cc)
		addMinerToMineral(miner, mf)
		return
	}
	log.Error("Should be unreachable")
}

func MicroMinerals() {
	cc := B.Units.My[terran.CommandCenter].First()
	for _, miner := range B.Units.My[terran.SCV] {
		mfTag := MineralForMiner[miner.Tag]
		target := TargetForMineral[mfTag]
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

func SplitGas() {
	// GasForMiner = map[api.UnitTag]api.UnitTag{}
	refs := B.Units.My[terran.Refinery]
	miners := B.Units.My[terran.SCV]
	for _, ref := range refs {
		miners.OrderByDistanceTo(ref, false)
		miners[:3].CommandTag(ability.Smart, ref.Tag)
		for _, miner := range miners[:3] {
			GasForMiner[miner.Tag] = ref.Tag
		}
		miners = miners[3:]
	}
}

func ChooseRefSaturatedLessThan(sat int, refs scl.Units, miner *scl.Unit) bool {
	refs.OrderByDistanceTo(miner, false)
nextRef:
	for _, ref := range refs {
		minersOnGas := 0
		for _, refTag := range GasForMiner {
			if refTag == ref.Tag {
				minersOnGas++
				if minersOnGas >= sat {
					continue nextRef
				}
			}
		}
		if minersOnGas < sat {
			GasForMiner[miner.Tag] = ref.Tag
			miner.CommandTag(ability.Smart, ref.Tag)
			return true
		}
	}
	return false
}

func NewScvToGas() {
	refs := B.Units.My[terran.Refinery]
	miners := B.Units.My[terran.SCV]
	for _, miner := range miners {
		if MineralForMiner[miner.Tag] != 0 {
			continue
		}
		if GasForMiner[miner.Tag] != 0 {
			continue
		}
		if !ChooseRefSaturatedLessThan(2, refs, miner) {
			ChooseRefSaturatedLessThan(3, refs, miner)
		}
	}
}

func MicroGas() {
	cc := B.Units.My[terran.CommandCenter].First()
	refs := B.Units.My[terran.Refinery]
	for _, miner := range B.Units.My[terran.SCV] {
		refTag := GasForMiner[miner.Tag]
		if refTag == 0 {
			continue
		}
		target := refs.ByTag(refTag).Towards(miner, float64(refs.First().Radius+miner.Radius))
		if !miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(1, target) && miner.IsCloserThan(2, target) {
			miner.CommandPos(ability.Move_Move, target)
			miner.CommandTagQueue(ability.Smart, refTag)
		}
		target = cc.Towards(miner, float64(cc.Radius+miner.Radius))
		if miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(1, target) && miner.IsCloserThan(2, target) {
			miner.CommandPos(ability.Move_Move, target)
			miner.CommandTagQueue(ability.Smart, cc.Tag)
		}
	}
}

func KillCsvsAndUpgradeCC() {
	for _, miner := range B.Units.My[terran.SCV] {
		B.DebugKillUnits(miner.Tag)
	}
	cc := B.Units.My[terran.CommandCenter].First()
	B.DebugKillUnits(cc.Tag)
	// B.DebugAddUnits(terran.MULE, B.Obs.PlayerCommon.PlayerId, cc.Towards(B.Locs.MapCenter, -3), 4)
	B.DebugAddUnits(terran.OrbitalCommand, B.Obs.PlayerCommon.PlayerId, point.Pt3(cc.Pos), 1)
	B.DebugSend()
}

func GiveCCEnergy() {
	log.Info("ok")
	cc := B.Units.My[terran.OrbitalCommand].First()
	B.DebugUnitValue(cc.Tag, api.DebugSetUnitValue_Energy, 200)
	B.DebugSend()
}

func CallMule() {
	cc := B.Units.My[terran.OrbitalCommand].First()
	mf := B.Units.Minerals.All().ClosestTo(cc)
	cc.CommandPosQueue(ability.Effect_CalldownMULE, mf.Towards(cc, 0.5))
}

func ManagedMule() {
	// todo
}

func SimpleMule() {
	if B.Loop == 0 {
		KillCsvsAndUpgradeCC()
	}
	if B.Loop >= 6 && B.Loop < 9 {
		GiveCCEnergy()
	}
	if B.Loop >= 9 && B.Units.My[terran.OrbitalCommand].First().Energy >= 50 {
		CallMule()
	}
	if mules := B.Units.My[terran.MULE].Filter(scl.Idle); mules.Exists() {
		mf := B.Units.Minerals.All().ClosestTo(mules.First())
		mules.CommandTag(ability.Smart, mf.Tag)
	}
	CheckTime()
}

func ManageGas() {
	if B.Loop == 0 {
		SplitMinerals()
	}
	if B.Loop == 3 {
		GasForMiner = map[api.UnitTag]api.UnitTag{}
		// SplitGas()
	}
	if B.Loop >= 3 {
		NewScvToGas()
		MicroGas()
	}

	BuildSCVs()
	CheckTime()
}

func SimpleGas() {
	if B.Loop == 0 {
		SplitMinerals()
	}
	if B.Loop == 3 {
		GasForMiner = map[api.UnitTag]api.UnitTag{}
		// SplitGas()
	}
	if B.Loop >= 3 {
		NewScvToGas()
	}

	BuildSCVs()
	CheckTime()
}

func MicroManage() {
	// time.Sleep(200 * time.Millisecond)
	if B.Loop == 0 {
		SplitMinerals()
		InitMinerals()
	}

	ManageNewMiner()
	MicroMinerals()
	BuildSCVs()
	CheckTime()
}

func SplitAndManage() {
	// time.Sleep(200 * time.Millisecond)
	if B.Loop == 0 {
		SplitMinerals()
	}

	ManageNewMiner()
	BuildSCVs()
	CheckTime()
}

func SplitAndForget() {
	if B.Loop == 0 {
		SplitMinerals()
	}

	BuildSCVs()
	CheckTime()
}

func SimpleLogic() {
	// Just add scvs, don't control them in any way
	BuildSCVs()
	CheckTime()
}

func BuildDrones() {
	larvae := B.Units.My[zerg.Larva]
	if larvae.Exists() && B.CanBuy(ability.Train_Drone) {
		larvae.First().Command(ability.Train_Drone)
	}
}

func MiningLib() {
	if B.MyRace() == api.Race_Terran {
		if B.Loop >= 102 && B.Loop < 105 {
			// B.RedistributeWorkersToRefineryIfNeeded(B.Units.My[terran.Refinery].First(), B.Units.My[terran.SCV], 3)
		}
		B.HandleMiners(B.Units.My[terran.SCV], B.Units.My[terran.CommandCenter], nil, 1, point.Pt0(), TurretsMiningPos)
		BuildSCVs()
	} else if B.MyRace() == api.Race_Zerg {
		if B.Loop >= 102 && B.Loop < 105 {
			B.RedistributeWorkersToRefineryIfNeeded(B.Units.My[zerg.Hatchery].First(), B.Units.My[zerg.Drone], 3)
		}
		B.HandleMiners(B.Units.My[zerg.Drone], B.Units.My[zerg.Hatchery], nil, 1, point.Pt0(), TurretsMiningPos)
		BuildDrones()
	}
	CheckTime()
	// B.DebugOrders()
	// B.DebugSend()
}

func Step() {
	B.Cmds = &scl.CommandsStack{}
	B.Loop = int(B.Obs.GameLoop)
	if B.Loop < B.LastLoop+B.FramesPerOrder {
		return // Skip frame repeat
	} else {
		B.LastLoop = B.Loop
	}

	B.ParseObservation()
	B.ParseUnits()
	B.ParseOrders()

	// SimpleLogic() // 1705
	// SplitAndForget() // 1745 - unlim, 1365 - lim12, 1675 - lim16
	// SplitAndManage() // 1750
	// MicroManage() // 1925 - unlim, 1510 - lim12, 1855 - lim16

	// SimpleGas() // No difference if gas is saturated
	// ManageGas()

	// SimpleMule()
	// ManagedMule()

	MiningLib()

	B.Cmds.Process(&B.Actions)
	if len(B.Actions) > 0 {
		// log.Info(B.Loop, len(B.Actions), B.Actions)
		// log.Info(B.Cmds)
		if resp, err := B.Client.Action(api.RequestAction{Actions: B.Actions}); err != nil {
			log.Error(err)
		} else {
			_ = resp.Result // todo: do something with it?
		}
		B.Actions = nil
	}
}

func AddBuildings() {
	if B.MyRace() == api.Race_Terran {
		B.DebugAddUnits(terran.SupplyDepot, B.Obs.PlayerCommon.PlayerId, B.Locs.MyStart.Towards(B.Locs.MapCenter, 3), 1)
		// B.DebugAddUnits(terran.MissileTurret, B.Obs.PlayerCommon.PlayerId, B.Locs.MyStart.Towards(B.Locs.MapCenter, -2), 1)
		/*cc := B.Units.My[terran.CommandCenter].First()
		geysers := B.Units.Geysers.All().CloserThan(10, cc)
		for _, geyser := range geysers {
			B.DebugKillUnits(geyser.Tag)
			B.DebugAddUnits(terran.Refinery, B.Obs.PlayerCommon.PlayerId, point.Pt3(geyser.Pos), 1)
		}*/
		for _, pos := range TurretsPos {
			B.DebugAddUnits(terran.MissileTurret, B.Obs.PlayerCommon.PlayerId, pos, 1)
		}
		B.DebugSend()
	} else if B.MyRace() == api.Race_Zerg {
		B.DebugAddUnits(zerg.Overlord, B.Obs.PlayerCommon.PlayerId, B.Locs.MyStart, 1)
		cc := B.Units.My[zerg.Hatchery].First()
		geysers := B.Units.Geysers.All().CloserThan(10, cc)
		for _, geyser := range geysers {
			B.DebugKillUnits(geyser.Tag)
			B.DebugAddUnits(zerg.Extractor, B.Obs.PlayerCommon.PlayerId, point.Pt3(geyser.Pos), 1)
		}
		B.DebugSend()
	}
}

var TurretsPos, TurretsMiningPos point.Points

func FindTurretPosition(ptr point.Pointer) {
	mfs := B.Units.Minerals.All().CloserThan(scl.ResourceSpreadDistance, ptr)
	if mfs.Len() != 8 {
		return // Unorthodox expansion, it's better to skip it
	}
	mfsCenter := mfs.Center()

	var corners point.Points
	ccVec := (mfsCenter - ptr.Point()).Norm()
	minSide := ccVec.Compas()
	ccDir := mfsCenter.Dir(ptr)
	geysers := B.Units.Geysers.All().CloserThan(10, ptr)
	if minSide.IsDiagonal() {
		// Minerals are in quarter-circle
		// We need to find minerals on the edge of field
		mfs.OrderByDistanceTo(mfsCenter, true)           // furthest are in corners
		corners = append(corners, mfs[0].Point()-1-0.5i) // Bottom left corner of the mineral field
		for _, mf := range mfs[1:4] {
			if mf.IsCloserThan(4, mfs[0]) {
				continue
			}
			corners = append(corners, mf.Point()-1-0.5i)
			break
		}
		// Move corners so they become turrets positions
		for n, corner := range corners {
			side := (corner - ptr.Point()).Compas()
			corners[n] += ccDir
			switch side {
			case point.N:
				corners[n] -= 1i
			case point.S:
				// nothing
			case point.E:
				corners[n] -= 1
				if imag(ccDir) == -1 {
					corners[n] -= 1i
				}
			case point.W:
				corners[n] += 1
				if imag(ccDir) == -1 {
					corners[n] -= 1i
				}
			}
		}
		// Check if geysers are close
		if len(geysers) == 2 && geysers[0].Dist2(geysers[1]) < 8*8 {
			geysersCenter := geysers.Center()
			n := 0
			if geysersCenter.Dist2(corners[1]) < geysersCenter.Dist2(corners[0]) {
				n = 1
			}
			// Here we need to move position horizontally or vertically to touch both geysers
			furthestGeyser := geysers.FurthestTo(corners[n])
			side := (corners[n] - furthestGeyser.Point()).Compas()
			switch side {
			case point.N:
				corners[n].SetY(furthestGeyser.Point().Y() + 1.5)
			case point.S:
				corners[n].SetY(furthestGeyser.Point().Y() - 3.5)
			case point.E:
				corners[n].SetX(furthestGeyser.Point().X() + 1.5)
			case point.W:
				corners[n].SetX(furthestGeyser.Point().X() - 3.5)
			}
		}
	} else {
		// Minerals are in line or half-circle
		for _, geyser := range geysers {
			corner := geyser.Point() - 1.5 - 1.5i
			side := (geyser.Point() - geysers.Center()).Compas()
			switch side {
			case point.N:
				corner -= 2i
			case point.S:
				corner += 3i
			case point.E:
				corner -= 2
			case point.W:
				corner += 3
			}
			switch minSide {
			case point.N:
				corner += 1i
			case point.E:
				corner += 1
			}
			corners = append(corners, corner)
		}
	}
	B.DebugPoints(corners...)

	var pos point.Point
	for _, corner := range corners {
		if !B.IsPosOk(corner, scl.S2x2, 0, scl.IsBuildable, scl.IsPathable, scl.IsNoCreep) {
			continue
		}
		pos = corner.CellCenter()
		if !TurretsPos.Has(pos) {
			TurretsPos.Add(pos)
		}
		for _, p := range B.GetBuildingPoints(pos, scl.S2x2) {
			B.Grid.SetBuildable(p, false)
			B.Grid.SetPathable(p, false)
		}
	}
}

const workersLimit = 16
const repeats = 10

func main() {
	log.SetConsoleLevel(log.L_debug) // L_info L_debug
	times := map[string][]float64{}
	var cfg *client.GameConfig
	for iter := 0; iter < repeats; iter++ {
		for _, mapName := range client.Maps2021season1 { // []string{"IceandChrome506"}
			// client.SetRealtime()
			if cfg == nil {
				// client.LaunchPortStart = 8268
				client.SetMap(mapName + ".SC2Map")
				bot := client.NewParticipant(api.Race_Terran, "MiningTest")
				cpu := client.NewComputer(api.Race_Protoss, api.Difficulty_Medium, api.AIBuild_RandomBuild)
				cfg = client.LaunchAndJoin(bot, cpu)
			} else {
				cfg.StartGame(mapName + ".SC2Map")
			}
			c := cfg.Client

			B = scl.New(c, nil)
			B.FramesPerOrder = 3
			B.LastLoop = -math.MaxInt
			stop := make(chan struct{})
			B.Init(stop)

			MineralForMiner = map[api.UnitTag]api.UnitTag{}
			// CCForMiner = map[api.UnitTag]api.UnitTag{}

			TurretsPos = nil
			FindTurretPosition(B.Locs.MyStart)
			for _, exp := range B.Locs.MyExps {
				FindTurretPosition(exp)
			}
			// Make positions for mining targets calculations
			TurretsMiningPos = make(point.Points, len(TurretsPos))
			copy(TurretsMiningPos, TurretsPos)
			for n := range TurretsMiningPos {
				TurretsMiningPos[n] += 0.5 + 0.5i
			}
			AddBuildings() // To prevent supply block
			B.InitMining(TurretsMiningPos)

			for B.Client.Status == api.Status_in_game {
				Step()

				if _, err := c.Step(api.RequestStep{Count: uint32(B.FramesPerOrder)}); err != nil {
					if err.Error() == "Not in a game" {
						log.Info("Game over")
						break
					}
					log.Fatal(err)
				}

				B.UpdateObservation()
			}
			stop <- struct{}{}

			times[mapName] = append(times[mapName], float64(B.Obs.Score.ScoreDetails.CollectedMinerals))
			// times[mapName] = append(times[mapName], float64(B.Obs.Score.ScoreDetails.CollectedVespene))
		}
	}
	for mapName, res := range times {
		log.Infof("%s, min: %f, max: %f, avg: %f, %v", mapName,
			floats.Min(res), floats.Max(res), floats.Sum(res)/float64(len(res)), res)
	}
}

/*
Repeat MicroManage test from lib
Submarine506,     min: 1880.0, max: 1890.0, avg: 1886.0, [1885 1890 1880 1885 1880 1890 1890 1885 1885 1890]
IceandChrome506,  min: 1880.0, max: 1890.0, avg: 1886.0, [1880 1885 1885 1890 1885 1885 1885 1890 1890 1885]
DeathAura506,     min: 1880.0, max: 1890.0, avg: 1885.0, [1880 1890 1880 1890 1890 1885 1880 1880 1890 1885]
EverDream506,     min: 1875.0, max: 1885.0, avg: 1882.0, [1885 1875 1880 1885 1885 1885 1880 1885 1885 1875]
EternalEmpire506, min: 1845.0, max: 1855.0, avg: 1853.0, [1855 1855 1855 1855 1855 1855 1850 1850 1845 1855]
PillarsofGold506, min: 1835.0, max: 1845.0, avg: 1843.0, [1845 1845 1845 1840 1845 1845 1845 1840 1845 1835]
GoldenWall506,    min: 1730.0, max: 1785.0, avg: 1757.5, [1760 1730 1785 1775 1760 1755 1740 1760 1755 1755]

Towers added
DeathAura506,     min: 1880.0, max: 1890.0, avg: 1884.5, [1880 1880 1885 1885 1880 1890 1885 1885 1885 1890]
PillarsofGold506, min: 1835.0, max: 1845.0, avg: 1842.0, [1845 1845 1840 1845 1845 1840 1845 1840 1835 1840]
EverDream506,     min: 1770.0, max: 1825.0, avg: 1798.5, [1810 1815 1810 1770 1825 1825 1780 1785 1770 1795]
IceandChrome506,  min: 1775.0, max: 1810.0, avg: 1794.5, [1775 1785 1795 1800 1810 1780 1790 1805 1795 1810]
Submarine506,     min: 1765.0, max: 1825.0, avg: 1793.5, [1815 1770 1805 1795 1790 1795 1765 1805 1770 1825]
EternalEmpire506, min: 1735.0, max: 1785.0, avg: 1759.0, [1735 1770 1750 1770 1765 1740 1765 1760 1785 1750]
GoldenWall506,    min: 1735.0, max: 1775.0, avg: 1753.5, [1760 1770 1755 1740 1740 1735 1775 1750 1745 1765]

Towers moved a little
Submarine506,     min: 1880.0, max: 1890.0, avg: 1885.0, [1890 1880 1885 1885 1885 1885 1885 1880 1885 1890]
EverDream506,     min: 1880.0, max: 1885.0, avg: 1884.5, [1885 1885 1885 1885 1885 1885 1885 1885 1885 1880]
DeathAura506,     min: 1875.0, max: 1895.0, avg: 1883.0, [1875 1885 1885 1880 1880 1880 1880 1885 1895 1885]
IceandChrome506,  min: 1870.0, max: 1890.0, avg: 1883.0, [1875 1885 1880 1890 1885 1885 1890 1870 1880 1890]
EternalEmpire506, min: 1825.0, max: 1850.0, avg: 1842.0, [1845 1830 1825 1850 1850 1840 1840 1850 1845 1845]
PillarsofGold506, min: 1840.0, max: 1845.0, avg: 1842.0, [1845 1845 1840 1840 1845 1840 1840 1840 1845 1840]
GoldenWall506,    min: 1805.0, max: 1820.0, avg: 1815.0, [1815 1820 1815 1805 1820 1815 1810 1820 1815 1815]
*/

/*
MicroManage
IceandChrome506,  min: 1870.0, max: 1890.0, avg: 1883.0, [1885 1885 1885 1890 1885 1890 1875 1875 1870 1890]
EverDream506,     min: 1870.0, max: 1890.0, avg: 1882.5, [1880 1885 1875 1890 1890 1870 1885 1870 1890 1890]
Submarine506,     min: 1865.0, max: 1890.0, avg: 1881.0, [1890 1865 1865 1880 1890 1890 1880 1880 1890 1880]
DeathAura506,     min: 1865.0, max: 1895.0, avg: 1879.0, [1885 1870 1890 1880 1875 1865 1875 1895 1875 1880]
EternalEmpire506, min: 1835.0, max: 1855.0, avg: 1844.0, [1840 1850 1840 1855 1835 1855 1835 1840 1855 1835]
PillarsofGold506, min: 1840.0, max: 1845.0, avg: 1842.0, [1840 1840 1845 1845 1840 1845 1845 1840 1840 1840]
GoldenWall506,    min: 1820.0, max: 1845.0, avg: 1834.5, [1840 1830 1830 1820 1835 1845 1845 1830 1835 1835]

SplitAndForget
Submarine506,     min: 1620.0, max: 1695.0, avg: 1675.0, [1690 1680 1620 1675 1675 1685 1660 1695 1690 1680]
GoldenWall506,    min: 1630.0, max: 1670.0, avg: 1658.0, [1665 1665 1665 1655 1660 1655 1630 1670 1650 1665]
DeathAura506,     min: 1590.0, max: 1695.0, avg: 1654.5, [1655 1660 1680 1650 1660 1620 1590 1695 1665 1670]
EverDream506,     min: 1610.0, max: 1705.0, avg: 1650.0, [1615 1610 1610 1670 1690 1610 1695 1705 1620 1675]
EternalEmpire506, min: 1590.0, max: 1670.0, avg: 1636.0, [1655 1635 1645 1590 1630 1595 1630 1660 1650 1670]
IceandChrome506,  min: 1565.0, max: 1690.0, avg: 1634.5, [1565 1690 1665 1615 1640 1650 1680 1600 1605 1635]
PillarsofGold506, min: 1595.0, max: 1620.0, avg: 1608.5, [1610 1605 1595 1620 1600 1615 1620 1605 1620 1595]

ManageGas
GoldenWall506,    min: 436.0, max: 440.0, avg: 437.2, [436 436 440 436 436 436 440 440 436 436]
PillarsofGold506, min: 436.0, max: 440.0, avg: 437.2, [436 436 440 440 436 436 440 436 436 436]
EternalEmpire506, min: 412.0, max: 428.0, avg: 424.0, [428 424 428 428 428 424 420 424 412 424]
DeathAura506,     min: 416.0, max: 424.0, avg: 421.2, [420 420 420 420 424 424 416 424 420 424]
EverDream506,     min: 416.0, max: 424.0, avg: 421.2, [420 424 424 424 420 420 416 424 416 424]
IceandChrome506,  min: 416.0, max: 424.0, avg: 419.2, [420 420 416 420 420 424 420 416 416 420]
Submarine506,     min: 412.0, max: 420.0, avg: 417.2, [416 420 420 416 416 416 412 420 420 416]

SimpleGas
GoldenWall506,    min: 420.0, max: 420.0, avg: 420.0, [420 420 420 420 420 420 420 420 420 420]
EternalEmpire506, min: 416.0, max: 416.0, avg: 416.0, [416 416 416 416 416 416 416 416 416 416]
PillarsofGold506, min: 416.0, max: 416.0, avg: 416.0, [416 416 416 416 416 416 416 416 416 416]
DeathAura506,     min: 408.0, max: 416.0, avg: 412.8, [416 416 412 408 412 408 412 412 416 416]
Submarine506,     min: 408.0, max: 412.0, avg: 410.0, [412 408 412 412 408 412 408 408 412 408]
EverDream506,     min: 408.0, max: 408.0, avg: 408.0, [408 408 408 408 408 408 408 408 408 408]
IceandChrome506,  min: 408.0, max: 408.0, avg: 408.0, [408 408 408 408 408 408 408 408 408 408]
*/
