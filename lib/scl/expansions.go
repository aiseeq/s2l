package scl

import (
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"math"
)

// Taken from Chippydip's lib
// Cluster breaks a list of units into clusters based on the given clustering distance.
func MakeCluster(units Units, distance float64) []UnitCluster {
	maxDistance := distance * distance

	var clusters []UnitCluster
	for _, u := range units {
		if !u.IsMineral() && !u.IsGeyser() {
			continue
		}
		// Find the nearest cluster
		minDist := math.MaxFloat64
		clusterIndex := -1
		for i, cluster := range clusters {
			d := point.Pt3(u.Pos).Dist2(cluster.Center())
			if d < minDist {
				minDist = d
				clusterIndex = i
			}
		}

		// If too far, add a new cluster
		if minDist > maxDistance || clusterIndex < 0 {
			clusterIndex = len(clusters)
			clusters = append(clusters, UnitCluster{})
		}

		clusters[clusterIndex].Add(u)
	}
	return clusters
}

// UnitCluster is a cluster of units and the associated center of mass.
type UnitCluster struct {
	sum   point.Point
	units Units
}

// Add adds a new unit to the cluster and updates the center of mass.
func (c *UnitCluster) Add(u *Unit) {
	c.sum = c.sum.Add(float64(u.Pos.X), float64(u.Pos.Y))
	c.units.Add(u)
}

// Center is the center of mass of the cluster.
func (c *UnitCluster) Center() point.Point {
	return c.sum.Mul(1 / float64(c.units.Len()))
}

// Units is the list of units in the cluster.
func (c *UnitCluster) Units() *Units {
	return &c.units
}

// Count returns the number of units in the cluster.
func (c *UnitCluster) Count() int {
	return c.units.Len()
}

// markUnbuildable marks a w x h area around px, py (minus corners) as unbuildable (red)
func markUnbuildable(placement api.ImageDataBytes, px, py, w, h int) {
	xMin, xMax := int32(px-3), int32(px+w+2)
	yMin, yMax := int32(py-3), int32(py+h+2)

	for y := yMin; y <= yMax; y++ {
		for x := xMin; x <= xMax; x++ {
			if (y == yMin || y == yMax) && (x == xMin || x == xMax) {
				continue // skip corners
			}
			if placement.Get(x, y) == 255 {
				placement.Set(x, y, 1)
			}
		}
	}
}

// expandUnbuildable marks any tile within 2 units of px, py as unbuildable (blue)
func expandUnbuildable(placement api.ImageDataBytes, px, py int32) {
	xMin, xMax := px-2, px+2
	yMin, yMax := py-2, py+2

	for y := yMin; y <= yMax; y++ {
		for x := xMin; x <= xMax; x++ {
			if placement.Get(x, y) == 255 {
				placement.Set(x, y, 128)
			}
		}
	}
}

// CalculateExpansionLocations groups resources into clusters and determines the best town hall location for each cluster.
// The Center() point of each cluster is the optimal town hall location. If debug is true then the results will also
// be visualized in-game (until new debug info is drawn).
func (b *Bot) CalculateExpansionLocations() []UnitCluster {
	// Start by finding resource clusters
	resources := append(b.Units.Minerals.All(), b.Units.Geysers.All()...)
	clusters := MakeCluster(resources, 15)

	// Add resource-restrictions to the placement grid
	placement := b.Info.StartRaw.PlacementGrid.Bits().ToBytes()
	for _, u := range resources {
		if u.IsMineral() {
			markUnbuildable(placement, int(u.Pos.X-0.5), int(u.Pos.Y), 2, 1)
		} else if u.IsGeyser() {
			markUnbuildable(placement, int(u.Pos.X-1), int(u.Pos.Y-1), 3, 3)
		}
	}

	// Mark locations which *can't* have an expansion centers
	for y := int32(0); y < placement.Height(); y++ {
		for x := int32(0); x < placement.Width(); x++ {
			if placement.Get(x, y) < 128 {
				expandUnbuildable(placement, x, y)
			}
		}
	}

	// Find the nearest remaining square to each cluster's CoM
	for i, cluster := range clusters {
		pt := cluster.Center()
		px, py := int32(pt.X()), int32(pt.Y())
		r2Min, xBest, yBest := int32(256), int32(-1), int32(-1)
		for r := int32(0); r*r <= r2Min; r++ { // search radius
			xMin, xMax, yMin, yMax := px-r, px+r, py-r, py+r
			for y := yMin; y <= yMax; y++ {
				for x := xMin; x <= xMax; x++ {
					// This is slightly inefficient, but much easier than repeating the same loop 4x for the edges
					if (x == xMin || x == xMax || y == yMin || y == yMax) && placement.Get(x, y) == 255 {
						dx, dy := x-px, y-py
						r2 := dx*dx + dy*dy
						if r2 < r2Min {
							r2Min = r2
							xBest = x
							yBest = y
						}
					}
				}
			}
		}

		// Update the Center to be the detected location rather than the actual CoM (just don't add new units)
		clusters[i].sum = point.Pt(float64(xBest), float64(yBest)).Mul(float64(cluster.units.Len()))
	}

	return clusters
}
