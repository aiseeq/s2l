package scl

import (
	"github.com/aiseeq/s2l/lib/grid"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/enums/ability"
	"github.com/aiseeq/s2l/protocol/enums/effect"
	"github.com/aiseeq/s2l/protocol/enums/neutral"
	"github.com/aiseeq/s2l/protocol/enums/protoss"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"github.com/aiseeq/s2l/protocol/enums/zerg"
	"math"
)

type Unit struct {
	api.Unit
	Bot             *Bot
	SpamCmds        bool
	HPS             float64
	Hits            float64
	HitsMax         float64
	HitsLost        float64
	LastMaxCooldown float64
	Abilities       []api.AbilityID
	TrueAbilities   []api.AbilityID
	PosDelta        point.Point
	Neighbours      Units
	Cluster         *Cluster
}

type Cost struct {
	Minerals int
	Vespene  int
	Food     int
	Time     int
}

type Weapon struct {
	ground, air             *api.Weapon
	groundDps, airDps       float64
	groundDamage, airDamage float64
}

type UnitOrder struct {
	Loop    int
	Ability api.AbilityID
	Pos     point.Point
	Tag     api.UnitTag
}

type UnitTypes []api.UnitTypeID
type Aliases map[api.UnitTypeID]UnitTypes

var Types []*api.UnitTypeData
var GroundAttackCircle = map[api.UnitTypeID]point.Points{}
var Upgrades []*api.UpgradeData
var Effects []*api.EffectData
var UnitCost = map[api.UnitTypeID]Cost{}
var AbilityCost = map[api.AbilityID]Cost{}
var AbilityUnit = map[api.AbilityID]api.UnitTypeID{}
var UnitAbility = map[api.UnitTypeID]api.AbilityID{}
var UnitAliases = Aliases{}
var UnitsOrders = map[api.UnitTag]UnitOrder{}

var attributes = map[api.UnitTypeID]map[api.Attribute]bool{}
var weapons = map[api.UnitTypeID]Weapon{}
var hitsHistory = map[api.UnitTag][]int{}
var prevUnits = map[api.UnitTag]*Unit{}

func InitUnits(typeData []*api.UnitTypeData) {
	Types = typeData

	for _, td := range Types {
		attributes[td.UnitId] = map[api.Attribute]bool{}
		for _, attribute := range td.Attributes {
			attributes[td.UnitId][attribute] = true
		}
		cost := Cost{
			Minerals: int(td.MineralCost),
			Vespene:  int(td.VespeneCost),
			Food:     int(td.FoodRequired - td.FoodProvided),
			Time:     int(td.BuildTime),
		}
		// fixes
		if td.Race == api.Race_Zerg && attributes[td.UnitId][api.Attribute_Structure] {
			cost.Minerals -= 50 // Why there is drone cost in buildings cost?
		}
		if td.AbilityId == ability.Train_Zergling {
			cost.Minerals = 50
			cost.Food = 1
		}
		if td.AbilityId == ability.Morph_OrbitalCommand || td.AbilityId == ability.Morph_PlanetaryFortress {
			cost.Minerals -= 400 // Deduct CC price
		}
		if td.UnitId == terran.Bunker {
			weapon := *Types[terran.Marine].Weapons[0] // Make copy
			td.Weapons = append(td.Weapons, &weapon)
			td.Weapons[0].Attacks *= 4 // 4 marines
			td.Weapons[0].Range++      // Bunker range boost
		}
		if td.UnitId == terran.Battlecruiser && len(td.Weapons) == 0 { // No weapons defined
			td.Weapons = []*api.Weapon{
				{
					Type:    api.Weapon_Air,
					Damage:  5,
					Attacks: 1,
					Range:   6,
					Speed:   0.224,
				},
				{
					Type:    api.Weapon_Ground,
					Damage:  8,
					Attacks: 1,
					Range:   6,
					Speed:   0.224,
				},
			}
		}
		UnitCost[td.UnitId] = cost
		AbilityCost[td.AbilityId] = cost
		AbilityUnit[td.AbilityId] = td.UnitId
		UnitAbility[td.UnitId] = td.AbilityId
		w := Weapon{}
		for _, weapon := range td.Weapons {
			if weapon.Type == api.Weapon_Ground || weapon.Type == api.Weapon_Any {
				w.ground = weapon
				w.groundDamage = float64(weapon.Damage * float32(weapon.Attacks))
				w.groundDps = w.groundDamage / float64(weapon.Speed)

				// No ground weapons radius for liberators. Evading via effects
				if td.UnitId == terran.Liberator || td.UnitId == terran.LiberatorAG {
					w.ground.Range = -2
				}
			}
			if weapon.Type == api.Weapon_Air || weapon.Type == api.Weapon_Any {
				w.air = weapon
				w.airDamage = float64(weapon.Damage * float32(weapon.Attacks))
				w.airDps = w.airDamage / float64(weapon.Speed)
			}
			weapons[td.UnitId] = w
		}
		UnitAliases.Add(td)

		// find cells of ground attack circles for units
		if weapons[td.UnitId].ground != nil {
			r := float64(weapons[td.UnitId].ground.Range)
			r += 2 // Max unit radius + max target radius
			r2 := r * r
			ps := point.Points{}
			// Count from center of the cell where unit is
			for y := -math.Ceil(r); y <= math.Ceil(r); y++ {
				for x := -math.Ceil(r); x <= math.Ceil(r); x++ {
					x2 := (x) * (x)
					y2 := (y) * (y)
					if x2+y2 <= r2 {
						ps.Add(point.Pt(x, y))
					}
				}
			}
			GroundAttackCircle[td.UnitId] = ps
		}
	}
}

