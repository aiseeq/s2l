package scl

import (
	"bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/grid"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/enums/terran"
	"math"
	"time"
)

type Cell struct {
	point.Point
	Distance float64
}
type CellsPair struct{ src, dst *Cell }
type Queue map[float64][]CellsPair
type Steps map[point.Point]*Cell // from -> next step
type Paths struct {
	Steps Steps
	Queue Queue
	B     MapAccessor
}

func (paths *Paths) Present(p point.Point) bool {
	_, ok := paths.Steps[p]
	return ok
}

func (paths *Paths) AddCell(src, dst *Cell, reaper bool) {
	if paths.Present(src.Point) {
		return
	}
	paths.Steps[src.Point] = dst

	var d float64
	for n, p := range src.Neighbours8(1) {
		if paths.Present(p) {
			continue
		}
		distMul := 1.0
		if !paths.B.IsPathable(p) {
			if !reaper || paths.B.IsBuildable(p) {
				continue
			}
			p += p - src.Point // Shift 1 cell further
			if !paths.B.IsBuildable(p) || !paths.B.IsPathable(p) ||
				paths.B.HeightAt(src.Point) == paths.B.HeightAt(p) {
				continue
			}
			// So, here p should be 2 cells from src, it is buildable and pathable at different height
			// But point between them isn't buildable or pathable => reaper can jump here
			distMul = 2.0
		}

		if n < 4 {
			d = src.Distance + distMul // Straight
		} else {
			d = src.Distance + distMul*math.Sqrt2 // Diagonal
		}

		nc := &Cell{
			Point:    p,
			Distance: d,
		}
		paths.Queue[d] = append(paths.Queue[d], CellsPair{src: nc, dst: src})
	}
}

func (q Queue) MinDist() ([]CellsPair, float64) {
	var cps []CellsPair
	minK := math.Inf(1)
	for k, v := range q {
		if k < minK {
			cps = v
			minK = k
		}
	}
	return cps, minK
}

func (steps Steps) From(ptr point.Pointer) point.Points {
	p := ptr.Point().Floor()
	ps := point.Points{p}
	for x := 0; x < 1000; x++ {
		c, ok := steps[p]
		if !ok {
			return nil
		}
		p = c.Point
		ps.Add(p)
		if c.Distance == 0 {
			return ps
		}
	}
	log.Errorf("Can't find path from %v", ptr)
	return nil
}

func (steps Steps) Follow(ptr point.Pointer, limit int) point.Point {
	p := ptr.Point().Floor()
	for x := 0; x < limit; x++ {
		c, ok := steps[p]
		if !ok {
			return 0
		}
		p = c.Point
	}
	return p
}

func (b *Bot) FindPaths(grid *grid.Grid, ptr point.Pointer, reaper bool) Steps {
	paths := Paths{
		Steps: Steps{},
		Queue: Queue{},
		B:     grid,
	}
	initCell := &Cell{Point: ptr.Point().Floor(), Distance: 0}
	// Last point of path links to itself
	paths.AddCell(initCell, initCell, reaper)
	for cps, key := paths.Queue.MinDist(); cps != nil; cps, key = paths.Queue.MinDist() {
		delete(paths.Queue, key) // Clear queue element to not repeat it
		for _, cp := range cps {
			paths.AddCell(cp.src, cp.dst, reaper)
		}
	}

	return paths.Steps
}

func (paths *Paths) FindPathableCell(src, dst *Cell) point.Point {
	if paths.Present(src.Point) {
		return 0
	}
	paths.Steps[src.Point] = dst

	var d float64
	for n, p := range src.Neighbours8(1) {
		if paths.Present(p) {
			continue
		}
		if paths.B.IsPathable(p) {
			return p
		}

		if n < 4 {
			d = src.Distance + 1 // Straight
		} else {
			d = src.Distance + math.Sqrt2 // Diagonal
		}

		nc := &Cell{
			Point:    p,
			Distance: d,
		}
		paths.Queue[d] = append(paths.Queue[d], CellsPair{src: nc, dst: src})
	}
	return 0
}

func (b *Bot) FindClosestPathable(grid *grid.Grid, ptr point.Pointer) point.Point {
	if grid.IsPathable(ptr) {
		return ptr.Point().Floor()
	}
	paths := Paths{
		Steps: Steps{},
		Queue: Queue{},
		B:     grid,
	}
	initCell := &Cell{Point: ptr.Point().Floor(), Distance: 0}
	// Last point of path links to itself
	if p := paths.FindPathableCell(initCell, initCell); p != 0 {
		return p
	}
	for cps, key := paths.Queue.MinDist(); cps != nil; cps, key = paths.Queue.MinDist() {
		delete(paths.Queue, key) // Clear queue element to not repeat it
		for _, cp := range cps {
			if p := paths.FindPathableCell(cp.src, cp.dst); p != 0 {
				return p
			}
		}
	}
	return 0
}

