package main

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/scl"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/client"
	"github.com/aiseeq/s2l/protocol/enums/ability"
	"github.com/aiseeq/s2l/protocol/enums/terran"
)

var B *scl.Bot
var MinersByCC map[api.UnitTag]map[api.UnitTag]struct{}
var MinersByMineral map[api.UnitTag][]api.UnitTag
var MineralForMiner map[api.UnitTag]api.UnitTag

func MicroManage() {
	// time.Sleep(200 * time.Millisecond)
	if B.Loop == 0 {
		SplitScvs()
	}

	ManageNewMiner()
	cc := B.Units.My[terran.CommandCenter].First()
	for _, miner := range B.Units.My[terran.SCV] {
		mfTag := MineralForMiner[miner.Tag]
		mf := B.Units.Minerals.All().ByTag(mfTag) // todo: more effective?

		if mf == nil {
			log.Error("Wat?")
			continue
		}

		target := mf.Towards(cc, float64(mf.Radius+miner.Radius))
		if !miner.IsReturning() && len(miner.Orders) < 2 &&
			miner.IsFurtherThan(1, target) && miner.IsCloserThan(2, target) {
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
	if cc.IsIdle() {
		cc.Command(ability.Train_SCV)
	}
}

func CheckTime() {
	if B.Loop >= scl.TimeToLoop(2, 0) {
		log.Infof("Total minerals collected: %f", B.Obs.Score.ScoreDetails.CollectedMinerals)
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
	// SplitAndForget() // 1745
	// SplitAndManage() // 1750
	MicroManage() // 1925 - for 0.5 position error (but it needs fixing). For 1 ~= 1900

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
	B.DebugSend()
}

func main() {
	// client.SetMap("GoldenWall506.SC2Map")
	// client.SetMap("DeathAura506.SC2Map")
	// client.SetRealtime()
	bot := client.NewParticipant(api.Race_Terran, "MiningTest")
	cpu := client.NewComputer(api.Race_Protoss, api.Difficulty_Medium, api.AIBuild_RandomBuild)
	c := client.LaunchAndJoin(bot, cpu)

	B = scl.New(c, nil)
	B.FramesPerOrder = 3
	B.Init()

	MineralForMiner = map[api.UnitTag]api.UnitTag{}
	MinersByMineral = map[api.UnitTag][]api.UnitTag{}
	MinersByCC = map[api.UnitTag]map[api.UnitTag]struct{}{}

	AddSupply() // To prevent supply block

	for B.Client.Status == api.Status_in_game {
		Step()

		if _, err := c.Step(api.RequestStep{Count: uint32(B.FramesPerOrder)}); err != nil {
			if err.Error() == "Not in a game" {
				log.Info("Game over")
				return
			}
			log.Fatal(err)
		}

		B.UpdateObservation()
	}
}
