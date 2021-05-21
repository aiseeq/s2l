package scl

import (
	"github.com/aiseeq/s2l/lib/actions"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
)

type CommandsStack struct {
	Simple      map[api.AbilityID]Units
	SimpleQueue map[api.AbilityID]Units
	Pos         map[api.AbilityID]map[point.Point]Units
	PosQueue    map[api.AbilityID]map[point.Point]Units
	Tag         map[api.AbilityID]map[api.UnitTag]Units
	TagQueue    map[api.AbilityID]map[api.UnitTag]Units
}

func (cs *CommandsStack) AddSimple(a api.AbilityID, queue bool, u ...*Unit) {
	if queue {
		if cs.SimpleQueue == nil {
			cs.SimpleQueue = map[api.AbilityID]Units{}
		}
		c := cs.SimpleQueue[a]
		c = append(c, u...)
		cs.SimpleQueue[a] = c
	} else {
		if cs.Simple == nil {
			cs.Simple = map[api.AbilityID]Units{}
		}
		c := cs.Simple[a]
		c = append(c, u...)
		cs.Simple[a] = c
	}
}

func (cs *CommandsStack) AddPos(a api.AbilityID, p point.Point, queue bool, u ...*Unit) {
	if queue {
		if cs.PosQueue == nil {
			cs.PosQueue = map[api.AbilityID]map[point.Point]Units{}
		}
		c := cs.PosQueue[a]

		if c == nil {
			c = map[point.Point]Units{}
		}
		cp := c[p]
		cp = append(cp, u...)
		c[p] = cp

		cs.PosQueue[a] = c
	} else {
		if cs.Pos == nil {
			cs.Pos = map[api.AbilityID]map[point.Point]Units{}
		}
		c := cs.Pos[a]

		if c == nil {
			c = map[point.Point]Units{}
		}
		cp := c[p]
		cp = append(cp, u...)
		c[p] = cp

		cs.Pos[a] = c
	}
}

func (cs *CommandsStack) AddTag(a api.AbilityID, t api.UnitTag, queue bool, u ...*Unit) {
	if queue {
		if cs.TagQueue == nil {
			cs.TagQueue = map[api.AbilityID]map[api.UnitTag]Units{}
		}
		c := cs.TagQueue[a]

		if c == nil {
			c = map[api.UnitTag]Units{}
		}
		ct := c[t]
		ct = append(ct, u...)
		c[t] = ct

		cs.TagQueue[a] = c
	} else {
		if cs.Tag == nil {
			cs.Tag = map[api.AbilityID]map[api.UnitTag]Units{}
		}
		c := cs.Tag[a]

		if c == nil {
			c = map[api.UnitTag]Units{}
		}
		ct := c[t]
		ct = append(ct, u...)
		c[t] = ct

		cs.Tag[a] = c
	}
}

func (cs *CommandsStack) ProcessSimple(actions *actions.Actions, simple map[api.AbilityID]Units, queue bool) {
	for ability, units := range simple {
		us := Units{}
		if queue {
			// No repeat checks for queued orders
			us = units
			for _, unit := range units {
				// But save history for the last action
				B.U.UnitsOrders[unit.Tag] = UnitOrder{
					Loop:    B.Loop,
					Ability: ability,
				}
			}
		} else {
			for _, unit := range units {
				// Check unit's current state and last order
				if unit.SpamCmds ||
					(unit.TargetAbility() != ability &&
						(unit.IsIdle() || B.U.UnitsOrders[unit.Tag].Ability != ability)) {
					us.Add(unit)
					B.U.UnitsOrders[unit.Tag] = UnitOrder{
						Loop:    B.Loop,
						Ability: ability,
					}
				}
			}
		}
		if us.Exists() {
			*actions = append(*actions, &api.Action{
				ActionRaw: &api.ActionRaw{
					Action: &api.ActionRaw_UnitCommand{
						UnitCommand: &api.ActionRawUnitCommand{
							AbilityId:    ability,
							UnitTags:     us.Tags(),
							QueueCommand: queue,
						}}}})
		}
	}
}

func (cs *CommandsStack) ProcessPos(actions *actions.Actions, posList map[api.AbilityID]map[point.Point]Units, queue bool) {
	for ability, positions := range posList {
		for position, units := range positions {
			us := Units{}
			if queue {
				// No repeat checks for queued orders
				us = units
				for _, unit := range units {
					// But save history for the last action
					B.U.UnitsOrders[unit.Tag] = UnitOrder{
						Loop:    B.Loop,
						Ability: ability,
						Pos:     position,
					}
				}
			} else {
				for _, unit := range units {
					// Check unit's current state and last order
					/*log.Info(unit.TargetAbility(), ability)
					log.Info((unit.TargetPos()-position).Len())
					log.Info(UnitsOrders[unit.Tag].Ability, ability)
					log.Info(UnitsOrders[unit.Tag].Pos, position)*/
					if unit.SpamCmds ||
						((unit.TargetAbility() != ability ||
							(unit.TargetPos()-position).Len() > samePoint) &&
							(unit.IsIdle() ||
								B.U.UnitsOrders[unit.Tag].Ability != ability ||
								B.U.UnitsOrders[unit.Tag].Pos != position)) {
						us.Add(unit)
						B.U.UnitsOrders[unit.Tag] = UnitOrder{
							Loop:    B.Loop,
							Ability: ability,
							Pos:     position,
						}
					}
				}
			}
			if us.Exists() {
				*actions = append(*actions, &api.Action{
					ActionRaw: &api.ActionRaw{
						Action: &api.ActionRaw_UnitCommand{
							UnitCommand: &api.ActionRawUnitCommand{
								AbilityId:    ability,
								UnitTags:     us.Tags(),
								QueueCommand: queue,
								Target: &api.ActionRawUnitCommand_TargetWorldSpacePos{
									TargetWorldSpacePos: position.To2D(),
								}}}}})
			}
		}
	}
}