func InitUpgrades(upgradeData []*api.UpgradeData) {
	Upgrades = upgradeData

	for _, ud := range Upgrades {
		cost := Cost{
			Minerals: int(ud.MineralCost),
			Vespene:  int(ud.VespeneCost),
			Food:     0,
			Time:     int(ud.ResearchTime),
		}
		AbilityCost[ud.AbilityId] = cost
		// api bug workaroubd: TerranVehicleArmorsLevel1 -> Research_TerranVehicleAndShipPlatingLevel1
		if ud.AbilityId == 852 {
			AbilityCost[864] = cost
		}
		if ud.AbilityId == 853 {
			AbilityCost[865] = cost
		}
		if ud.AbilityId == 854 {
			AbilityCost[866] = cost
		}
		// log.Info(ud)
	}
}

func InitEffects(effectData []*api.EffectData) {
	Effects = effectData
}

func (b *Bot) NewUnit(unit *api.Unit) (*Unit, bool) {
	u := &Unit{
		Unit:    *unit,
		Bot:     b,
		Hits:    float64(unit.Health + unit.Shield),
		HitsMax: float64(unit.HealthMax + unit.ShieldMax),
	}
	if u.Alliance == api.Alliance_Neutral || u.DisplayType == api.DisplayType_Snapshot {
		return u, false
	}

	// Check saved orders, because order itself is not in observation yet if u.Bot.FramesPerOrder not passed
	order, ok := UnitsOrders[u.Tag]
	if ok && order.Loop+u.Bot.FramesPerOrder > u.Bot.Loop {
		uo := api.UnitOrder{AbilityId: order.Ability}
		if order.Pos != 0 {
			uo.Target = &api.UnitOrder_TargetWorldSpacePos{TargetWorldSpacePos: order.Pos.To3D()}
		}
		if order.Tag != 0 {
			uo.Target = &api.UnitOrder_TargetUnitTag{TargetUnitTag: order.Tag}
		}
		u.Orders = []*api.UnitOrder{&uo} // append(u.Orders, &uo)
	}

	isNew := true
	pu, ok := prevUnits[u.Tag]
	if ok {
		isNew = false
	} else {
		pu = u
	}
	hits := hitsHistory[u.Tag]

	if len(hits) > 0 && hits[0] < b.Loop-HitHistoryLoops {
		hits = hits[2:]
	}
	if u.Hits < pu.Hits {
		// Received damage
		u.HitsLost = pu.Hits - u.Hits
		hits = append(hits, b.Loop, int(u.HitsLost))
	}
	if len(hits) > 0 {
		for x, hit := range hits {
			if x%2 == 0 {
				continue // Skip time
			}
			u.HPS += float64(hit)
		}
		u.HPS /= float64(len(hits) / 2)
	}

	u.PosDelta = pu.Point() - u.Point()
	if u.WeaponCooldown == 0 {
		u.LastMaxCooldown = 0
	} else if float64(u.WeaponCooldown) > u.LastMaxCooldown {
		u.LastMaxCooldown = float64(u.WeaponCooldown)
	}

	hitsHistory[u.Tag] = hits
	prevUnits[u.Tag] = u
	return u, isNew
}

