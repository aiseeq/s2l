package scl

import (
	"bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/enums/neutral"
	"github.com/aiseeq/s2l/protocol/enums/protoss"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"github.com/aiseeq/s2l/protocol/enums/zerg"
	"math"
)

type Ramp struct {
	Top    point.Point
	Vec    point.Point
	Size   int
	Height float64
}

type BuildingSize int
type CheckMap int
type PathableCells int

const (
	S2x1 BuildingSize = iota + 1
	S2x2
	S3x3
	S5x3
	S5x5
	UnbuildableRocks
	BreakableRocks2x2
	BreakableRocks4x4
	BreakableRocks4x2
	BreakableRocks2x4
	BreakableRocks6x2
	BreakableRocks2x6
	BreakableRocks6x6
	BreakableRocksDiagBLUR
	BreakableRocksDiagULBR
	BreakableHorizontalHuge
	BreakableVerticalHuge
)
const (
	IsBuildable CheckMap = iota + 1
	IsPathable
	IsVisible
	IsExplored
	IsCreep
	IsNoCreep
)
const (
	Zero PathableCells = iota
	One
	Two
)

var DestructibleSize = map[api.UnitTypeID]BuildingSize{
	neutral.UnbuildableBricksDestructible:          UnbuildableRocks,
	neutral.UnbuildableBricksSmallUnit:             UnbuildableRocks,
	neutral.UnbuildableBricksUnit:                  UnbuildableRocks,
	neutral.UnbuildablePlatesDestructible:          UnbuildableRocks,
	neutral.UnbuildablePlatesSmallUnit:             UnbuildableRocks,
	neutral.UnbuildablePlatesUnit:                  UnbuildableRocks,
	neutral.UnbuildableRocksDestructible:           UnbuildableRocks,
	neutral.UnbuildableRocksSmallUnit:              UnbuildableRocks,
	neutral.UnbuildableRocksUnit:                   UnbuildableRocks,
	neutral.Rocks2x2NonConjoined:                   BreakableRocks2x2,
	neutral.DestructibleCityDebris4x4:              BreakableRocks4x4,
	neutral.DestructibleDebris4x4:                  BreakableRocks4x4,
	neutral.DestructibleIce4x4:                     BreakableRocks4x4,
	neutral.DestructibleRock4x4:                    BreakableRocks4x4,
	neutral.DestructibleRockEx14x4:                 BreakableRocks4x4,
	neutral.DestructibleCityDebris2x4Horizontal:    BreakableRocks4x2,
	neutral.DestructibleIce2x4Horizontal:           BreakableRocks4x2,
	neutral.DestructibleRock2x4Horizontal:          BreakableRocks4x2,
	neutral.DestructibleRockEx12x4Horizontal:       BreakableRocks4x2,
	neutral.DestructibleCityDebris2x4Vertical:      BreakableRocks2x4,
	neutral.DestructibleIce2x4Vertical:             BreakableRocks2x4,
	neutral.DestructibleRock2x4Vertical:            BreakableRocks2x4,
	neutral.DestructibleRockEx12x4Vertical:         BreakableRocks2x4,
	neutral.DestructibleCityDebris2x6Horizontal:    BreakableRocks6x2,
	neutral.DestructibleIce2x6Horizontal:           BreakableRocks6x2,
	neutral.DestructibleRock2x6Horizontal:          BreakableRocks6x2,
	neutral.DestructibleRockEx12x6Horizontal:       BreakableRocks6x2,
	neutral.DestructibleCityDebris2x6Vertical:      BreakableRocks2x6,
	neutral.DestructibleIce2x6Vertical:             BreakableRocks2x6,
	neutral.DestructibleRock2x6Vertical:            BreakableRocks2x6,
	neutral.DestructibleRockEx12x6Vertical:         BreakableRocks2x6,
	neutral.DestructibleCityDebris6x6:              BreakableRocks6x6,
	neutral.DestructibleDebris6x6:                  BreakableRocks6x6,
	neutral.DestructibleExpeditionGate6x6:          BreakableRocks6x6,
	neutral.DestructibleIce6x6:                     BreakableRocks6x6,
	neutral.DestructibleRock6x6:                    BreakableRocks6x6,
	neutral.DestructibleRock6x6Weak:                BreakableRocks6x6,
	neutral.DestructibleRockEx16x6:                 BreakableRocks6x6,
	neutral.DestructibleCityDebrisHugeDiagonalBLUR: BreakableRocksDiagBLUR,
	neutral.DestructibleDebrisRampDiagonalHugeBLUR: BreakableRocksDiagBLUR,
	neutral.DestructibleIceDiagonalHugeBLUR:        BreakableRocksDiagBLUR,
	neutral.DestructibleRampDiagonalHugeBLUR:       BreakableRocksDiagBLUR,
	neutral.DestructibleRockEx1DiagonalHugeBLUR:    BreakableRocksDiagBLUR,
	neutral.DestructibleCityDebrisHugeDiagonalULBR: BreakableRocksDiagULBR,
	neutral.DestructibleDebrisRampDiagonalHugeULBR: BreakableRocksDiagULBR,
	neutral.DestructibleIceDiagonalHugeULBR:        BreakableRocksDiagULBR,
	neutral.DestructibleRampDiagonalHugeULBR:       BreakableRocksDiagULBR,
	neutral.DestructibleRockEx1DiagonalHugeULBR:    BreakableRocksDiagULBR,
	neutral.DestructibleIceHorizontalHuge:          BreakableHorizontalHuge,
	neutral.DestructibleRampHorizontalHuge:         BreakableHorizontalHuge,
	neutral.DestructibleRockEx1HorizontalHuge:      BreakableHorizontalHuge,
	neutral.DestructibleIceVerticalHuge:            BreakableVerticalHuge,
	neutral.DestructibleRampVerticalHuge:           BreakableVerticalHuge,
	neutral.DestructibleRockEx1VerticalHuge:        BreakableVerticalHuge,
}

