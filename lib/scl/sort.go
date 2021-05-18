package scl

import (
	"github.com/aiseeq/s2l/lib/point"
	"sort"
)

func SortByOther(mainSlice []point.Point, otherSlice []float64) []point.Point {
	sbo := sortByOther{mainSlice, otherSlice}
	sort.Sort(sbo)
	return sbo.mainSlice
}

type sortByOther struct {
	mainSlice  []point.Point
	otherSlice []float64
}

func (sbo sortByOther) Len() int {
	return len(sbo.mainSlice)
}

func (sbo sortByOther) Swap(i, j int) {
	sbo.mainSlice[i], sbo.mainSlice[j] = sbo.mainSlice[j], sbo.mainSlice[i]
	sbo.otherSlice[i], sbo.otherSlice[j] = sbo.otherSlice[j], sbo.otherSlice[i]
}

func (sbo sortByOther) Less(i, j int) bool {
	return sbo.otherSlice[i] < sbo.otherSlice[j]
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