func (u *Unit) Point() point.Point {
	return point.Pt3(u.Pos)
}

func (u *Unit) Dist(ptr point.Pointer) float64 {
	return u.Point().Dist(ptr)
}

func (u *Unit) Dist2(ptr point.Pointer) float64 {
	return u.Point().Dist2(ptr)
}

func (u *Unit) Towards(ptr point.Pointer, offset float64) point.Point {
	return u.Point().Towards(ptr, offset)
}

func (u *Unit) GetWayMap(safe bool) (*grid.Grid, WaypointsMap) {
	var navGrid *grid.Grid
	var waymap WaypointsMap
	if safe {
		navGrid = u.Bot.SafeGrid
		waymap = u.Bot.SafeWayMap
		if u.UnitType == terran.Reaper && u.Bot.ReaperGrid != nil && u.Bot.ReaperWayMap != nil {
			navGrid = u.Bot.ReaperSafeGrid
			waymap = u.Bot.ReaperSafeWayMap
		}
	} else {
		navGrid = u.Bot.Grid
		waymap = u.Bot.WayMap
		if u.UnitType == terran.Reaper && u.Bot.ReaperGrid != nil && u.Bot.ReaperWayMap != nil {
			navGrid = u.Bot.ReaperGrid
			waymap = u.Bot.ReaperWayMap
		}
	}
	return navGrid, waymap
}

func (u *Unit) GroundTowards(ptr point.Pointer, offset float64, safe bool) point.Point {
	// slow, todo: use something else
	navGrid, waymap := u.GetWayMap(safe)
	path, _ := NavPath(navGrid, waymap, u, ptr)
	if path.Len() > 1 {
		return u.Towards(path[1], offset)
	}
	return 0
}

func (u *Unit) Is(ids ...api.UnitTypeID) bool {
	for _, id := range ids {
		if u.UnitType == id {
			return true
		}
	}
	return false
}

func (u *Unit) IsNot(ids ...api.UnitTypeID) bool {
	return !u.Is(ids...)
}

func (u *Unit) IsIdle() bool {
	return len(u.Orders) == 0
}

func (u *Unit) IsUnused() bool {
	if u.AddOnTag == 0 {
		return len(u.Orders) == 0
	}
	reactor := u.Bot.Units.My.OfType(UnitAliases.For(terran.Reactor)...).ByTag(u.AddOnTag)
	if reactor != nil && reactor.IsReady() {
		return len(u.Orders) < 2
	}
	return len(u.Orders) == 0
}

func (u *Unit) IsMoving() bool {
	return len(u.Orders) > 0 && u.Orders[0].AbilityId == ability.Move
}

func (u *Unit) IsCool() bool {
	return AttackDelay.UnitIsCool(u)
}

func (u *Unit) IsHalfCool() bool {
	return float64(u.WeaponCooldown) <= u.LastMaxCooldown/2
}

func (u *Unit) IsVisible() bool {
	return u.DisplayType == api.DisplayType_Visible
}

func (u *Unit) IsPosVisible() bool {
	return u.Bot.Grid.IsVisible(u)
}

var GatheringAbilities = map[api.AbilityID]bool{
	ability.Harvest_Gather_SCV:   true,
	ability.Harvest_Gather_Mule:  true,
	ability.Harvest_Gather_Drone: true,
	ability.Harvest_Gather_Probe: true,
}

func (u *Unit) IsGathering() bool {
	return len(u.Orders) > 0 && GatheringAbilities[u.Orders[0].AbilityId]
}

