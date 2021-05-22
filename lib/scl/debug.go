package scl

import (
	log "bitbucket.org/aisee/minilog"
	"github.com/aiseeq/s2l/lib/grid"
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
	"github.com/aiseeq/s2l/protocol/enums/ability"
)

var Red = api.Color{R: 255, G: 1, B: 44}
var Yellow = api.Color{R: 232, G: 177, B: 12}
var Green = api.Color{R: 1, G: 255, B: 48}
var Blue = api.Color{R: 12, G: 51, B: 232}
var White = api.Color{R: 255, G: 255, B: 255}

func (b *Bot) DebugSend() {
	if len(b.DebugCommands) > 0 {
		if err := b.Client.Debug(api.RequestDebug{
			Debug: b.DebugCommands,
		}); err != nil {
			log.Error(err)
		}
		b.DebugCommands = nil
	}
}

func (b *Bot) DebugAdd(cmd *api.DebugCommand) {
	b.DebugCommands = append(b.DebugCommands, cmd)
}

func (b *Bot) DebugAddBoxes(boxes []*api.DebugBox) {
	b.DebugAdd(&api.DebugCommand{
		Command: &api.DebugCommand_Draw{
			Draw: &api.DebugDraw{
				Boxes: boxes}}})
}

func (b *Bot) DebugAddSpheres(spheres []*api.DebugSphere) {
	b.DebugAdd(&api.DebugCommand{
		Command: &api.DebugCommand_Draw{
			Draw: &api.DebugDraw{
				Spheres: spheres}}})
}

func (b *Bot) DebugAddLines(lines []*api.DebugLine) {
	b.DebugAdd(&api.DebugCommand{
		Command: &api.DebugCommand_Draw{
			Draw: &api.DebugDraw{
				Lines: lines}}})
}

func (b *Bot) DebugAddUnits(unitType api.UnitTypeID, owner api.PlayerID, pos point.Point, qty uint32) {
	b.DebugAdd(&api.DebugCommand{
		Command: &api.DebugCommand_CreateUnit{
			CreateUnit: &api.DebugCreateUnit{
				UnitType: unitType,
				Owner:    owner,
				Pos:      pos.To2D(),
				Quantity: qty,
			}}})
}

func (b *Bot) DebugKillUnits(tags ...api.UnitTag) {
	b.DebugAdd(&api.DebugCommand{
		Command: &api.DebugCommand_KillUnit{
			KillUnit: &api.DebugKillUnit{Tag: tags}}})
}

