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
	SpamCmds     bool
	HPS          float64
	Hits         float64
	HitsMax      float64
	HitsLost     float64
	Abilities    []api.AbilityID
	IrrAbilities []api.AbilityID
	PosDelta     point.Point
	Neighbours   Units
	Cluster      *Cluster
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

func (b *Bot) InitUnits(typeData []*api.UnitTypeData) {
	b.U.Types = typeData

	for _, td := range b.U.Types {
		b.U.Attributes[td.UnitId] = map[api.Attribute]bool{}
		for _, attribute := range td.Attributes {
			b.U.Attributes[td.UnitId][attribute] = true
		}
		cost := Cost{
			Minerals: int(td.MineralCost),
			Vespene:  int(td.VespeneCost),
			Food:     int(td.FoodRequired - td.FoodProvided),
			Time:     int(td.BuildTime),
		}
		// fixes
		if td.Race == api.Race_Zerg && b.U.Attributes[td.UnitId][api.Attribute_Structure] {
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
			weapon := *b.U.Types[terran.Marine].Weapons[0] // Make copy
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
					Speed:   0.224, // Cooldown: 0.16 seconds / 16 frames * 22.4 frames = 0.224
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
		b.U.UnitCost[td.UnitId] = cost
		b.U.AbilityCost[td.AbilityId] = cost
		b.U.AbilityUnit[td.AbilityId] = td.UnitId
		b.U.UnitAbility[td.UnitId] = td.AbilityId
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
			b.U.Weapons[td.UnitId] = w
		}
		b.U.UnitAliases.Add(td)

		// find cells of ground attack circles for units
		if b.U.Weapons[td.UnitId].ground != nil {
			r := float64(b.U.Weapons[td.UnitId].ground.Range)
			r += 2 // Max unit radius + max target radius
			r2 := r * r
			ps := point.Points{}
			// Count from center of the cell where unit is
			for y := -math.Ceil(r); y <= math.Ceil(r); y++ {
				for x := -math.Ceil(r); x <= math.Ceil(r); x++ {
					if x*x+y*y <= r2 {
						ps.Add(point.Pt(x, y))
					}
				}
			}
			b.U.GroundAttackCircle[td.UnitId] = ps
		}
	}
}

func (b *Bot) InitUpgrades(upgradeData []*api.UpgradeData) {
	b.U.Upgrades = upgradeData

	for _, ud := range b.U.Upgrades {
		cost := Cost{
			Minerals: int(ud.MineralCost),
			Vespene:  int(ud.VespeneCost),
			Food:     0,
			Time:     int(ud.ResearchTime),
		}
		b.U.AbilityCost[ud.AbilityId] = cost
		// api bug workaroubd: TerranVehicleArmorsLevel1 -> Research_TerranVehicleAndShipPlatingLevel1
		if ud.AbilityId == 852 {
			b.U.AbilityCost[864] = cost
		}
		if ud.AbilityId == 853 {
			b.U.AbilityCost[865] = cost
		}
		if ud.AbilityId == 854 {
			b.U.AbilityCost[866] = cost
		}
		// log.Info(ud)
	}
}

func (b *Bot) InitEffects(effectData []*api.EffectData) {
	b.U.Effects = effectData
}

func (b *Bot) NewUnit(unit *api.Unit) (*Unit, bool) {
	u := &Unit{
		Unit:    *unit,
		Hits:    float64(unit.Health + unit.Shield),
		HitsMax: float64(unit.HealthMax + unit.ShieldMax),
	}
	if u.Alliance == api.Alliance_Neutral || u.DisplayType == api.DisplayType_Snapshot {
		return u, false
	}

	// Check saved orders, because order itself is not in observation yet if B.FramesPerOrder not passed
	order, ok := b.U.UnitsOrders[u.Tag]
	if ok && order.Loop+B.FramesPerOrder > B.Loop {
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
	pu, ok := b.U.PrevUnits[u.Tag]
	if ok {
		isNew = false
	} else {
		pu = u
	}
	hits := b.U.HitsHistory[u.Tag]

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

	b.U.HitsHistory[u.Tag] = hits
	b.U.PrevUnits[u.Tag] = u
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
		navGrid = B.SafeGrid
		waymap = B.SafeWayMap
		if u.UnitType == terran.Reaper && B.ReaperGrid != nil && B.ReaperWayMap != nil {
			navGrid = B.ReaperSafeGrid
			waymap = B.ReaperSafeWayMap
		}
	} else {
		navGrid = B.Grid
		waymap = B.WayMap
		if u.UnitType == terran.Reaper && B.ReaperGrid != nil && B.ReaperWayMap != nil {
			navGrid = B.ReaperGrid
			waymap = B.ReaperWayMap
		}
	}
	return navGrid, waymap
}

