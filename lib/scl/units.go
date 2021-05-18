package scl

import (
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"math"
	"sort"
)

type Units []*Unit
type UnitsMap map[*Unit]struct{}
type Filter func(unit *Unit) bool
type Compare func(unit *Unit) float64
type Cluster struct {
	Units UnitsMap
	Food  float64
}

func (us *Units) Add(units ...*Unit) {
	*us = append(*us, units...)
}

func (us *Units) Remove(unit *Unit) {
	for k, u := range *us {
		if u == unit {
			if len(*us) > k+1 {
				*us = append((*us)[:k], (*us)[k+1:]...)
			} else {
				*us = (*us)[:k] // Remove last
			}
		}
	}
}

func (us *Units) RemoveTag(tag api.UnitTag) {
	for k, u := range *us {
		if u.Tag == tag {
			*us = append((*us)[:k], (*us)[k+1:]...)
		}
	}
}

func (us *Units) Pop() *Unit {
	if us.Empty() {
		return nil
	}
	u := (*us)[us.Len()-1]
	*us = (*us)[:us.Len()-1]
	return u
}

func (us Units) Len() int {
	return len(us)
}

func (us Units) Empty() bool {
	return len(us) == 0
}

func (us Units) Exists() bool {
	return len(us) > 0
}

// While using each method you should check that it returns not nil
func (us Units) First(filters ...Filter) *Unit {
	if len(filters) == 0 {
		if len(us) == 0 {
			return nil
		}
		return us[0]
	}
NextUnit:
	for _, unit := range us {
		for _, filter := range filters {
			if !filter(unit) {
				continue NextUnit
			}
		}
		return unit
	}
	return nil
}

func (us Units) Min(comp Compare) *Unit {
	var unit *Unit
	minVal := math.Inf(1)
	for _, u := range us {
		val := comp(u)
		if val < minVal {
			unit = u
			minVal = val
		}
	}
	return unit
}

func (us Units) Max(comp Compare) *Unit {
	var unit *Unit
	maxVal := math.Inf(-1)
	for _, u := range us {
		val := comp(u)
		if val > maxVal {
			unit = u
			maxVal = val
		}
	}
	return unit
}

func (us Units) Sum(comp Compare) float64 {
	sum := 0.0
	for _, u := range us {
		sum += comp(u)
	}
	return sum
}

func (us Units) OrderBy(comp Compare, desc bool) {
	sort.Slice(us, func(i, j int) bool {
		return desc != (comp(us[i]) < comp(us[j]))
	})
}

func (us Units) OrderByDistanceTo(ptr point.Pointer, desc bool) {
	pos := ptr.Point()
	sort.Slice(us, func(i, j int) bool {
		return desc != (us[i].Dist2(pos) < us[j].Dist2(pos))
	})
}

func (us Units) Filter(filters ...Filter) Units {
	if len(filters) == 0 {
		return us
	}
	res := Units{}
NextUnit:
	for _, unit := range us {
		for _, filter := range filters {
			if !filter(unit) {
				continue NextUnit
			}
		}
		res = append(res, unit)
	}
	return res
}

func (us Units) ByTag(tag api.UnitTag) *Unit {
	return us.First(func(unit *Unit) bool { return unit.Tag == tag })
}

func (us Units) ByTags(tags Tags) Units {
	tagMap := map[api.UnitTag]bool{}
	for _, tag := range tags {
		tagMap[tag] = true
	}
	return us.Filter(func(unit *Unit) bool { return tagMap[unit.Tag] })
}

func (us Units) ClosestTo(ptr point.Pointer) *Unit {
	p := ptr.Point()
	var closest *Unit
	for _, unit := range us {
		if closest == nil || p.Dist2(closest) > p.Dist2(unit) {
			closest = unit
		}
	}
	return closest
}

func (us Units) FurthestTo(ptr point.Pointer) *Unit {
	p := ptr.Point()
	var furthest *Unit
	for _, unit := range us {
		if furthest == nil || p.Dist2(furthest) < p.Dist2(unit) {
			furthest = unit
		}
	}
	return furthest
}

func (us Units) CloserThan(dist float64, ptr point.Pointer) Units {
	pos := ptr.Point()
	dist2 := dist * dist
	units := Units{}
	for _, unit := range us {
		if unit.Dist2(pos) <= dist2 {
			units.Add(unit)
		}
	}
	return units
}

func (us Units) FurtherThan(dist float64, ptr point.Pointer) Units {
	pos := ptr.Point()
	dist2 := dist * dist
	units := Units{}
	for _, unit := range us {
		if unit.Dist2(pos) >= dist2 {
			units.Add(unit)
		}
	}
	return units
}

// List of units that are in range of u
func (us Units) InRangeOf(u *Unit, gap float64) Units {
	units := Units{}
	for _, unit := range us {
		if u.InRange(unit, gap) {
			units.Add(unit)
		}
	}
	return units
}

// List of units that can attack u
func (us Units) CanAttack(u *Unit, gap float64) Units {
	units := Units{}
	for _, unit := range us {
		if unit.InRange(u, gap) {
			units.Add(unit)
		}
	}
	return units
}

func (us Units) Center() point.Point {
	points := point.Points{}
	for _, unit := range us {
		points.Add(unit.Point())
	}
	return points.Center()
}

func (us Units) Tags() []api.UnitTag {
	var uTags []api.UnitTag
	for _, unit := range us {
		uTags = append(uTags, unit.Tag)
	}
	return uTags
}

func (us Units) Attack(targetsGroups ...Units) {
	for _, u := range us {
		u.Attack(targetsGroups...)
	}
}

// Comparers
func CmpTags(unit *Unit) float64         { return float64(unit.Tag) }
func CmpGroundDamage(unit *Unit) float64 { return unit.GroundDamage() }
func CmpGroundDPS(unit *Unit) float64    { return unit.GroundDPS() }
func CmpGroundScore(unit *Unit) float64  { return unit.GroundDPS() * unit.Hits }
func CmpFood(unit *Unit) float64 {
	if req := Types[unit.UnitType].FoodRequired; req > 0 {
		return float64(req)
	}
	return 0
}
func CmpHits(unit *Unit) float64 { return unit.Hits }