func (b *Bot) CheckPoints(ps point.Points, check CheckMap) bool {
	for _, p := range ps {
		switch check {
		case IsBuildable:
			if !b.Grid.IsBuildable(p) {
				return false
			}
		case IsPathable:
			if !b.Grid.IsPathable(p) {
				return false
			}
		case IsVisible:
			if !b.Grid.IsVisible(p) {
				return false
			}
		case IsExplored:
			if !b.Grid.IsExplored(p) {
				return false
			}
		case IsCreep:
			if !b.Grid.IsCreep(p) {
				return false
			}
		case IsNoCreep:
			if b.Grid.IsCreep(p) {
				return false
			}
		default:
			log.Fatalf("%v check is not implemented", check)
		}
	}
	return true
}

func (b *Bot) GetBuildingPoints(ptr point.Pointer, size BuildingSize) point.Points {
	pos := ptr.Point().Floor()
	switch size {
	case S2x1:
		return point.Points{pos, pos + 1}
	case S2x2:
		fallthrough
	case BreakableRocks2x2:
		fallthrough
	case UnbuildableRocks:
		return point.Points{pos, pos + 1i, pos + 1 + 1i, pos + 1}
	case S3x3:
		return append(pos.Neighbours8(1), pos)
	case S5x3:
		return append(b.GetBuildingPoints(pos, S3x3), b.GetBuildingPoints(pos+2-1i, S2x2)...)
	case S5x5:
		ps := point.Points{}
		for y := pos.Y() - 2; y <= pos.Y()+2; y++ {
			for x := pos.X() - 2; x <= pos.X()+2; x++ {
				ps.Add(point.Pt(x, y))
			}
		}
		return ps
	case BreakableRocks4x2: // non-tested
		return append(b.GetBuildingPoints(pos, S2x2), b.GetBuildingPoints(pos-2, S2x2)...)
	case BreakableRocks2x4: // non-tested
		return append(b.GetBuildingPoints(pos, S2x2), b.GetBuildingPoints(pos-2i, S2x2)...)
	case BreakableRocks6x2: // non-tested
		return append(b.GetBuildingPoints(pos, BreakableRocks4x2), b.GetBuildingPoints(pos+2, S2x2)...)
	case BreakableRocks2x6: // non-tested
		return append(b.GetBuildingPoints(pos, BreakableRocks2x4), b.GetBuildingPoints(pos+2i, S2x2)...)
	case BreakableRocks4x4:
		ps := append(b.GetBuildingPoints(pos, S2x2), b.GetBuildingPoints(pos-2, S2x2)...)
		ps = append(ps, b.GetBuildingPoints(pos-2i, S2x2)...)
		ps = append(ps, b.GetBuildingPoints(pos-2-2i, S2x2)...)
		return ps
	case BreakableRocks6x6:
		return append(b.GetBuildingPoints(pos, BreakableRocks4x4),
			pos-2-3i, pos-1-3i, pos-3i, pos+1-3i, pos-2+2i, pos-1+2i, pos+2i, pos+1+2i,
			pos-3-2i, pos-3-1i, pos-3, pos-3+1i, pos+2-2i, pos+2-1i, pos+2, pos+2+1i)
	case BreakableRocksDiagBLUR:
		ps := point.Points{}
		for y := -4.0; y <= 6; y++ {
			switch y {
			case -4:
				ps = append(ps, pos-point.Pt(y+2, y))
			case -3:
				ps = append(ps, pos-point.Pt(y, y), pos-point.Pt(y+1, y), pos-point.Pt(y+2, y))
			default:
				ps = append(ps, pos-point.Pt(y-2, y), pos-point.Pt(y-1, y), pos-point.Pt(y, y),
					pos-point.Pt(y+1, y), pos-point.Pt(y+2, y))
			case 4:
				ps = append(ps, pos-point.Pt(y-2, y), pos-point.Pt(y-1, y), pos-point.Pt(y, y))
			case 5:
				ps = append(ps, pos-point.Pt(y-2, y))
			}
		}
		return ps
	case BreakableRocksDiagULBR:
		ps := point.Points{}
		for y := -4.0; y <= 6; y++ {
			switch y {
			case -4:
				ps = append(ps, pos-point.Pt(-y-1, y))
			case -3:
				ps = append(ps, pos-point.Pt(-y-1, y), pos-point.Pt(-y, y), pos-point.Pt(-y+1, y))
			default:
				ps = append(ps, pos-point.Pt(-y-1, y), pos-point.Pt(-y, y), pos-point.Pt(-y+1, y),
					pos-point.Pt(-y+2, y), pos-point.Pt(-y+3, y))
			case 4:
				ps = append(ps, pos-point.Pt(-y+1, y), pos-point.Pt(-y+2, y), pos-point.Pt(-y+3, y))
			case 5:
				ps = append(ps, pos-point.Pt(-y+3, y))
			}
		}
		return ps
	case BreakableHorizontalHuge:
		ps := append(b.GetBuildingPoints(pos, BreakableRocks4x4), b.GetBuildingPoints(pos-4, BreakableRocks4x4)...)
		ps = append(ps, b.GetBuildingPoints(pos+4, BreakableRocks4x4)...)
		return ps
	case BreakableVerticalHuge:
		ps := append(b.GetBuildingPoints(pos, BreakableRocks4x4), b.GetBuildingPoints(pos-4i, BreakableRocks4x4)...)
		ps = append(ps, b.GetBuildingPoints(pos+4i, BreakableRocks4x4)...)
		return ps
	}
	log.Warningf("Building size %v is not implemented", size)
	return nil
}

