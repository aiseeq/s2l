package scl

import (
	"github.com/aiseeq/s2l/lib/point"
	"github.com/beefsack/go-astar"
	"math"
)

type MapAccessor interface {
	IsPathable(p point.Pointer) bool
	IsBuildable(p point.Pointer) bool
	HeightAt(p point.Pointer) float64
}

// A Tile is a tile in a grid which implements Pather.
type Tile struct {
	// X and Y are the coordinates of the tile.
	X, Y float64
	// B is a reference to the Bot
	B MapAccessor
	// Tiles storage
	M map[float64]map[float64]*Tile
}

func (t *Tile) Map(x, y float64) *Tile {
	if t.M == nil {
		t.M = map[float64]map[float64]*Tile{}
	}
	row := t.M[y]
	if row == nil {
		row = map[float64]*Tile{}
	}
	tile, ok := row[x]
	if ok {
		return tile
	}
	tile = &Tile{x, y, t.B, t.M}
	row[x] = tile
	t.M[y] = row
	return tile
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Tile) PathNeighbors() []astar.Pather {
	var neighbors []astar.Pather
	pos := point.Pt(t.X, t.Y)
	for _, p := range pos.Neighbours8(1) {
		if !t.B.IsPathable(p) {
			continue
		}
		neighbors = append(neighbors, t.Map(p.X(), p.Y()))
	}
	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	t2 := to.(*Tile)
	p1 := point.Pt(t.X, t.Y)
	p2 := point.Pt(t2.X, t2.Y)
	delta := p2 - p1

	if delta.X() != 0 && delta.Y() != 0 {
		return math.Sqrt2
	}
	return 1
}

func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	t2 := to.(*Tile)
	// Fast, but suboptimal
	return math.Abs(t.X-t2.X) + math.Abs(t.Y-t2.Y)
	// Slow (3-4 times), but correct
	/*p := point.Pt(math.Abs(t.X - t2.X), math.Abs(t.Y - t2.Y))
	max := math.Max(p.X(), p.Y())
	straight := math.Abs(p.X() - p.Y())
	diagonal := max - straight
	return straight + diagonal * math.Sqrt2*/
}

// Params in reverse order because astar.Path returns reversed list
func (b *Bot) Path(toPtr, fromPtr point.Pointer) (point.Points, float64) {
	from := fromPtr.Point().Floor()
	to := toPtr.Point().Floor()
	f := &Tile{X: from.X(), Y: from.Y(), B: b.Grid}
	t := f.Map(to.X(), to.Y())
	// start := time.Now()
	path, dist, found := astar.Path(f, t)
	// log.Info(time.Now().Sub(start))
	if !found {
		return nil, 0
	}

	ps := point.Points{}
	for _, i := range path {
		t := i.(*Tile)
		ps.Add(point.Pt(t.X, t.Y))
	}
	return ps, dist
}
