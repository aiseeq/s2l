package main

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/scl"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/client"
	"github.com/aiseeq/s2l/protocol/enums/protoss"
	"github.com/google/gxui/math"
)

var B *scl.Bot

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

	log.Info(B.Units.Enemy.All())
}

func AddDebug() {
	myId := B.Obs.PlayerCommon.PlayerId
	enemyId := 3 - myId
	B.DebugAddUnits(protoss.DarkTemplar, enemyId, B.Locs.MyStart.Towards(B.Locs.MapCenter, 3), 1)
	B.DebugAddUnits(protoss.Zealot, enemyId, B.Locs.MyStart.Towards(B.Locs.MapCenter, 3), 1)
	B.DebugAddUnits(protoss.Observer, enemyId, B.Locs.MyStart.Towards(B.Locs.MapCenter, 3), 1)
	// [display_type:Hidden alliance:Enemy tag:4347920385 unit_type:76 pos:<x:39.96875 y:38.91944 z:11.988729 >
	// radius:0.375 cloak:Cloaked is_on_screen:true || display_type:Visible alliance:Enemy tag:4348182529 unit_type:73
	// owner:2 pos:<x:37.189568 y:39.881966 z:11.988839 > facing:6.009264 radius:0.5 build_progress:1 cloak:NotCloaked
	// is_on_screen:true is_active:true health:100 health_max:100 shield:50 shield_max:50 || display_type:Hidden
	// alliance:Enemy tag:4348444673 unit_type:82 pos:<x:35.66594 y:43.079185 z:15.738281 > radius:0.5
	// cloak:Cloaked is_on_screen:true is_flying:true ]
	B.DebugAddUnits(protoss.Observer, myId, B.Locs.MyStart.Towards(B.Locs.MapCenter, 3), 1)
	// [display_type:Visible alliance:Enemy tag:4354736129 unit_type:82 owner:2 pos:<x:39.307133 y:37.307133
	// z:15.738281 > facing:0.7356 radius:0.5 build_progress:1 cloak:CloakedDetected detect_range:11 is_on_screen:true
	// is_active:true health:40 health_max:40 shield:20 shield_max:20 is_flying:true || display_type:Visible
	// alliance:Enemy tag:4354473985 unit_type:73 owner:2 pos:<x:33.59772 y:41.81865 z:11.989014 > facing:1.7741566
	// radius:0.5 build_progress:1 cloak:NotCloaked is_active:true health:100 health_max:100 shield:50 shield_max:50 ||
	// display_type:Visible alliance:Enemy tag:4354211841 unit_type:76 owner:2 pos:<x:34.279373 y:41.37503 z:11.989014 >
	// facing:1.7582874 radius:0.375 build_progress:1 cloak:CloakedDetected is_active:true health:40 health_max:40
	// shield:80 shield_max:80 ]
	B.DebugSend()
}

func main() {
	log.SetConsoleLevel(log.L_debug) // L_info L_debug
	client.SetRealtime()
	bot := client.NewParticipant(api.Race_Terran, "MiningTest")
	cpu := client.NewComputer(api.Race_Protoss, api.Difficulty_Medium, api.AIBuild_RandomBuild)
	cfg := client.LaunchAndJoin(bot, cpu)
	c := cfg.Client

	B = scl.New(c, nil)
	B.FramesPerOrder = 3
	B.LastLoop = -math.MaxInt
	B.Init(false) // we don't need to renew paths here

	AddDebug()

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
}