var ReturningAbilities = map[api.AbilityID]bool{
	ability.Harvest_Return_SCV:   true,
	ability.Harvest_Return_Mule:  true,
	ability.Harvest_Return_Drone: true,
	ability.Harvest_Return_Probe: true,
}

func (u *Unit) IsReturning() bool {
	return len(u.Orders) > 0 && ReturningAbilities[u.Orders[0].AbilityId]
}

var MineralTypes = map[api.UnitTypeID]bool{
	neutral.MineralField: true, neutral.MineralField750: true,
	neutral.RichMineralField: true, neutral.RichMineralField750: true,
	neutral.PurifierMineralField: true, neutral.PurifierMineralField750: true,
	neutral.PurifierRichMineralField: true, neutral.PurifierRichMineralField750: true,
	neutral.BattleStationMineralField: true, neutral.BattleStationMineralField750: true,
	neutral.LabMineralField: true, neutral.LabMineralField750: true,
}

func (u *Unit) IsMineral() bool {
	return MineralTypes[u.UnitType]
}

var GeyserTypes = map[api.UnitTypeID]bool{
	neutral.ProtossVespeneGeyser: true, neutral.PurifierVespeneGeyser: true,
	neutral.RichVespeneGeyser: true, neutral.ShakurasVespeneGeyser: true,
	neutral.SpacePlatformGeyser: true, neutral.VespeneGeyser: true,
}

func (u *Unit) IsGeyser() bool {
	return GeyserTypes[u.UnitType]
}

func (u *Unit) IsReady() bool {
	return u.BuildProgress == 1
}

func (u *Unit) IsStructure() bool {
	return attributes[u.UnitType][api.Attribute_Structure]
}

func (u *Unit) IsArmored() bool {
	return attributes[u.UnitType][api.Attribute_Armored]
}

func (u *Unit) IsLight() bool {
	return attributes[u.UnitType][api.Attribute_Light]
}

func (u *Unit) IsWorker() bool {
	return u.UnitType == terran.SCV || u.UnitType == terran.MULE ||
		u.UnitType == zerg.Drone || u.UnitType == protoss.Probe
}

func (u *Unit) IsDefensive() bool {
	return u.UnitType == terran.Bunker || u.UnitType == terran.MissileTurret || u.UnitType == terran.AutoTurret ||
		u.UnitType == terran.PlanetaryFortress || u.UnitType == zerg.SpineCrawler || u.UnitType == zerg.SporeCrawler ||
		u.UnitType == protoss.PhotonCannon
}

func (u *Unit) HasBuff(b api.BuffID) bool {
	for _, buff := range u.BuffIds {
		if buff == b {
			return true
		}
	}
	return false
}

func (u *Unit) HasAbility(a api.AbilityID) bool {
	for _, abil := range u.Abilities {
		if abil == a {
			return true
		}
	}
	return false
}

func (u *Unit) HasTrueAbility(a api.AbilityID) bool {
	for _, abil := range u.TrueAbilities {
		if abil == a {
			return true
		}
	}
	return false
}

func (u *Unit) HasTechlab() bool {
	if u.AddOnTag != 0 {
		tl := u.Bot.Units.My.OfType(UnitAliases.For(terran.TechLab)...).ByTag(u.AddOnTag)
		if tl != nil && tl.IsReady() {
			return true
		}
	}
	return false
}

func (u *Unit) HasReactor() bool {
	if u.AddOnTag != 0 {
		tl := u.Bot.Units.My.OfType(UnitAliases.For(terran.Reactor)...).ByTag(u.AddOnTag)
		if tl != nil && tl.IsReady() {
			return true
		}
	}
	return false
}

func (u *Unit) Speed() float64 {
	return float64(Types[u.UnitType].MovementSpeed)
}

func (u *Unit) GroundDPS() float64 {
	return weapons[u.UnitType].groundDps
}

func (u *Unit) AirDPS() float64 {
	return weapons[u.UnitType].airDps
}

func (u *Unit) GroundDamage() float64 {
	return weapons[u.UnitType].groundDamage
}

func (u *Unit) AirDamage() float64 {
	return weapons[u.UnitType].airDamage
}

func (u *Unit) IsArmed() bool {
	return u.GroundDamage() > 0 || u.AirDamage() > 0
}