func (u *Unit) GroundTowards(ptr point.Pointer, offset float64, safe bool) point.Point {
	navGrid, waymap := u.GetWayMap(safe)
	path, _ := NavPath(navGrid, waymap, u, ptr) // This function eats most (~20%) of the main thread
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
	reactor := B.Units.My.OfType(B.U.UnitAliases.For(terran.Reactor)...).ByTag(u.AddOnTag)
	if reactor != nil && reactor.IsReady() {
		return len(u.Orders) < 2
	}
	return len(u.Orders) == 0
}

func (u *Unit) IsMoving() bool {
	return len(u.Orders) > 0 && (u.Orders[0].AbilityId == ability.Move || u.Orders[0].AbilityId == ability.Move_Move)
}

// Used to check if it is ok to issue next _move_ order without interrupting current attack action
func (u *Unit) IsCoolToMove() bool {
	if delay, ok := B.U.AfterAttack[u.UnitType]; ok && B.Loop-B.U.LastAttack[u.Tag] < delay {
		return false
	}
	return true
	// return B.U.AfterAttack.UnitIsCool(u)
}

// Тут нужно определять способен ли юнит нанести удар развернувшись без дополнительной задержки
// Но так же надо следить, успеют ли укусить рипера или другой отступающий юнит
// Пока что надёжнее дожидаться u.WeaponCooldown == 0, тогда
// Рипер убегая от лингов или зилота атакует реже и тем самым может сохранять дистанцию
// Used to check if unit is ready to attack right away
func (u *Unit) IsCoolToAttack() bool {
	return u.WeaponCooldown <= 0 // Sometimes it is NEGATIVE!
}

// Used to prevent switches between targets with same priority without actually attacking anything
func (u *Unit) IsAlreadyAttackingTargetInRange() bool {
	target := B.Enemies.All.ByTag(u.TargetTag())
	if target != nil && u.InRange(target, 0) {
		return true
	}
	return false
}

func (u *Unit) IsVisible() bool {
	return u.DisplayType == api.DisplayType_Visible
}

func (u *Unit) IsHidden() bool {
	return u.DisplayType == api.DisplayType_Hidden
}

func (u *Unit) IsPosVisible() bool {
	return B.Grid.IsVisible(u)
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
	return B.U.Attributes[u.UnitType][api.Attribute_Structure]
}

func (u *Unit) IsArmored() bool {
	return B.U.Attributes[u.UnitType][api.Attribute_Armored]
}