func (cs *CommandsStack) ProcessTag(actions *actions.Actions, tagList map[api.AbilityID]map[api.UnitTag]Units, queue bool) {
	for ability, tags := range tagList {
		for tag, units := range tags {
			us := Units{}
			if queue {
				// No repeat checks for queued orders
				us = units
				for _, unit := range units {
					// But save history for the last action
					B.U.UnitsOrders[unit.Tag] = UnitOrder{
						Loop:    B.Loop,
						Ability: ability,
						Tag:     tag,
					}
				}
			} else {
				for _, unit := range units {
					// Check unit's current state and last order
					if unit.SpamCmds ||
						((unit.TargetAbility() != ability ||
							unit.TargetTag() != tag) &&
							(unit.IsIdle() ||
								B.U.UnitsOrders[unit.Tag].Ability != ability ||
								B.U.UnitsOrders[unit.Tag].Tag != tag)) {
						us.Add(unit)
						B.U.UnitsOrders[unit.Tag] = UnitOrder{
							Loop:    B.Loop,
							Ability: ability,
							Tag:     tag,
						}
					}
				}
			}
			if us.Exists() {
				*actions = append(*actions, &api.Action{
					ActionRaw: &api.ActionRaw{
						Action: &api.ActionRaw_UnitCommand{
							UnitCommand: &api.ActionRawUnitCommand{
								AbilityId:    ability,
								UnitTags:     us.Tags(),
								QueueCommand: queue,
								Target: &api.ActionRawUnitCommand_TargetUnitTag{
									TargetUnitTag: tag,
								}}}}})
			}
		}
	}
}

func (cs *CommandsStack) Process(actions *actions.Actions) {
	cs.ProcessSimple(actions, cs.Simple, false)
	cs.ProcessPos(actions, cs.Pos, false)
	cs.ProcessTag(actions, cs.Tag, false)

	// todo: Тут, скорее всего, бага. Но она проявится если давать несколько команд в очереди сразу
	// они могут быть разного типа и идти в любом порядке. Здесь же, походу, пределён порядок типов
	// Короче, будет работать только если сначала давать одну команду без очереди, а потом одну с очередью и всё
	cs.ProcessSimple(actions, cs.SimpleQueue, true)
	cs.ProcessPos(actions, cs.PosQueue, true)
	cs.ProcessTag(actions, cs.TagQueue, true)
}

func (u *Unit) Command(ability api.AbilityID) {
	B.Cmds.AddSimple(ability, false, u)
}

func (u *Unit) CommandQueue(ability api.AbilityID) {
	B.Cmds.AddSimple(ability, true, u)
}

func (u *Unit) CommandPos(ability api.AbilityID, target point.Pointer) {
	B.Cmds.AddPos(ability, target.Point(), false, u)
}

func (u *Unit) CommandPosQueue(ability api.AbilityID, target point.Pointer) {
	B.Cmds.AddPos(ability, target.Point(), true, u)
}

func (u *Unit) CommandTag(ability api.AbilityID, target api.UnitTag) {
	B.Cmds.AddTag(ability, target, false, u)
}

func (u *Unit) CommandTagQueue(ability api.AbilityID, target api.UnitTag) {
	B.Cmds.AddTag(ability, target, true, u)
}

func (us Units) Command(ability api.AbilityID) {
	if us.Empty() {
		return
	}
	B.Cmds.AddSimple(ability, false, us...)
}

func (us Units) CommandQueue(ability api.AbilityID) {
	if us.Empty() {
		return
	}
	B.Cmds.AddSimple(ability, true, us...)
}

func (us Units) CommandPos(ability api.AbilityID, target point.Pointer) {
	if us.Empty() {
		return
	}
	B.Cmds.AddPos(ability, target.Point(), false, us...)
}

func (us Units) CommandPosQueue(ability api.AbilityID, target point.Pointer) {
	if us.Empty() {
		return
	}
	B.Cmds.AddPos(ability, target.Point(), true, us...)
}

func (us Units) CommandTag(ability api.AbilityID, target api.UnitTag) {
	if us.Empty() {
		return
	}
	B.Cmds.AddTag(ability, target, false, us...)
}

func (us Units) CommandTagQueue(ability api.AbilityID, target api.UnitTag) {
	if us.Empty() {
		return
	}
	B.Cmds.AddTag(ability, target, true, us...)
}