func (b *Bot) GetPathablePoints(ptr point.Pointer, size BuildingSize, cells PathableCells) point.Points {
	if cells == Zero {
		return b.GetBuildingPoints(ptr, size)
	}
	pos := ptr.Point().Floor()
	ps := point.Points{}
	if cells == One {
		switch size {
		case S2x2:
			for y := pos.Y() - 1; y <= pos.Y()+2; y++ {
				for x := pos.X() - 1; x <= pos.X()+2; x++ {
					ps.Add(point.Pt(x, y))
				}
			}
		case S3x3:
			for y := pos.Y() - 2; y <= pos.Y()+2; y++ {
				for x := pos.X() - 2; x <= pos.X()+2; x++ {
					ps.Add(point.Pt(x, y))
				}
			}
		case S5x3:
			// todo: optimize - remove intersection
			ps = append(b.GetPathablePoints(pos, S3x3, One), b.GetPathablePoints(pos+2-1i, S2x2, One)...)
		case S5x5:
			ps = b.GetPathablePoints(pos, S3x3, Two)
		default:
			log.Fatalf("Building size %v is not implemented for cells count %v", size, cells)
		}
		return ps
	}
	if cells == Two {
		switch size {
		case S2x2:
			for y := pos.Y() - 2; y <= pos.Y()+3; y++ {
				for x := pos.X() - 2; x <= pos.X()+3; x++ {
					if math.Abs(pos.X()-x)+math.Abs(pos.Y()-y) == 5 {
						continue // remove corners
					}
					ps.Add(point.Pt(x, y))
				}
			}
		case S3x3:
			for y := pos.Y() - 3; y <= pos.Y()+3; y++ {
				for x := pos.X() - 3; x <= pos.X()+3; x++ {
					if math.Abs(pos.X()-x)+math.Abs(pos.Y()-y) == 6 {
						continue // remove corners
					}
					ps.Add(point.Pt(x, y))
				}
			}
		case S5x3:
			// todo: optimize - remove intersection
			ps = append(b.GetPathablePoints(pos, S3x3, Two), b.GetPathablePoints(pos+2-1i, S2x2, Two)...)
		default:
			log.Fatalf("Building size %v is not implemented for cells count %v", size, cells)
		}
		return ps
	}
	log.Fatalf("Cells count %v is not implemented", cells)
	return nil
}