func (u *Unit) GroundRange() float64 {
	if weapon := weapons[u.UnitType].ground; weapon != nil {
		return float64(weapon.Range)
	}
	return -1
}

func (u *Unit) AirRange() float64 {
	if weapon := weapons[u.UnitType].air; weapon != nil {
		return float64(weapon.Range)
	}
	return -1
}

func (u *Unit) SightRange() float64 {
	return float64(Types[u.UnitType].SightRange)
}

func (u *Unit) RangeDelta(target *Unit, gap float64) float64 {
	unitRange := -100.0
	if u.GroundDPS() > 0 && !target.IsFlying {
		unitRange = u.GroundRange()
	}
	// Air range is always larger than ground
	if u.AirDPS() > 0 && target.IsFlying {
		unitRange = u.AirRange()
	}

	// todo: remove after unit upgrades analysis will be done
	if u.Alliance == api.Alliance_Enemy {
		if u.UnitType == zerg.Hydralisk || u.UnitType == terran.PlanetaryFortress ||
			u.UnitType == terran.MissileTurret || u.UnitType == terran.AutoTurret {
			unitRange += 1
		}
		if u.UnitType == protoss.Phoenix || u.UnitType == protoss.Colossus {
			unitRange += 2
		}
	}

	dist := u.Dist(target)
	return dist - gap - float64(u.Radius+target.Radius) - unitRange
}

func (u *Unit) InRange(target *Unit, gap float64) bool {
	return u.RangeDelta(target, gap) <= 0
}

func (u *Unit) InRangeOf(us Units, gap float64) Units {
	return us.CanAttack(u, gap)
}

func (u *Unit) CanAttack(us Units, gap float64) Units {
	return us.InRangeOf(u, gap)
}

func (u *Unit) AirEvade(enemies Units, gap float64, ptr point.Pointer) (point.Point, bool) { // bool = is safe
	pos := ptr.Point()
	if enemies.Empty() {
		return pos, true // Unit can just move to desired position
	}

	// Copy of unit
	cu := *u
	delta := pos - cu.Point()
	if delta.Len() > 1 {
		delta = delta.Norm()
	}
	// Move it 1 cell to the new desirable position
	cu.Pos = (cu.Point() + delta).To3D()
	// Enemy with largest range overlap
	hazard := enemies.Min(func(unit *Unit) float64 {
		return unit.RangeDelta(&cu, gap)
	})
	outrange := hazard.RangeDelta(&cu, gap)
	// No one can reach our unit
	if outrange >= 0 {
		return pos, true // Unit can just move to desired position
	}
	// Move to enemy range border
	if outrange > -1 {
		rangeVec := (cu.Point() - hazard.Point()).Norm().Mul(-outrange)
		tangVec := (rangeVec * 1i).Mul(math.Sqrt(1 - outrange*outrange))
		p1 := cu.Point() + rangeVec + tangVec
		p2 := cu.Point() + rangeVec - tangVec
		if p1.Dist2(u) > p2.Dist2(u) {
			return u.Point() + (p1 - cu.Point()).Norm().Mul(airSpeedBoostRange), false
		}
		return u.Point() + (p2 - cu.Point()).Norm().Mul(airSpeedBoostRange), false
	}
	// Move directly from enemy
	escVec := (pos - hazard.Point()).Norm().Mul(airSpeedBoostRange)
	return u.Point() + escVec, false
}

