package scl

import (
	"bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/protocol/api"
)

type UnitTypes []api.UnitTypeID
type Aliases map[api.UnitTypeID]UnitTypes
type TagsMap map[api.UnitTag]bool
type TagsByTypes map[api.UnitTypeID]TagsMap
type UnitsByTypes map[api.UnitTypeID]Units
type AttackDelays map[api.UnitTypeID]int

func (ut *UnitTypes) Add(ids ...api.UnitTypeID) {
NextId:
	for _, id := range ids {
		for _, i := range *ut {
			if i == id {
				continue NextId
			}
		}
		*ut = append(*ut, id)
	}
}

func (ut UnitTypes) Contain(id api.UnitTypeID) bool {
	for _, i := range ut {
		if i == id {
			return true
		}
	}
	return false
}

func (as Aliases) Add(td *api.UnitTypeData) {
	aliases := as[td.UnitId]
	aliases.Add(td.UnitId)
	for _, ta := range td.TechAlias {
		aliases.Add(ta)
		aliases.Add(B.U.UnitAliases[ta]...)
	}
	if td.UnitAlias != 0 {
		aliases.Add(td.UnitAlias)
	}
	as[td.UnitId] = aliases
	for _, a := range aliases {
		subAliases := as[a]
		subAliases.Add(aliases...)
		as[a] = subAliases
	}
}

func (as Aliases) For(ut api.UnitTypeID) UnitTypes {
	aliases, ok := as[ut]
	if !ok {
		log.Warningf("No alias for %s", B.U.Types[ut].Name)
		aliases = UnitTypes{ut}
	}
	return aliases
}

func (as Aliases) Min(ut api.UnitTypeID) api.UnitTypeID {
	min := ut
	for _, tid := range as.For(ut) {
		if tid < min {
			min = tid
		}
	}
	return min
}

func (ut UnitsByTypes) Add(utype api.UnitTypeID, unit *Unit) {
	ut[utype] = append(ut[utype], unit)
}

func (ut UnitsByTypes) OfType(ids ...api.UnitTypeID) Units {
	u := Units{}
	for _, id := range ids {
		u = append(u, ut[id]...)
	}
	return u
}

func (ut UnitsByTypes) All() Units {
	u := Units{}
	for _, units := range ut {
		u = append(u, units...)
	}
	return u
}

func (ut UnitsByTypes) Empty() bool {
	for _, units := range ut {
		if units.Len() > 0 {
			return false
		}
	}
	return true
}

func (ut UnitsByTypes) Exists() bool {
	return !ut.Empty()
}

func (ad AttackDelays) Max(ut api.UnitTypeID, frames int) int {
	if delay, ok := ad[ut]; ok && delay > frames {
		return delay
	}
	return frames
}

func (ad AttackDelays) IsCool(ut api.UnitTypeID, cooldown float32, frames int) bool {
	return cooldown < float32(ad.Max(ut, frames))
}

func (ad AttackDelays) UnitIsCool(u *Unit) bool {
	return ad.IsCool(u.UnitType, u.WeaponCooldown, B.FramesPerOrder)
}

func (tt *TagsByTypes) Add(ut api.UnitTypeID, tag api.UnitTag) {
	if *tt == nil {
		*tt = TagsByTypes{}
	}
	ut = B.U.UnitAliases.Min(ut)
	if (*tt)[ut] == nil {
		(*tt)[ut] = TagsMap{}
	}
	(*tt)[ut][tag] = true
}

func (tt TagsByTypes) Len(ut api.UnitTypeID) int {
	return len(tt[B.U.UnitAliases.Min(ut)])
}

func (tt TagsByTypes) Score(uts ...api.UnitTypeID) int {
	score := 0
	for _, ut := range uts {
		ut = B.U.UnitAliases.Min(ut)
		score += len(tt[ut]) * int(B.U.Types[ut].MineralCost+B.U.Types[ut].VespeneCost)
	}
	return score
}
