package scl

import (
	"github.com/aiseeq/s2l/lib/grid"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/beefsack/go-astar"
	"math"
)

type Waypoint struct {
	point.Point
	M WaypointsMap
}
type Waypoints []*Waypoint
type WaypointsMap map[*Waypoint]Waypoints

func (wpm WaypointsMap) NewWaypoint(p point.Point) *Waypoint {
	return &Waypoint{
		Point: p,
		M:     wpm,
	}
}

func BresenhamsLineDrawable(p0, p1 point.Point, grid *grid.Grid) bool {
	x0 := int(p0.X())
	y0 := int(p0.Y())
	x1 := int(p1.X())
	y1 := int(p1.Y())

	// implemented straight from WP pseudocode
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		if !grid.IsPathableFast(x0, y0) {
			return false
		}
		if x0 == x1 && y0 == y1 {
			return true
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (b *Bot) FindWaypoints(wpm WaypointsMap, grid *grid.Grid) Waypoints {
	waypoints := Waypoints{}
	pa := b.Info.StartRaw.PlayableArea
	p0x := int(pa.P0.X)
	p0y := int(pa.P0.Y)
	p1x := int(pa.P1.X)
	p1y := int(pa.P1.Y)
	for y := p0y; y <= p1y; y++ {
		for x := p0x; x <= p1x; x++ {
			if !grid.IsPathableFast(x, y) {
				continue
			}
			for _, dy := range []int{-1, 1} {
				for _, dx := range []int{-1, 1} {
					if !grid.IsPathableFast(x+dx, y+dy) &&
						grid.IsPathableFast(x+dx, y) &&
						grid.IsPathableFast(x, y+dy) &&
						// Diagonal lines optimization
						(grid.IsPathableFast(x+2*dx, y) || grid.IsPathableFast(x, y+2*dy)) {
						waypoints = append(waypoints, wpm.NewWaypoint(point.Pt(float64(x), float64(y))))
						break
					}
				}
			}
		}
	}
	return waypoints
}

func FindNeighbours(wpm WaypointsMap, from *Waypoint, waypoints Waypoints, skip int, grid *grid.Grid) Waypoints {
	neighbours := Waypoints{}
	for _, to := range waypoints[skip:] {
		if to != from && BresenhamsLineDrawable(from.Point, to.Point, grid) {
			wpm[from] = append(wpm[from], to)
			wpm[to] = append(wpm[to], from)
		}
	}
	return neighbours
}

func (b *Bot) FindWaypointsMap(grid *grid.Grid) WaypointsMap {
	wpm := WaypointsMap{}
	// start := time.Now()
	waypoints := b.FindWaypoints(wpm, grid)
	// log.Info(time.Now().Sub(start))

	// start := time.Now()
	for skip, waypoint := range waypoints {
		FindNeighbours(wpm, waypoint, waypoints, skip+1, grid)
	}
	// log.Info(time.Now().Sub(start))
	return wpm
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Waypoint) PathNeighbors() []astar.Pather {
	// return t.M[t] - possible if: type WaypointsMap map[*Waypoint][]astar.Pather
	var neighbors = make([]astar.Pather, len(t.M[t]))
	for k, p := range t.M[t] {
		neighbors[k] = p
	}
	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Waypoint) PathNeighborCost(to astar.Pather) float64 {
	return t.PathEstimatedCost(to)
}

func (t *Waypoint) PathEstimatedCost(to astar.Pather) float64 {
	t2 := to.(*Waypoint)
	delta := t.Point - t2.Point
	return math.Hypot(real(delta), imag(delta))
}

func NavPath(grid *grid.Grid, wpm WaypointsMap, fromPtr, toPtr point.Pointer) (point.Points, float64) {
	var f, t *Waypoint
	from := fromPtr.Point().Floor()
	to := toPtr.Point().Floor()
	if BresenhamsLineDrawable(from, to, grid) {
		// point.Points are on the pathable line, we don't need to search anything
		return point.Points{from, to}, from.Dist(to)
	}

	var wps = make(Waypoints, 0, len(wpm)+2)
	// Maybe one of points is already in the waypoints map, let's find it
	for wp := range wpm {
		if wp.Point == from {
			f = wp
		} else if wp.Point == to {
			t = wp
		}
		wps = append(wps, wp)
	}
	// If it isn't, we need to add start and end points to the map
	if f == nil {
		f = wpm.NewWaypoint(from)
		FindNeighbours(wpm, f, wps, 0, grid)
		wps = append(wps, f)
	}
	if t == nil {
		t = wpm.NewWaypoint(to)
		FindNeighbours(wpm, t, wps, 0, grid)
		wps = append(wps, t)
	}

	// start := time.Now()
	path, dist, found := astar.Path(t, f) // Params in reverse order because astar.Path returns reversed list
	// log.Info(time.Now().Sub(start))
	if !found {
		return nil, 0
	}

	var ps point.Points
	for _, i := range path {
		t := i.(*Waypoint)
		ps.Add(t.Point)
	}
	return ps, dist
}