func (u *Unit) GroundEvade(enemies Units, gap float64, ptr point.Pointer) (point.Point, bool) { // bool = is safe
	pos := ptr.Point()
	if enemies.Empty() {
		return pos, true // Unit can just move to desired position
	}

	delta := u.PosDelta.Norm()
	if u.PosDelta == 0 {
		delta = (enemies.Center() - u.Point()).Norm()
	}

	// Copy of unit
	cu := *u
	// Move it 1 cell further to the new position
	cu.Pos = (cu.Point() + delta).To3D()
	// Enemy with largest range overlap
	hazard := enemies.Min(func(unit *Unit) float64 {
		return unit.RangeDelta(&cu, gap)
	})
	outrange := hazard.RangeDelta(&cu, gap)
	// No one can reach our unit
	if outrange >= 0 {
		return pos, true // Unit can just move to desired position
	}
	// Move to enemy range border
	var escVec point.Point
	if outrange > -1 {
		rangeVec := (cu.Point() - hazard.Point()).Norm().Mul(-outrange)
		tangVec := (rangeVec * 1i).Mul(math.Sqrt(1 - outrange*outrange))
		p1 := cu.Point() + rangeVec + tangVec
		p2 := cu.Point() + rangeVec - tangVec
		if p1.Dist2(u) > p2.Dist2(u) {
			escVec = (p1 - cu.Point()).Norm().Mul(airSpeedBoostRange)
		} else {
			escVec = (p2 - cu.Point()).Norm().Mul(airSpeedBoostRange)
		}
	} else {
		// Move directly from enemy
		escVec = (pos - hazard.Point()).Norm().Mul(airSpeedBoostRange)
	}
	if !u.Bot.Grid.IsPathable(u.Point() + escVec) {
		for x := 1.0; x < 4; x++ {
			esc1 := u.Point() + escVec.Rotate(math.Pi*2.0/16.0*x)
			if u.Bot.Grid.IsPathable(esc1) {
				return esc1, false
			}
			esc2 := u.Point() + escVec.Rotate(-math.Pi*2.0/16.0*x)
			if u.Bot.Grid.IsPathable(esc2) {
				return esc2, false
			}
		}
		return u.Bot.Locs.MyStart, false // Try to go home
	}
	return u.Point() + escVec, false
}

/*func (u *Unit) GroundFallbackPos(enemies Units, gap float64, safePath Steps, dist int) (point.Point, bool) { // bool = is safe
	safePos := safePath.Follow(u, dist)
	if safePos == 0 {
		safePos = u.Bot.Locs.MyStart
	}
	if enemies.Empty() {
		return safePos, true
	}

	// Copy of unit
	cu := *u
	// Move it to the new position
	cu.Pos = safePos.To3D()
	// escVec := (safePos - p).Norm()
	score := 0.0
	for _, e := range enemies {
		// RangeDelta < 0 if unit in range
		rd := e.RangeDelta(&cu, gap)
		if rd < 0 {
			score += e.GroundDPS() * (1 - rd/(e.GroundRange()+float64(e.Radius+u.Radius)))
		}
	}
	fbp := safePos

	var prevPoint point.Point
	for x := 0.0; x < 16; x++ {
		vec := point.Pt(1, 0).Rotate(math.Pi * 2.0 / 16.0 * x)
		nextPoint := u.Point() + vec.Mul(float64(dist))
		if prevPoint.Floor() == nextPoint.Floor() || !u.Bot.Grid.IsPathable(nextPoint) {
			continue
		}

		// Copy of unit
		cu := *u
		// Move it to the new position
		cu.Pos = nextPoint.To3D()
		newScore := 0.0
		for _, e := range enemies {
			rd := e.RangeDelta(&cu, gap)
			if rd < 0 {
				newScore += e.GroundDPS() * (1 - rd/(e.GroundRange()+float64(e.Radius+u.Radius)))
			}
		}

		if newScore < score {
			fbp = nextPoint
			score = newScore
		}
		prevPoint = nextPoint
	}

	isSafe := fbp == safePos
	return fbp, isSafe
}*/

func (u *Unit) GroundFallback(enemies Units, gap float64, safePos point.Point) {
	if UnitsOrders[u.Tag].Loop+AttackDelay.Max(u.UnitType, u.Bot.FramesPerOrder) > u.Bot.Loop {
		return // Not more than FramesPerOrder
	}
	// fbp, _ := u.GroundFallbackPos(enemies, gap, safePath, 5)
	fbp := safePos
	if !u.Bot.SafeGrid.IsPathable(fbp) {
		if pos := u.Bot.FindClosestPathable(u.Bot.SafeGrid, fbp); pos != 0 {
			fbp = pos
		}
	}
	from := u.Point()
	if !u.Bot.SafeGrid.IsPathable(from) {
		if pos := u.Bot.FindClosestPathable(u.Bot.SafeGrid, from); pos != 0 {
			from = pos
		}
	}
	if from != 0 {
		navGrid, waymap := u.GetWayMap(true)
		path, _ := NavPath(navGrid, waymap, u, fbp)
		pos := path.FirstFurtherThan(2, u)
		if pos != 0 {
			fbp = pos
		}
	}
	if u.WeaponCooldown > 0 && u.PosDelta == 0 {
		u.SpamCmds = true
	}
	u.CommandPos(ability.Move, fbp)
}

