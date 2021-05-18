package scl

import "github.com/aiseeq/s2l/protocol/api"

type Tags []api.UnitTag
type GroupID int
type Group struct {
	Tags  Tags
	Units Units
	// Task  int // this will change
}
type Groups struct {
	list     map[GroupID]Group
	names    map[string]GroupID
	units    map[api.UnitTag]GroupID
	MaxGroup GroupID
}

func NewGroups(maxGroup GroupID) *Groups {
	gs := Groups{}
	gs.list = map[GroupID]Group{}
	gs.names = map[string]GroupID{}
	gs.units = map[api.UnitTag]GroupID{}
	gs.MaxGroup = maxGroup
	return &gs
}

func (gs *Groups) Add(group GroupID, units ...*Unit) {
	if group > gs.MaxGroup {
		gs.MaxGroup = group
	}
	for _, unit := range units {
		// Always one group per unit. To remove unit from group, just add it to another
		oldGroup, ok := gs.units[unit.Tag]
		if ok && oldGroup == group {
			continue
		}
		g := gs.list[group]
		g.Tags.Add(unit.Tag)
		g.Units.Add(unit)
		gs.list[group] = g

		og := gs.list[oldGroup]
		og.Tags.Remove(unit.Tag)
		og.Units.RemoveTag(unit.Tag)
		gs.list[oldGroup] = og

		gs.units[unit.Tag] = group
	}
}

func (gs *Groups) AddN(groupName string, units ...*Unit) {
	group, ok := gs.names[groupName]
	if !ok {
		group = gs.MaxGroup + 1
		gs.names[groupName] = group
	}
	gs.Add(group, units...)
}

func (gs *Groups) New(units ...*Unit) GroupID {
	newGroup := gs.MaxGroup + 1
	gs.Add(newGroup, units...)
	return newGroup
}

// Add unit info to the corresponding group
func (gs *Groups) Fill(unit *Unit) {
	if group, ok := gs.units[unit.Tag]; ok {
		g := gs.list[group]
		g.Units.Add(unit)
		gs.list[group] = g
	}
}

func (gs *Groups) Get(group GroupID) Group {
	return gs.list[group]
}

func (gs *Groups) GetN(groupName string) Group {
	if group, ok := gs.names[groupName]; ok {
		return gs.list[group]
	}
	return Group{}
}

func (gs *Groups) ClearUnits() {
	for group, list := range gs.list {
		list.Units = nil
		gs.list[group] = list
	}
}

func (ts *Tags) Add(tags ...api.UnitTag) {
	*ts = append(*ts, tags...)
}

func (ts *Tags) Remove(tag api.UnitTag) {
	for k, t := range *ts {
		if t == tag {
			if len(*ts) > k+1 {
				*ts = append((*ts)[:k], (*ts)[k+1:]...)
			} else {
				*ts = (*ts)[:k] // Remove last
			}
		}
	}
}

func (ts Tags) Len() int {
	return len(ts)
}

func (ts Tags) Empty() bool {
	return len(ts) == 0
}

func (ts Tags) Exists() bool {
	return len(ts) > 0
}