func (b *Bot) DebugMap() {
	var boxes []*api.DebugBox

	pa := b.Info.StartRaw.PlayableArea
	for y := pa.P0.Y; y <= pa.P1.Y; y++ {
		for x := pa.P0.X; x <= pa.P1.X; x++ {
			p := point.Pt(float64(x), float64(y))
			buildable := b.Grid.IsBuildable(p)
			pathable := b.Grid.IsPathable(p)
			// pathable := b.Grid.IsPathableFast(int(p.X()), int(p.Y()))
			z := b.Grid.HeightAt(p)
			color := &Green
			if !buildable && !pathable {
				color = &Red
			} else if !pathable {
				color = &Blue
			} else if !buildable {
				color = &Yellow
			}
			boxes = append(boxes, &api.DebugBox{
				Color: color,
				Min:   &api.Point{X: float32(x) + 0.25, Y: float32(y) + 0.25, Z: float32(z) - 50},
				Max:   &api.Point{X: float32(x) + 0.75, Y: float32(y) + 0.75, Z: float32(z) + 0.01},
			})
		}
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) DebugSafeGrid(normal, safe *grid.Grid) {
	if normal == nil || safe == nil {
		return
	}

	var boxes []*api.DebugBox

	pa := b.Info.StartRaw.PlayableArea
	for y := pa.P0.Y; y <= pa.P1.Y; y++ {
		for x := pa.P0.X; x <= pa.P1.X; x++ {
			p := point.Pt(float64(x), float64(y))
			pathable := normal.IsPathable(p)
			isSafe := safe.IsPathable(p)
			z := b.Grid.HeightAt(p)
			color := &Green
			if !isSafe && !pathable {
				color = &Red
			} else if !pathable {
				color = &Blue
			} else if !isSafe {
				color = &Yellow
			}
			boxes = append(boxes, &api.DebugBox{
				Color: color,
				Min:   &api.Point{X: float32(x) + 0.25, Y: float32(y) + 0.25, Z: float32(z) - 50},
				Max:   &api.Point{X: float32(x) + 0.75, Y: float32(y) + 0.75, Z: float32(z) + 0.01},
			})
		}
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) DebugRamps() {
	var boxes []*api.DebugBox
	for _, ramp := range b.Ramps.All {
		z := b.Grid.HeightAt(ramp.Top)
		boxes = append(boxes, &api.DebugBox{
			Color: &White,
			Min:   &api.Point{X: float32(ramp.Top.X()) + 0.25, Y: float32(ramp.Top.Y()) + 0.25, Z: float32(z) - 100},
			Max:   &api.Point{X: float32(ramp.Top.X()) + 0.75, Y: float32(ramp.Top.Y()) + 0.75, Z: float32(z) + 0.5},
		})
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) DebugEnemyUnits() {
	var spheres []*api.DebugSphere
	for _, u := range b.Enemies.All {
		spheres = append(spheres, &api.DebugSphere{
			Color: &Red,
			P:     u.Pos,
			R:     u.Radius,
		})
	}
	b.DebugAddSpheres(spheres)
}

func (b *Bot) DebugPath(path point.Points, color api.Color) {
	var boxes []*api.DebugBox
	for _, p := range path {
		z := b.Grid.HeightAt(p)
		boxes = append(boxes, &api.DebugBox{
			Color: &color,
			Min:   &api.Point{X: float32(p.X()) + 0.25, Y: float32(p.Y()) + 0.25, Z: float32(z) - 100},
			Max:   &api.Point{X: float32(p.X()) + 0.75, Y: float32(p.Y()) + 0.75, Z: float32(z) + 0.5},
		})
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) DebugLines(lines point.Lines, color api.Color) {
	var dls []*api.DebugLine
	for _, l := range lines {
		p0 := l.A.To3D()
		p0.Z = float32(b.Grid.HeightAt(l.A) + 0.5)
		p1 := l.B.To3D()
		p1.Z = float32(b.Grid.HeightAt(l.B) + 0.5)
		dls = append(dls, &api.DebugLine{
			Color: &color,
			Line: &api.Line{
				P0: p0,
				P1: p1,
			},
		})
	}
	b.DebugAddLines(dls)
}

func (b *Bot) DebugOrders() {
	var dls []*api.DebugLine
	everything := b.Units.My.All()
	everything.Add(b.Units.AllEnemy.All()...)
	everything.Add(b.Units.Minerals.All()...)
	everything.Add(b.Units.Geysers.All()...)
	everything.Add(b.Units.Neutral.All()...)
	for _, u := range everything {
		if u.IsIdle() {
			continue
		}
		color := &Green
		pos := u.TargetPos()
		if pos == 0 {
			tag := u.TargetTag()
			if tag != 0 {
				if target := everything.ByTag(tag); target != nil {
					color = &Yellow
					pos = target.Point()
				}
			}
		}
		if pos != 0 {
			if u.TargetAbility() == ability.Attack_Attack {
				color = &Red
			}
			p0 := u.Point().To3D()
			p0.Z = float32(b.Grid.HeightAt(u) + 0.5)
			p1 := pos.To3D()
			p1.Z = float32(b.Grid.HeightAt(pos) + 0.5)
			dls = append(dls, &api.DebugLine{
				Color: color,
				Line: &api.Line{
					P0: p0,
					P1: p1,
				},
			})
		}
	}
	b.DebugAddLines(dls)
}

func (b *Bot) DebugPoints(ps ...point.Point) {
	var boxes []*api.DebugBox
	for _, p := range ps {
		z := b.Grid.HeightAt(p)
		boxes = append(boxes, &api.DebugBox{
			Color: &Yellow,
			Min:   &api.Point{X: float32(p.X()) - 0.25, Y: float32(p.Y()) - 0.25, Z: float32(z) - 50},
			Max:   &api.Point{X: float32(p.X()) + 0.25, Y: float32(p.Y()) + 0.25, Z: float32(z) + 0.01},
		})
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) DebugCircles(cs ...point.Circle) {
	var spheres []*api.DebugSphere
	for _, c := range cs {
		p := c.Point.To3D()
		p.Z = float32(b.Grid.HeightAt(c.Point)) + 0.01
		spheres = append(spheres, &api.DebugSphere{
			Color: &Yellow,
			P:     p,
			R:     float32(c.R),
		})
	}
	b.DebugAddSpheres(spheres)
}

func (b *Bot) Debug2x2Buildings(ps ...point.Point) {
	var boxes []*api.DebugBox
	for _, p := range ps {
		z := b.Grid.HeightAt(p)
		p = p.Floor()
		boxes = append(boxes, &api.DebugBox{
			Color: &Yellow,
			Min:   &api.Point{X: float32(p.X()), Y: float32(p.Y()), Z: float32(z) - 100},
			Max:   &api.Point{X: float32(p.X()) + 2, Y: float32(p.Y()) + 2, Z: float32(z) + 0.5},
		})
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) Debug3x3Buildings(ps ...point.Point) {
	var boxes []*api.DebugBox
	for _, p := range ps {
		z := b.Grid.HeightAt(p)
		p = p.Floor()
		boxes = append(boxes, &api.DebugBox{
			Color: &White,
			Min:   &api.Point{X: float32(p.X()) - 1, Y: float32(p.Y()) - 1, Z: float32(z) - 100},
			Max:   &api.Point{X: float32(p.X()) + 2, Y: float32(p.Y()) + 2, Z: float32(z) + 0.5},
		})
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) Debug5x3Buildings(ps ...point.Point) {
	var boxes []*api.DebugBox
	for _, p := range ps {
		z := b.Grid.HeightAt(p)
		p = p.Floor()
		boxes = append(boxes, &api.DebugBox{
			Color: &Green,
			Min:   &api.Point{X: float32(p.X()) - 1, Y: float32(p.Y()) - 1, Z: float32(z) - 100},
			Max:   &api.Point{X: float32(p.X()) + 2, Y: float32(p.Y()) + 2, Z: float32(z) + 0.5},
		}, &api.DebugBox{
			Color: &Green,
			Min:   &api.Point{X: float32(p.X()) + 2, Y: float32(p.Y()) - 1, Z: float32(z) - 100},
			Max:   &api.Point{X: float32(p.X()) + 4, Y: float32(p.Y()) + 1, Z: float32(z) + 0.5},
		})
	}
	b.DebugAddBoxes(boxes)
}

func (b *Bot) DebugWayMap(wpm WaypointsMap, showLines bool) {
	wps := point.Points{}
	lines := point.Lines{}
	for p, ns := range wpm {
		wps = append(wps, p.Point)
		if showLines {
			for _, n := range ns {
				lines.Add(point.Line{A: p.Point, B: n.Point})
			}
		}
	}
	b.DebugPath(wps, White)
	if showLines {
		b.DebugLines(lines, White)
	}
}

func (b *Bot) DebugClusters() {
	lines := point.Lines{}
	for n, cluster := range b.Enemies.Clusters {
		for u1 := range cluster.Units {
			for u2 := range cluster.Units {
				lines.Add(point.Line{A: u1.Point(), B: u2.Point()})
			}
		}
		color := Red
		color.R -= uint32(n) * 10
		b.DebugLines(lines, color)
	}
}

/*for _, ramp := range B.Ramps {
	B.Debug2x2Buildings(B.FindRamp2x2Positions(ramp)...)
	B.Debug3x3Buildings(B.FindRampBarracksPositions(ramp))
}*/

/*start = time.Now()
for x := 1; x < 100; x++ {
	B.Path(B.Ramps.My.Top, B.Ramps.Enemy.Top)
}
log.Info(time.Now().Sub(start))*/
/*path, dist := B.Path(B.Ramps.My.Top, B.EnemyRamp.Top)
log.Info(time.Now().Sub(start), dist, path)
B.DebugPath(path)
B.DebugSend()*/

/*start := time.Now()
paths := B.FindPaths(B.Ramps.My.Top)
log.Info(time.Now().Sub(start), paths)
path := paths.From(B.EnemyRamp.Top)
B.DebugPath(path)
B.DebugSend()*/

/*start := time.Now()
path := B.HomePaths.From(B.EnemyRamp.Top)
log.Info(time.Now().Sub(start))
B.DebugPath(path)
B.DebugSend()*/