func (u *Unit) IsCloserThan(dist float64, ptr point.Pointer) bool {
	return u.Dist2(ptr) < dist*dist
}

func (u *Unit) IsFurtherThan(dist float64, ptr point.Pointer) bool {
	return u.Dist2(ptr) > dist*dist
}

func (u *Unit) IsFarFrom(ptr point.Pointer) bool {
	return u.IsFurtherThan(u.SightRange()/2, ptr)
}

func (u *Unit) EstimatePositionAfter(frames int) point.Point {
	return u.Point() + u.PosDelta.Norm().Mul(float64(u.Speed())*float64(frames)/22.4)
}

func (u *Unit) FramesToPos(ptr point.Pointer) float64 {
	return u.Dist(ptr) / u.Speed() * 16
}

func (u *Unit) TargetAbility() api.AbilityID {
	if len(u.Orders) == 0 {
		return 0
	}
	return u.Orders[0].AbilityId
}

func (u *Unit) TargetPos() point.Point {
	if len(u.Orders) == 0 {
		return 0
	}
	return point.Pt3(u.Orders[0].GetTargetWorldSpacePos())
}

func (u *Unit) TargetTag() api.UnitTag {
	if len(u.Orders) == 0 {
		return 0
	}
	return u.Orders[0].GetTargetUnitTag()
}

func (u *Unit) FindAssignedBuilder(builders Units) *Unit {
	for _, builder := range builders {
		// log.Info(builder.TargetAbility(), builder.TargetPos(), u.Point(), builder.TargetTag(), u.Tag)
		if builder.TargetAbility() == ability.Build_Refinery {
			geyser := u.Bot.Units.Geysers.All().ByTag(builder.TargetTag())
			// log.Info(geyser.Point(), u.Point())
			if geyser != nil && geyser.Point() == u.Point() {
				return builder
			}
		}
		if builder.TargetPos() == u.Point() || builder.TargetTag() == u.Tag {
			return builder
		}
	}
	return nil
}

func (u *Unit) FindAssignedRepairers(reps Units) Units {
	us := Units{}
	for _, rep := range reps {
		if rep.TargetTag() == u.Tag {
			us.Add(rep)
		}
	}
	return us
}

type AttackFunc func(u *Unit, priority int, targets Units) bool
type MoveFunc func(u *Unit, target *Unit)

func DefaultAttackFunc(u *Unit, priority int, targets Units) bool {
	if priority > 0 && !u.IsHalfCool() { // todo: test IsHalfCool!
		return false // Don't focus on secondary targets if weapons are not cool yet
	}
	closeTargets := targets.Filter(Visible).InRangeOf(u, 0)
	if closeTargets.Exists() {
		target := closeTargets.Min(func(unit *Unit) float64 {
			return unit.Hits
		})
		u.CommandTag(ability.Attack_Attack, target.Tag)
		return true
	}
	return false
}

func DefaultMoveFunc(u *Unit, target *Unit) {
	// Unit need to be closer to the target to shoot?
	if !u.InRange(target, -0.1) || !target.IsVisible() || !target.IsPosVisible() {
		u.AttackMove(target)
	}
}