func (b *Bot) IsPosOk(ptr point.Pointer, size BuildingSize, cells PathableCells, flags ...CheckMap) bool {
	ps := b.GetBuildingPoints(ptr, size)
	for _, flag := range flags {
		if !b.CheckPoints(ps, flag) {
			return false
		}
	}
	if cells != Zero {
		ps = b.GetPathablePoints(ptr, size, cells)
		return b.CheckPoints(ps, IsPathable)
	}
	return true
}

// Return 0 if not found
func (b *Bot) FindClosestPos(ptr point.Pointer, size BuildingSize, cells PathableCells, maxOffset, step int, flags ...CheckMap) point.Point {
	pos := ptr.Point().Floor()
	for offset := 0; offset <= maxOffset; offset += step {
		for y := -float64(offset); y <= float64(offset); y++ {
			for x := -float64(offset); x <= float64(offset); x++ {
				if offset != 0 && math.Abs(x) != float64(offset) && math.Abs(y) != float64(offset) {
					continue // Don't check points in the center again
				}
				p := point.Pt(pos.X()+x, pos.Y()+y)
				if b.IsPosOk(p, size, cells, flags...) {
					return p
				}
			}
		}
	}
	return 0
}

func (b *Bot) FindClusterTopPoints(cluster *point.Points) (point.Points, float64) {
	var ps point.Points
	min := math.MaxFloat64
	max := -min
	h := math.Inf(-1)
	for _, p := range *cluster {
		hp := b.Grid.HeightAt(p)
		if hp < min {
			min = hp
		}
		if hp > max {
			max = hp
		}
		if hp > h {
			ps = nil
			ps.Add(p)
			h = hp
		} else if hp == h {
			ps.Add(p)
		} else {
			// lower point, don't add
		}
	}
	return ps, max - min
}

func (b *Bot) FindRampCluster(p point.Point, cluster *point.Points, rampPoints map[point.Point]bool) {
	if rampPoints[p] {
		return // This is already a part of known ramp
	}

	buildable := b.Grid.IsBuildable(p)
	pathable := b.Grid.IsPathable(p)
	if pathable && !buildable {
		// Probably a part of ramp
		cluster.Add(p)
		rampPoints[p] = true
		for _, np := range p.Neighbours4(1) {
			b.FindRampCluster(np, cluster, rampPoints)
		}
	}
}