func (b *Bot) FindHomeMineralsVector() {
	var vec point.Point
	homeMinerals := b.Units.Minerals.All().CloserThan(ResourceSpreadDistance, b.Locs.MyStart)
	if homeMinerals.Exists() {
		vec = homeMinerals.Center().Dir(b.Locs.MyStart)
	}
	if vec.Len() <= 1 {
		vec = b.Locs.MyStart.Dir(b.Locs.MapCenter)
	}
	b.Locs.MyStartMinVec = vec
}

func (b *Bot) RenewPaths() {
	for {
		b.Grid.Lock()
		navGrid := grid.New(b.Grid.StartRaw, b.Grid.MapState)
		// this also locks units remap todo: separate?
		reapersExists := b.Units.My[terran.Reaper].Exists()
		b.Grid.Unlock()

		lastLoop := b.Loop

		/*if b.HomePaths == nil { // don't rebuild them
			b.HomePaths = b.FindPaths(navGrid, b.Locs.MyStart-b.Locs.MyStartMinVec*3, false)
			// log.Info(time.Now().Sub(start))
			b.HomeReaperPaths = b.FindPaths(navGrid, b.Locs.MyStart-b.Locs.MyStartMinVec*3, true)
			for key, pos := range b.Locs.MyExps {
				if !navGrid.IsBuildable(pos) {
					b.ExpPaths[key] = b.FindPaths(navGrid, pos-b.Locs.MyStartMinVec*3, false)
				}
			}
			// b.DebugPath(b.HomeReaperPaths.From(b.EnemyRamp.Top))
			// b.DebugSend()
		}*/

		// s := time.Now()
		b.WayMap = b.FindWaypointsMap(navGrid)

		if reapersExists {
			// Need to renew it because it can't be locked somewhere else
			reaperGrid := grid.New(navGrid.StartRaw, navGrid.MapState)
			pa := b.Info.StartRaw.PlayableArea
			for y := pa.P0.Y; y <= pa.P1.Y; y++ {
				for x := pa.P0.X; x <= pa.P1.X; x++ {
					p := point.Pt(float64(x), float64(y))
					if reaperGrid.IsPathable(p) {
						continue
					}
					// if points on left & right or up & down are pathable and on different height,
					// reaper can jump there
					pr := p + 1
					pl := p - 1
					pu := p + 1i
					pd := p - 1i
					pur := p + 1 + 1i
					pdr := p + 1 - 1i
					pdl := p - 1 - 1i
					pul := p - 1 + 1i
					if (reaperGrid.IsPathable(pl) && reaperGrid.IsPathable(pr) &&
						reaperGrid.HeightAt(pl) != reaperGrid.HeightAt(pr)) ||
						(reaperGrid.IsPathable(pu) && reaperGrid.IsPathable(pd) &&
							reaperGrid.HeightAt(pu) != reaperGrid.HeightAt(pd)) ||
						(reaperGrid.IsPathable(pur) && reaperGrid.IsPathable(pdl) &&
							reaperGrid.HeightAt(pur) != reaperGrid.HeightAt(pdl)) ||
						(reaperGrid.IsPathable(pul) && reaperGrid.IsPathable(pdr) &&
							reaperGrid.HeightAt(pul) != reaperGrid.HeightAt(pdr)) {
						reaperGrid.SetPathable(p, true)
					}
				}
			}
			b.ReaperGrid = reaperGrid
			b.ReaperWayMap = b.FindWaypointsMap(b.ReaperGrid)
		}

		// s := time.Now()
		safeGrid := grid.New(navGrid.StartRaw, navGrid.MapState)
		var reaperSafeGrid *grid.Grid
		if reapersExists {
			reaperSafeGrid = grid.New(navGrid.StartRaw, navGrid.MapState)
		}
		for _, u := range b.Enemies.AllReady {
			pos := u.Point().Floor()
			ps := b.U.GroundAttackCircle[u.UnitType]
			for _, p := range ps {
				safeGrid.SetPathable(pos+p, false)
				if reapersExists {
					reaperSafeGrid.SetPathable(pos+p, false)
				}
			}
		}
		// todo: add effects
		b.SafeGrid = safeGrid
		b.SafeWayMap = b.FindWaypointsMap(b.SafeGrid)
		if reapersExists {
			b.ReaperSafeGrid = reaperSafeGrid
			b.ReaperSafeWayMap = b.FindWaypointsMap(b.ReaperSafeGrid)
		}
		/*log.Info(time.Now().Sub(s))
		wps := point.Points{}
		for p := range b.SafeWayMap {
			wps = append(wps, p.Point)
		}
		path, _ := NavPath(b.SafeGrid, b.SafeWayMap, b.Locs.MyStart-3, b.Locs.EnemyStart-3)
		b.Grid.Lock() // prevents grid rewrite because debug uses b.Grid
		b.DebugPath(wps, White)
		b.DebugPath(path, Yellow)
		b.Grid.Unlock()
		b.DebugSend()*/

		// continue // don't wait, more updates!
		for lastLoop+3 > b.Loop {
			time.Sleep(time.Millisecond)
		}
	}
}