func (u *Unit) EvadeEffectsPos(ptr point.Pointer, checkKD8 bool, eids ...api.EffectID) (point.Point, bool) { // bool - is safe
	upos := ptr.Point()
	// And also reaper mines
	if !u.IsFlying && checkKD8 {
		kds := append(u.Bot.Units.My[terran.KD8Charge], u.Bot.Units.Enemy[terran.KD8Charge]...)
		if kds.Exists() {
			kd := kds.ClosestTo(upos)
			gap := upos.Dist(kd) - float64(u.Radius) - KD8Radius - 0.1
			// Negative if under effect
			if gap < 0 {
				// Negative towards = outwards
				pos := upos.Towards(kd, gap-1)
				if u.Bot.Grid.IsPathable(pos) {
					return pos, false
				}
			}
		}
	}
	for _, e := range u.Bot.Obs.RawData.Effects {
		for _, eid := range eids {
			if e.EffectId == eid {
				for _, p2 := range e.Pos {
					p := point.Pt2(p2)
					gap := upos.Dist(p) - float64(Effects[eid].Radius+u.Radius) - 0.1
					if gap < 0 {
						pos := upos.Towards(p, gap-1)
						if u.IsFlying || u.Bot.Grid.IsPathable(pos) {
							return pos, false
						}
					}
				}
			}
		}
	}
	return upos, true
}

func (u *Unit) AttackMove(target *Unit) {
	npos := u.Towards(target, 2)
	if !u.IsFlying {
		if p := u.GroundTowards(target, 2, false); p != 0 {
			npos = p
		}
	}
	// todo: move into safe grids
	effects := []api.EffectID{effect.PsiStorm, effect.CorrosiveBile}
	if !u.IsFlying {
		effects = append(effects, effect.LiberatorDefenderZoneSetup, effect.LiberatorDefenderZone)
	}
	pos, safe := u.EvadeEffectsPos(npos, true, effects...)
	if safe {
		enemies := u.Bot.Enemies.AllReady
		if u.IsFlying {
			pos, safe = u.AirEvade(enemies, 2, npos)
		} else {
			pos, safe = u.GroundEvade(enemies, 2, npos)
		}
		if !safe {
			friendsDPS := u.Bot.Units.My.All().CloserThan(7, u).Sum(CmpGroundDPS)
			enemiesDPS := enemies.CloserThan(7, target).Sum(CmpGroundDPS)
			if friendsDPS >= enemiesDPS {
				safe = true
			}
		}
	}
	if u.WeaponCooldown > 0 && u.PosDelta == 0 {
		u.SpamCmds = true // Spamming this thing is the key. Or orders will be ignored (or postponed)
	}
	if safe {
		// Move closer
		u.CommandPos(ability.Move, target)
	} else {
		u.CommandPos(ability.Move, pos)
	}
}

func (u *Unit) AttackCustom(attackFunc AttackFunc, moveFunc MoveFunc, targetsGroups ...Units) {
	if UnitsOrders[u.Tag].Loop+u.Bot.FramesPerOrder > u.Bot.Loop {
		return // Not more than FramesPerOrder
	}

	// Here we try to shoot at any target close enough
	for priority, targets := range targetsGroups {
		if attackFunc(u, priority, targets) {
			return
		}
	}

	// Can't shoot anyone. Move closer to targets
	for _, targets := range targetsGroups {
		target := targets.ClosestTo(u)
		if target == nil {
			continue
		}

		moveFunc(u, target)
		return // No orders if unit is close enough
	}
}

func (u *Unit) Attack(targetsGroups ...Units) { // Targets in priority from higher to lower
	u.AttackCustom(DefaultAttackFunc, DefaultMoveFunc, targetsGroups...)
}

// Filters
func Idle(u *Unit) bool       { return u.IsIdle() }
func Unused(u *Unit) bool     { return u.IsUnused() }
func Ready(u *Unit) bool      { return u.IsReady() }
func Gathering(u *Unit) bool  { return u.IsGathering() }
func Visible(u *Unit) bool    { return u.IsVisible() }
func PosVisible(u *Unit) bool { return u.IsPosVisible() }
func Structure(u *Unit) bool  { return u.IsStructure() }
func Flying(u *Unit) bool     { return u.IsFlying }
func NotFlying(u *Unit) bool  { return !u.IsFlying }
func NotWorker(u *Unit) bool  { return !u.IsWorker() }
func DpsGt5(u *Unit) bool     { return u.GroundDPS() > 5 }
func NoAddon(u *Unit) bool    { return u.AddOnTag == 0 }
func HasTechlab(u *Unit) bool { return u.HasTechlab() }
func HasReactor(u *Unit) bool { return u.HasReactor() }
func Mineral(u *Unit) bool    { return u.IsMineral() }