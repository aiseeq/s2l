package main

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/lib/scl"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/client"
	"github.com/aiseeq/s2l/protocol/enums/ability"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"github.com/gonum/floats"
	"github.com/google/gxui/math"
)

var B *scl.Bot
var MinersByCC map[api.UnitTag]map[api.UnitTag]struct{}
var MinersByMineral map[api.UnitTag][]api.UnitTag
var MineralForMiner map[api.UnitTag]api.UnitTag
var TargetForMineral map[api.UnitTag]point.Point

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

func MicroManage() {
	// time.Sleep(200 * time.Millisecond)
	if B.Loop == 0 {
		SplitScvs()
		InitMinerals()
	}

	ManageNewMiner()
	cc := B.Units.My[terran.CommandCenter].First()
	for _, miner := range B.Units.My[terran.SCV] {
		mfTag := MineralForMiner[miner.Tag]
		target := TargetForMineral[mfTag] // todo: check if still exist
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

	BuildSCVs()
	CheckTime()
}

func addMinerToMineral(miner, mf, cc *scl.Unit) {
	miner.CommandTag(ability.Smart, mf.Tag)
	MinersByCC[cc.Tag][miner.Tag] = struct{}{}
	MinersByMineral[mf.Tag] = append(MinersByMineral[mf.Tag], miner.Tag)
	MineralForMiner[miner.Tag] = mf.Tag
}

func ManageNewMiner() {
	cc := B.Units.My[terran.CommandCenter].First()
	miners := B.Units.My[terran.SCV]
	if len(miners) > len(MinersByCC[cc.Tag]) {
		// New SCV found
		var miner *scl.Unit
		for _, scv := range miners {
			if _, ok := MinersByCC[cc.Tag][scv.Tag]; !ok {
				miner = scv
			}
		}
		if miner == nil {
			log.Error("Wat?")
			return
		}

		bestMfs := scl.Units{}
		mfs := B.Units.Minerals.All().CloserThan(scl.ResourceSpreadDistance, cc)
		for _, mf := range mfs {
			if len(MinersByMineral[mf.Tag]) == 0 {
				// We found free crystal, use it
				addMinerToMineral(miner, mf, cc)
				return
			}
			if len(MinersByMineral[mf.Tag]) == 1 {
				// Non-saturated crystal
				bestMfs.Add(mf)
			}
		}
		if bestMfs.Exists() {
			// Send to closest mineral
			mf := bestMfs.ClosestTo(cc)
			addMinerToMineral(miner, mf, cc)
			return
		}
		// All minerals are saturated
		for _, mf := range mfs {
			if len(MinersByMineral[mf.Tag]) == 2 {
				bestMfs.Add(mf)
			}
		}
		if bestMfs.Exists() {
			// Send to farest mineral
			mf := bestMfs.FurthestTo(cc)
			addMinerToMineral(miner, mf, cc)
			return
		}
		log.Error("Should be unreachable")
	}
}

func SplitAndManage() {
	// time.Sleep(200 * time.Millisecond)
	if B.Loop == 0 {
		SplitScvs()
	}

	ManageNewMiner()
	BuildSCVs()
	CheckTime()
}

func SplitScvs() {
	cc := B.Units.My[terran.CommandCenter].First()
	MinersByCC[cc.Tag] = map[api.UnitTag]struct{}{}

	miners := B.Units.My[terran.SCV]
	mfs := B.Units.Minerals.All().CloserThan(scl.ResourceSpreadDistance, cc)
	for _, mf := range append(mfs, mfs...) {
		miner := miners.ClosestTo(mf)
		miner.CommandTag(ability.Smart, mf.Tag)
		MinersByCC[cc.Tag][miner.Tag] = struct{}{}
		MinersByMineral[mf.Tag] = append(MinersByMineral[mf.Tag], miner.Tag)
		MineralForMiner[miner.Tag] = mf.Tag
		miners.Remove(miner)
		if miners.Empty() {
			break
		}
	}
}

func SplitAndForget() {
	if B.Loop == 0 {
		SplitScvs()
	}

	BuildSCVs()
	CheckTime()
}

func BuildSCVs() {
	cc := B.Units.My[terran.CommandCenter].First()
	if cc.IsIdle() && B.Units.My[terran.SCV].Len() < workersLimit {
		cc.Command(ability.Train_SCV)
	}
}

func CheckTime() {
	if B.Loop >= scl.TimeToLoop(2, 0) {
		log.Infof("Total: %f, rate: %f, final: %d",
			B.Obs.Score.ScoreDetails.CollectedMinerals,
			B.Obs.Score.ScoreDetails.CollectionRateMinerals,
			B.Minerals)
		if err := B.Client.LeaveGame(); err != nil {
			log.Error(err)
		}
	}
}

func SimpleLogic() {
	// Just add scvs, don't control them in any way
	BuildSCVs()
	CheckTime()
}

func Step() {
	B.Cmds = &scl.CommandsStack{}
	B.Loop = int(B.Obs.GameLoop)
	if B.Loop != 0 && B.Loop < B.LastLoop+B.FramesPerOrder {
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
	MicroManage() // 1925 - unlim, 1510 - lim12, 1855 - lim16

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

func AddSupply() {
	B.DebugAddUnits(terran.SupplyDepot, B.Obs.PlayerCommon.PlayerId, B.Locs.MyStart.Towards(B.Locs.MapCenter, 3), 1)
	// B.DebugAddUnits(terran.MissileTurret, B.Obs.PlayerCommon.PlayerId, B.Locs.MyStart.Towards(B.Locs.MapCenter, -2), 1)
	B.DebugSend()
}

const workersLimit = 16
const repeats = 10

func main() {
	times := map[string][]float64{}
	var cfg *client.GameConfig
	for iter := 0; iter < repeats; iter++ {
		for _, mapName := range client.Maps2021season1 { // []string{"IceandChrome506"}
			// client.SetRealtime()
			if cfg == nil {
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
			B.Init(false) // we don't need to renew paths here

			MineralForMiner = map[api.UnitTag]api.UnitTag{}
			MinersByMineral = map[api.UnitTag][]api.UnitTag{}
			MinersByCC = map[api.UnitTag]map[api.UnitTag]struct{}{}

			AddSupply() // To prevent supply block

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
			times[mapName] = append(times[mapName], float64(B.Obs.Score.ScoreDetails.CollectedMinerals))
		}
	}
	for mapName, res := range times {
		log.Infof("%s, min: %f, max: %f, avg: %f, %v", mapName,
			floats.Min(res), floats.Max(res), floats.Sum(res)/float64(len(res)), res)
	}
}