func (u *Unit) IsLight() bool {
	return B.U.Attributes[u.UnitType][api.Attribute_Light]
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

func (u *Unit) HasIrrAbility(a api.AbilityID) bool { // Ignore resource requirement
	for _, abil := range u.IrrAbilities {
		if abil == a {
			return true
		}
	}
	return false
}

func (u *Unit) HasTechlab() bool {
	if u.AddOnTag != 0 {
		tl := B.Units.My.OfType(B.U.UnitAliases.For(terran.TechLab)...).ByTag(u.AddOnTag)
		if tl != nil && tl.IsReady() {
			return true
		}
	}
	return false
}

func (u *Unit) HasReactor() bool {
	if u.AddOnTag != 0 {
		tl := B.Units.My.OfType(B.U.UnitAliases.For(terran.Reactor)...).ByTag(u.AddOnTag)
		if tl != nil && tl.IsReady() {
			return true
		}
	}
	return false
}

func (u *Unit) Speed() float64 {
	return float64(B.U.Types[u.UnitType].MovementSpeed)
}

func (u *Unit) GroundDPS() float64 {
	return B.U.Weapons[u.UnitType].groundDps
}

func (u *Unit) AirDPS() float64 {
	return B.U.Weapons[u.UnitType].airDps
}

func (u *Unit) GroundDamage() float64 {
	return B.U.Weapons[u.UnitType].groundDamage
}

func (u *Unit) AirDamage() float64 {
	return B.U.Weapons[u.UnitType].airDamage
}

func (u *Unit) IsArmed() bool {
	return u.GroundDamage() > 0 || u.AirDamage() > 0
}

func (u *Unit) GroundRange() float64 {
	if weapon := B.U.Weapons[u.UnitType].ground; weapon != nil {
		return float64(weapon.Range)
	}
	return -1
}

func (u *Unit) AirRange() float64 {
	if weapon := B.U.Weapons[u.UnitType].air; weapon != nil {
		return float64(weapon.Range)
	}
	return -1
}

func (u *Unit) SightRange() float64 {
	return float64(B.U.Types[u.UnitType].SightRange)
}

func (u *Unit) RangeDelta(target *Unit, gap float64) float64 {
	unitRange := -100.0
	if u.GroundDPS() > 0 && Ground(target) {
		unitRange = u.GroundRange()
	}
	// Air range is always larger than ground
	if u.AirDPS() > 0 && Flying(target) {
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

func (u *Unit) AssessStrength(attackers Units) (outranged, stronger bool) {
	closestUnit := attackers.CanAttack(u, 4).ClosestTo(u)
	if closestUnit == nil {
		stronger = true
		return
	}
	if Ground(u) {
		outranged = closestUnit.GroundRange() >= u.GroundRange()
	}
	if Flying(u) {
		outranged = closestUnit.AirRange() >= math.Max(u.GroundRange(), u.AirRange())
	}
	if outranged {
		// 14 - max possible unit range (Tempest)
		friendsScore := B.Units.My.All().CloserThan(14, u).Filter(DpsGt5).Sum(CmpTotalScore)
		enemiesScore := B.Enemies.AllReady.CloserThan(14, closestUnit).Filter(DpsGt5).Sum(CmpTotalScore)
		// log.Info(friendsScore, enemiesScore, friendsScore*1.25 >= enemiesScore)
		if friendsScore*1.25 >= enemiesScore {
			stronger = true
		}
	}
	return
}

func (u *Unit) AirEvade(enemies Units, gap float64, ptr point.Pointer) (point.Point, bool) { // bool = is safe
	pos := ptr.Point()
	if enemies.Empty() {
		return pos, true // Unit can just move to desired position
	}

	// Copy of unit
	cu := *u
	delta := (pos - u.Point()).Norm()
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
	if outrange > -1 && u.IsNot(terran.Medivac, terran.Raven) { // Close to the border, less than 1
		// Not for units without attack
		rangeVec := (u.Point() - hazard.Point()).Norm()
		tangVec := (rangeVec * 1i).Mul(airSpeedBoostRange)
		p1 := u.Point() + tangVec
		p2 := u.Point() - tangVec
		if p1.Dist2(pos) < p2.Dist2(pos) {
			return p1, false
		}
		return p2, false
	}
	// Move directly from enemy
	return u.Towards(hazard, -airSpeedBoostRange), false
}

func (u *Unit) GroundEvade(enemies Units, gap float64, ptr point.Pointer) (point.Point, bool) { // bool = is safe
	pos := ptr.Point()
	if enemies.Empty() {
		return pos, true // Unit can just move to desired position
	}

	delta := (pos - u.Point()).Norm()

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
		rangeVec := (u.Point() - hazard.Point()).Norm()
		tangVec := rangeVec * 1i
		p1 := u.Point() + tangVec
		p2 := u.Point() - tangVec
		if p1.Dist2(pos) < p2.Dist2(pos) {
			escVec = p1
		} else {
			escVec = p2
		}
	} else {
		// Move directly from enemy
		escVec = (u.Point() - hazard.Point()).Norm()
	}
	if !B.Grid.IsPathable(u.Point() + escVec) {
		for x := 1.0; x < 4; x++ {
			esc1 := u.Point() + escVec.Rotate(math.Pi*2.0/16.0*x)
			if B.Grid.IsPathable(esc1) {
				return esc1, false
			}
			esc2 := u.Point() + escVec.Rotate(-math.Pi*2.0/16.0*x)
			if B.Grid.IsPathable(esc2) {
				return esc2, false
			}
		}
		return B.Locs.MyStart.Towards(B.Locs.MapCenter, -3), false // Try to go home
	}
	return u.Point() + escVec, false
}

func (u *Unit) GroundFallback(safePos point.Pointer, ignoreAttackAbility bool) {
	// ignoreAttackAbility is for cyclone
	if !u.IsCoolToMove() || !ignoreAttackAbility && u.IsCoolToAttack() && u.IsAlreadyAttackingTargetInRange() {
		return // Don't move until attack is done
	}
	// fbp, _ := u.GroundFallbackPos(enemies, gap, safePath, 5)
	fbp := safePos
	navGrid, waymap := u.GetWayMap(true)
	if !navGrid.IsPathable(fbp) {
		if pos := B.FindClosestPathable(navGrid, fbp); pos != 0 {
			fbp = pos
		}
	}
	from := u.Point()
	if !navGrid.IsPathable(from) {
		if pos := B.FindClosestPathable(navGrid, from); pos != 0 {
			from = pos
		}
	}
	if from != 0 {
		path, _ := NavPath(navGrid, waymap, u, fbp)
		pos := path.FirstFurtherThan(2, u)
		if pos != 0 {
			fbp = pos
		}
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
	return u.Point() + u.PosDelta.Norm().Mul(u.Speed()*float64(frames)/22.4)
}

func (u *Unit) FramesToPos(ptr point.Pointer) float64 {
	return u.Dist(ptr) / u.Speed() * 22.4
}

func (u *Unit) FramesToDistantPos(ptr point.Pointer) float64 {
	return B.RequestPathing(u, ptr) / u.Speed() * 22.4
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
			geyser := B.Units.Geysers.All().ByTag(builder.TargetTag())
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
	if priority > 0 && !u.IsCoolToAttack() {
		return false // Don't focus on secondary targets if weapons are not cool yet
	}
	closeTargets := targets.Filter(Visible).InRangeOf(u, 0)
	if closeTargets.Exists() && u.IsCoolToAttack() {
		target := closeTargets.Min(func(unit *Unit) float64 {
			return unit.Hits
		})
		// log.Info(u.UnitType, B.U.BeforeAttack[u.UnitType], B.Loop - B.U.LastAttack[u.Tag])
		u.CommandTag(ability.Attack_Attack, target.Tag)
		B.U.LastAttack[u.Tag] = B.Loop
		return true
	}
	return false
}

func (u *Unit) GetEffectsList() []api.EffectID {
	// todo: move into safe grids
	effects := []api.EffectID{effect.PsiStorm, effect.CorrosiveBile}
	if !u.IsFlying {
		effects = append(effects, effect.LiberatorDefenderZoneSetup, effect.LiberatorDefenderZone,
			effect.ThermalLance, effect.BlindingCloud, effect.LurkerSpines, effect.TemporalFieldGrowing,
			effect.TemporalField)
	}
	return effects
}

func (u *Unit) EvadeEffects() bool {
	pos, safe := u.EvadeEffectsPos(u, true, u.GetEffectsList()...)
	if !safe {
		u.CommandPos(ability.Move, pos)
		return true
	}
	return false
}

func DefaultMoveFunc(u *Unit, target *Unit) {
	// Unit need to be closer to the target to shoot?
	if !u.InRange(target, -0.1) || !target.IsVisible() || !target.IsPosVisible() {
		u.AttackMove(target)
	} else {
		// Evade effects
		u.EvadeEffects()
	}
}

func (u *Unit) EvadeEffectsPos(ptr point.Pointer, checkKD8 bool, eids ...api.EffectID) (point.Point, bool) { // bool - is safe
	upos := ptr.Point()
	// And also reaper mines
	if !u.IsFlying && checkKD8 {
		kds := append(B.Units.My[terran.KD8Charge], B.Units.Enemy[terran.KD8Charge]...)
		if kds.Exists() {
			kd := kds.ClosestTo(upos)
			gap := upos.Dist(kd) - float64(u.Radius) - KD8Radius - 0.5
			// Negative if under effect
			if gap < 0 {
				// Negative towards = outwards
				pos := upos.Towards(kd, gap-1)
				if B.Grid.IsPathable(pos) {
					return pos, false
				}
			}
		}
	}
	for _, e := range append(B.Obs.RawData.Effects, B.RecentEffects[0]...) {
		for _, eid := range eids {
			if e.EffectId == eid {
				for _, p2 := range e.Pos {
					p := point.Pt2(p2)
					gap := upos.Dist(p) - float64(B.U.Effects[eid].Radius+u.Radius) - 0.5
					if gap < 0 {
						pos := upos.Towards(p, gap-1)
						if upos == p {
							// Rare case when effect is directly above the unit (not so rare vs bots)
							pos = upos.Towards(B.Locs.MapCenter, gap-1)
						}
						return pos, false
					}
				}
			}
		}
	}
	return upos, true
}

func (u *Unit) AttackMove(target *Unit) {
	dist := u.Dist(target)
	rads := float64(u.Radius + target.Radius)
	npos := u.Towards(target, math.Min(2, dist-rads))
	if !u.IsFlying && dist-rads > 2 {
		if p := u.GroundTowards(target, 2, false); p != 0 {
			npos = p
		}
	}

	effects := u.GetEffectsList()
	pos, safe := u.EvadeEffectsPos(u, true, effects...)
	if safe {
		pos, safe = u.EvadeEffectsPos(npos, true, effects...)
		if safe {
			enemies := B.Enemies.AllReady
			if u.IsFlying {
				pos, safe = u.AirEvade(enemies, 2, npos)
			} else {
				pos, safe = u.GroundEvade(enemies, 2, npos)
			}
			if !safe && target.Cloak != api.CloakState_Cloaked {
				outranged, stronger := u.AssessStrength(enemies)
				if !outranged || stronger {
					safe = true
				}
			}
		}
	}
	if safe {
		// Move closer
		u.CommandPos(ability.Move, target)
	} else {
		u.CommandPos(ability.Move, pos)
	}
}

func (u *Unit) AttackCustom(attackFunc AttackFunc, moveFunc MoveFunc, targetsGroups ...Units) {
	if B.U.UnitsOrders[u.Tag].Loop+B.FramesPerOrder > B.Loop {
		return // Not more than FramesPerOrder
	}

	// Don't send another attack command or that could switch targets and attack will fail
	if u.IsCoolToAttack() && !u.IsAlreadyAttackingTargetInRange() {
		// Here we try to shoot at any target close enough
		for priority, targets := range targetsGroups {
			if attackFunc(u, priority, targets) {
				return
			}
		}
	}

	// Don't move until previous attack is done
	if u.IsCoolToMove() {
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
}

func (u *Unit) Attack(targetsGroups ...Units) { // Targets in priority from higher to lower
	u.AttackCustom(DefaultAttackFunc, DefaultMoveFunc, targetsGroups...)
}

func (u *Unit) IsSafeToApproach(p point.Pointer) bool {
	if !B.SafeGrid.IsPathable(p) {
		if pathablePos := B.FindClosestPathable(B.SafeGrid, p); pathablePos != 0 {
			p = pathablePos
		}
	}
	navGrid, waymap := u.GetWayMap(true)
	if path, _ := NavPath(navGrid, waymap, u, p); path == nil {
		return false
	}
	return true
}

// Filters
func Idle(u *Unit) bool         { return u.IsIdle() }
func Unused(u *Unit) bool       { return u.IsUnused() }
func Ready(u *Unit) bool        { return u.IsReady() || u.Cloak == api.CloakState_Cloaked }
func Gathering(u *Unit) bool    { return u.IsGathering() }
func Visible(u *Unit) bool      { return u.IsVisible() }
func PosVisible(u *Unit) bool   { return u.IsPosVisible() }
func Hidden(u *Unit) bool       { return u.IsHidden() }
func Structure(u *Unit) bool    { return u.IsStructure() }
func NotStructure(u *Unit) bool { return !u.IsStructure() }
func Flying(u *Unit) bool       { return u.IsFlying || u.UnitType == protoss.Colossus }
func Ground(u *Unit) bool       { return !u.IsFlying || u.UnitType == protoss.Colossus }
func NotWorker(u *Unit) bool    { return !u.IsWorker() }
func DpsGt5(u *Unit) bool       { return u.GroundDPS() > 5 || u.Is(terran.Hellion, protoss.Sentry) }
func NoAddon(u *Unit) bool      { return u.AddOnTag == 0 }
func HasTechlab(u *Unit) bool   { return u.HasTechlab() }
func HasReactor(u *Unit) bool   { return u.HasReactor() }
func Mineral(u *Unit) bool      { return u.IsMineral() }
