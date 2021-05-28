package grid

import (
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"sync"
)

type Grid struct {
	StartRaw *api.StartRaw
	MapState *api.MapState
	Mutex    *sync.Mutex

	PathingSizeX int
	PathingSizeY int
}

func New(startRaw *api.StartRaw, mapState *api.MapState) *Grid {
	g := Grid{Mutex: &sync.Mutex{}}
	g.copy(startRaw, mapState)
	g.PathingSizeX = int(g.StartRaw.PathingGrid.Size_.X)
	g.PathingSizeY = int(g.StartRaw.PathingGrid.Size_.Y)
	return &g
}

func (g *Grid) Renew(startRaw *api.StartRaw, mapState *api.MapState) {
	g.Lock()
	g.copy(startRaw, mapState)
	g.Unlock()
}

func (g *Grid) copy(startRaw *api.StartRaw, mapState *api.MapState) {
	g.StartRaw = &api.StartRaw{}
	data, _ := startRaw.Marshal()
	g.StartRaw.Unmarshal(data)

	g.MapState = &api.MapState{}
	data, _ = mapState.Marshal()
	g.MapState.Unmarshal(data)
}

func (g *Grid) GetBitMapAddr(m *api.ImageData, ptr point.Pointer) (int, int) {
	p := ptr.Point()
	addr := int(int32(p.X())+(int32(p.Y()))*m.Size_.X) / 8
	if addr < 0 || addr > len(m.Data)-1 || p.X() < 0 || p.X() > float64(m.Size_.X-1) || p.Y() < 0 {
		return -1, -1
	}
	return addr, 7 - int(p.X())%8
}

func (g *Grid) GetBitMapData(m *api.ImageData, p point.Pointer) int {
	addr, offset := g.GetBitMapAddr(m, p)
	if addr == -1 {
		return -1
	}
	return int(m.Data[addr] & byte(1<<offset))
}

func (g *Grid) GetMapAddr(m *api.ImageData, ptr point.Pointer) int {
	p := ptr.Point()
	addr := int(int32(p.X()) + (int32(p.Y()))*m.Size_.X)
	if addr < 0 || addr > len(m.Data)-1 || p.X() < 0 || p.X() > float64(m.Size_.X-1) || p.Y() < 0 {
		return -1
	}
	return addr
}

func (g *Grid) GetMapData(m *api.ImageData, p point.Pointer) int {
	addr := g.GetMapAddr(m, p)
	if addr == -1 {
		return -1
	}
	return int(m.Data[addr])
}

func (g *Grid) HeightAt(p point.Pointer) float64 {
	m := g.StartRaw.TerrainHeight // m.BitsPerPixel == 8

	data := g.GetMapData(m, p)
	if data == -1 {
		return 0
	}
	return (float64(data) - 127) / 8
}

func (g *Grid) IsBuildable(p point.Pointer) bool {
	m := g.StartRaw.PlacementGrid

	data := g.GetBitMapData(m, p)
	if data == -1 {
		return false
	}
	return data != 0
}

func (g *Grid) SetBuildable(p point.Pointer, buildable bool) {
	m := g.StartRaw.PlacementGrid

	addr, offset := g.GetBitMapAddr(m, p)
	if addr == -1 {
		return
	}
	if buildable {
		m.Data[addr] |= 1 << offset
	} else {
		m.Data[addr] &^= 1 << offset
	}
}

func (g *Grid) IsPathable(p point.Pointer) bool {
	m := g.StartRaw.PathingGrid

	data := g.GetBitMapData(m, p)
	if data == -1 {
		return false
	}
	return data != 0
}

func (g *Grid) IsPathableFast(x, y int) bool {
	m := g.StartRaw.PathingGrid
	addr := (x + y*g.PathingSizeX) / 8
	return m.Data[addr]&(1<<(7-byte(x)%8)) != 0
}

func (g *Grid) SetPathable(p point.Pointer, pathable bool) {
	m := g.StartRaw.PathingGrid

	addr, offset := g.GetBitMapAddr(m, p)
	if addr == -1 {
		return
	}
	if pathable {
		m.Data[addr] |= 1 << offset
	} else {
		m.Data[addr] &^= 1 << offset
	}
}

func (g *Grid) Lock() {
	g.Mutex.Lock()
}

func (g *Grid) Unlock() {
	g.Mutex.Unlock()
}

func (g *Grid) IsCreep(p point.Pointer) bool {
	m := g.MapState.Creep

	data := g.GetBitMapData(m, p)
	if data == -1 {
		return false
	}
	return data != 0
}

func (g *Grid) IsVisible(ptr point.Pointer) bool {
	m := g.MapState.Visibility

	data := g.GetMapData(m, ptr)
	if data == -1 {
		return false
	}
	return data == 2 // uint8. 0=Hidden, 1=Fogged, 2=Visible, 3=FullHidden
}

func (g *Grid) IsExplored(p point.Pointer) bool {
	m := g.MapState.Visibility

	data := g.GetMapData(m, p)
	if data == -1 {
		return false
	}
	return data != 0 // uint8. 0=Hidden, 1=Fogged, 2=Visible, 3=FullHidden
}