func (b *Bot) FindBaseCluster(p point.Point, cluster *point.Points, basePoints map[point.Point]bool) {
	if basePoints[p] {
		return // This is already a part of base
	}

	if b.Grid.IsBuildable(p) {
		cluster.Add(p)
		basePoints[p] = true
		for _, np := range p.Neighbours4(1) {
			b.FindBaseCluster(np, cluster, basePoints)
		}
	}
}

func (b *Bot) FindRamps() {
	rampPoints := map[point.Point]bool{}
	pa := b.Info.StartRaw.PlayableArea
	for y := pa.P0.Y; y <= pa.P1.Y; y++ {
		for x := pa.P0.X; x <= pa.P1.X; x++ {
			var cluster point.Points
			p := point.Pt(float64(x), float64(y))
			b.FindRampCluster(p, &cluster, rampPoints)
			if cluster.Len() < minRampSize {
				continue // Too small for a real ramp
			}

			top, height := b.FindClusterTopPoints(&cluster)
			if top.Len() == cluster.Len() {
				continue // Flat - not a ramp
			}
			tc := top.Center()
			pt := tc.Floor()
			vec := cluster.Center().Dir(pt)
			for x := 0; x < 10; x++ { // Finite cycle in case of very strange ramps
				if b.Grid.IsBuildable(pt) {
					// Sometimes this point is on edge, try to find closer point
					p1 := pt - point.Pt(vec.X(), 0)
					p2 := pt - point.Pt(0, vec.Y())
					// Pick first point that is closer to the center
					if tc.Dist2(p2) < tc.Dist2(p1) {
						p1, p2 = p2, p1
					}
					for _, np := range []point.Point{p1, p2, pt} {
						if b.Grid.IsBuildable(np) {
							b.Ramps.All = append(b.Ramps.All, Ramp{
								Top:    np,
								Vec:    vec,
								Size:   cluster.Len(),
								Height: height})
							break
						}
					}
					break
				}
				pt += vec
			}
		}
	}
}

func (b *Bot) FindBaseCenter(basePos point.Point) point.Point {
	basePoints := map[point.Point]bool{}
	var cluster point.Points
	b.FindBaseCluster(basePos, &cluster, basePoints)
	return cluster.Center()
}

func (b *Bot) InitLocations() {
	pa := b.Info.StartRaw.PlayableArea
	b.Locs.MapCenter = (point.PtI(pa.P0) + point.PtI(pa.P1)).Mul(0.5)

	// My CC is on start position
	b.Locs.MyStart = b.Units.My.OfType(terran.CommandCenter, zerg.Hatchery, protoss.Nexus).First().Point().Floor()
	esl := b.Info.StartRaw.StartLocations
	b.Locs.EnemyStart = point.Pt2(esl[0])
	eslps := point.Points{}
	if len(esl) > 1 {
		for _, p := range esl {
			eslps.Add(point.Pt2(p))
		}
	}
	for p := eslps.ClosestTo(b.Locs.MyStart); eslps.Exists(); p = eslps.ClosestTo(p) {
		b.Locs.EnemyStarts.Add(p)
		eslps.Remove(p)
	}
	b.Locs.EnemyMainCenter = b.FindBaseCenter(b.Locs.EnemyStart)

	b.FindHomeMineralsVector()
}

func (b *Bot) InitRamps() {
	// Find ramps closest to start locations
	for _, ramp := range b.Ramps.All {
		if ramp.Top == 164+44i && ramp.Vec == 1+1i && ramp.Size == 16 ||
			ramp.Top == 43+44i && ramp.Vec == -1+1i && ramp.Size == 16 {
			continue // Strange blockaded ramp on Golden Wall
		}
		if b.Ramps.My.Top == 0 || ramp.Top.Dist2(b.Locs.MyStart) < b.Ramps.My.Top.Dist2(b.Locs.MyStart) {
			b.Ramps.My = ramp
		}
		if b.Ramps.Enemy.Top == 0 || ramp.Top.Dist2(b.Locs.EnemyStart) < b.Ramps.Enemy.Top.Dist2(b.Locs.EnemyStart) {
			b.Ramps.Enemy = ramp
		}
	}
}

func (b *Bot) RequestPathing(p1, p2 point.Pointer) float64 {
	rqps := []*api.RequestQueryPathing{{
		Start: &api.RequestQueryPathing_StartPos{
			StartPos: p1.Point().To2D(),
		},
		EndPos: p2.Point().To2D(),
	}}
	if resp, err := B.Client.Query(api.RequestQuery{Pathing: rqps}); err != nil || len(resp.Pathing) == 0 {
		log.Error(err)
		return 0
	} else {
		return float64(resp.Pathing[0].Distance)
	}
}

func (b *Bot) RequestPlacement(ability api.AbilityID, pos point.Point, builder *Unit) bool {
	var tag api.UnitTag
	if builder != nil {
		tag = builder.Tag
	}
	rps := []*api.RequestQueryBuildingPlacement{{
		AbilityId:      ability,
		TargetPos:      pos.To2D(),
		PlacingUnitTag: tag,
	}}
	if resp, err := B.Client.Query(api.RequestQuery{Placements: rps}); err != nil || len(resp.Placements) == 0 {
		log.Error(err)
		return false
	} else {
		return resp.Placements[0].Result == api.ActionResult_Success
	}
}

func (b *Bot) FindExpansions() {
	b.Locs.MyExps = nil
	var expDists, enemyExpDists []float64
	// Find expansions locations
	for _, uc := range b.CalculateExpansionLocations() {
		center := uc.Center()
		// Fill expansions locations list
		if center == b.Locs.MyStart {
			continue
		}
		b.Locs.MyExps = append(b.Locs.MyExps, center)
		// Make pathing queries. One after another because seems that results can be shuffled (or not and it was me)
		// From my base to that expansion
		dist := b.RequestPathing(b.Locs.MyStart, center)
		if dist == 0 {
			dist = b.Locs.MyStart.Dist(center) * 100
		}
		expDists = append(expDists, dist)
		// From enemy base to the same expansion
		dist = b.RequestPathing(b.Locs.EnemyStart, center)
		if dist == 0 {
			dist = b.Locs.EnemyStart.Dist(center) * 100
		}
		enemyExpDists = append(enemyExpDists, dist)
	}
	b.Locs.EnemyExps = make(point.Points, len(b.Locs.MyExps))
	copy(b.Locs.EnemyExps, b.Locs.MyExps)

	// Sort expansins locations by walking distance from base
	b.Locs.MyExps = SortByOther(b.Locs.MyExps, expDists)
	b.Locs.EnemyExps = SortByOther(b.Locs.EnemyExps, enemyExpDists)
	// b.ExpPaths = make([]Steps, b.Locs.MyExps.Len())
	// log.Info(b.Locs.MyStart, b.Locs.MyExps, b.Locs.EnemyStart, b.Locs.EnemyExps)
}

func (b *Bot) FindRamp2x2Positions(ramp Ramp) point.Points {
	return point.Points{ramp.Top + ramp.Vec*1i*1.5, ramp.Top - ramp.Vec*1i*1.5}
}

// First position is for initial building, second is for addon
func (b *Bot) FindRampBarracksPositions(ramp Ramp) point.Points {
	if ramp.Vec.X() > 0 {
		return point.Points{ramp.Top + ramp.Vec, ramp.Top + ramp.Vec}
	}
	return point.Points{ramp.Top + ramp.Vec, ramp.Top + ramp.Vec - 2}
}

func (b *Bot) EffectPoints(effect api.EffectID) point.Points {
	ps := point.Points{}
	for _, e := range b.Obs.RawData.Effects {
		if e.EffectId == effect {
			for _, p := range e.Pos {
				ps.Add(point.Pt2(p))
			}
		}
	}
	return ps
}
