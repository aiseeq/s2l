package point

import (
	"fmt"
	"github.com/aiseeq/s2l/protocol/api"
	"math"
	"math/cmplx"
	"sort"
)

type Pointer interface {
	Point() Point
}

type Point complex128
type Points []Point
type Line struct {
	A, B Point
}
type Lines []Line
type Circle struct {
	Point
	R float64
}

func Pt(x, y float64) Point {
	return Point(complex(x, y))
}

func Pt0() Point {
	return 0
}

func Pt2(a *api.Point2D) Point {
	if a == nil {
		return 0
	}
	return Point(complex(a.X, a.Y))
}

func Pt3(a *api.Point) Point {
	if a == nil {
		return 0
	}
	return Point(complex(a.X, a.Y))
}

func PtI(a *api.PointI) Point {
	if a == nil {
		return 0
	}
	return Point(complex(float64(a.X), float64(a.Y)))
}

func (a Point) Point() Point {
	return a
}

func (a Point) String() string {
	c := complex128(a)
	return fmt.Sprintf("(%.2f, %.2f)", real(c), imag(c))
}

func (a Point) X() float64 {
	return real(complex128(a))
}

func (a Point) Y() float64 {
	return imag(complex128(a))
}

func (a Point) Floor() Point {
	c := complex128(a)
	return Point(complex(math.Floor(real(c)), math.Floor(imag(c))))
}

func (a Point) To2D() *api.Point2D {
	c := complex128(a)
	return &api.Point2D{X: float32(real(c)), Y: float32(imag(c))}
}

func (a Point) To3D() *api.Point {
	c := complex128(a)
	return &api.Point{X: float32(real(c)), Y: float32(imag(c))}
}

func (a Point) Add(x, y float64) Point {
	c := complex128(a)
	return Point(complex(real(c)+x, imag(c)+y))
}

func (a Point) Mul(x float64) Point {
	c := complex128(a)
	return Point(complex(real(c)*x, imag(c)*x))
}

func (a Point) Dist(ptr Pointer) float64 {
	b := ptr.Point()
	return cmplx.Abs(complex128(b) - complex128(a))
}

func (a Point) Dist2(ptr Pointer) float64 {
	b := ptr.Point()
	c := complex128(b - a)
	r := real(c)
	i := imag(c)
	return r*r + i*i
}

func (a Point) Manhattan(ptr Pointer) float64 {
	b := ptr.Point()
	x := complex128(b) - complex128(a)
	return real(x) + imag(x)
}

// Direction on grid
func (a Point) Dir(ptr Pointer) Point {
	b := ptr.Point()
	c := complex128(b - a)
	return Point(complex(math.Copysign(1, real(c)), math.Copysign(1, imag(c))))
}

func (a Point) Rotate(rad float64) Point {
	r, th := cmplx.Polar(complex128(a))
	th += rad
	return Point(cmplx.Rect(r, th))
}

func (a Point) Len() float64 {
	return cmplx.Abs(complex128(a))
}

func (a Point) Norm() Point {
	if l := a.Len(); l > 0 {
		return a / Point(complex(l, 0))
	} else {
		return 0
	}
}

func (a Point) Towards(ptr Pointer, offset float64) Point {
	b := ptr.Point()
	return a + (b-a).Norm()*Point(complex(offset, 0))
}

func (a Point) Neighbours4(offset float64) Points {
	return Points{a.Add(0, offset), a.Add(offset, 0), a.Add(0, -offset), a.Add(-offset, 0)}
}

func (a Point) NeighboursDiagonal4(offset float64) Points {
	return Points{a.Add(offset, offset), a.Add(offset, -offset), a.Add(-offset, -offset), a.Add(-offset, offset)}
}

func (a Point) Neighbours8(offset float64) Points {
	return append(a.Neighbours4(offset), a.NeighboursDiagonal4(offset)...)
}

// User should check that he receives not nil
func (a Point) Closest(ps []Point) *Point {
	var closest *Point
	var dist = math.Inf(1)
	for _, b := range ps {
		if closest == nil || dist > a.Dist2(b) {
			p := b
			closest = &p
			dist = a.Dist2(*closest)
		}
	}
	return closest
}

func (a Point) IsCloserThan(dist float64, ptr Pointer) bool {
	b := ptr.Point()
	return a.Dist2(b) < dist*dist
}

func (a Point) IsFurtherThan(dist float64, ptr Pointer) bool {
	b := ptr.Point()
	return a.Dist2(b) > dist*dist
}

func (a Point) S2x2Fix() Point {
	return a
}

func (a Point) CellCenter() Point {
	return a
}

func (ps *Points) Add(p ...Point) {
	*ps = append(*ps, p...)
}

func (ps *Points) Remove(point Point) {
	for k, p := range *ps {
		if p == point {
			if len(*ps) > k+1 {
				*ps = append((*ps)[:k], (*ps)[k+1:]...)
			} else {
				*ps = (*ps)[:k] // Remove last
			}
		}
	}
}

func (ps Points) Len() int {
	return len(ps)
}

func (ps Points) Empty() bool {
	return len(ps) == 0
}

func (ps Points) Exists() bool {
	return len(ps) > 0
}

func (ps Points) Has(p Point) bool {
	for _, pt := range ps {
		if p == pt {
			return true
		}
	}
	return false
}

func (ps Points) Intersect(ps2 Points) Points {
	res := Points{}
	pmap := map[Point]bool{}
	for _, pt := range ps {
		pmap[pt] = true
	}
	for _, pt := range ps2 {
		if pmap[pt] {
			res.Add(pt)
		}
	}
	return res
}

func (ps Points) Center() Point {
	var sum Point
	if len(ps) == 0 {
		return 0
	}
	for _, p := range ps {
		sum += p
	}
	return sum.Mul(1.0 / float64(len(ps)))
}

func (ps Points) ClosestTo(ptr Pointer) Point {
	point := ptr.Point()
	var closest Point
	for _, p := range ps {
		if closest == 0 || point.Dist2(closest) > point.Dist2(p) {
			closest = p
		}
	}
	return closest
}

func (ps Points) FurthestTo(ptr Pointer) Point {
	point := ptr.Point()
	var furthest Point
	for _, p := range ps {
		if furthest == 0 || point.Dist2(furthest) < point.Dist2(p) {
			furthest = p
		}
	}
	return furthest
}

func (ps Points) CloserThan(dist float64, ptr Pointer) Points {
	pos := ptr.Point()
	dist2 := dist * dist
	closer := Points{}
	for _, p := range ps {
		if p.Dist2(pos) <= dist2 {
			closer.Add(p)
		}
	}
	return closer
}

func (ps Points) OrderByDistanceTo(ptr Pointer, desc bool) {
	// todo: optimize? (via sort by other)
	pos := ptr.Point()
	sort.Slice(ps, func(i, j int) bool {
		return desc != (ps[i].Dist2(pos) < ps[j].Dist2(pos))
	})
}

func (ps Points) FirstFurtherThan(dist float64, from Pointer) Point {
	dist2 := dist * dist
	for _, p := range ps {
		if p.Dist2(from) >= dist2 {
			return p
		}
	}
	return 0
}

func (ls *Lines) Add(l ...Line) {
	*ls = append(*ls, l...)
}

func NewCircle(x, y, r float64) *Circle {
	return &Circle{Point(complex(x, y)), r}
}

func PtCircle(p *Point, r float64) *Circle {
	return &Circle{*p, r}
}

// Find the intersection of the two circles, the number of intersections may have 0, 1, 2
func Intersect(a *Circle, b *Circle) (ps Points) {
	if a.X() > b.X() {
		return Intersect(b, a) // Try to fix intersection bug. Looks like, first circle should be on the left
	}
	dx, dy := b.X()-a.X(), b.Y()-a.Y()
	lr := a.R + b.R                //radius and
	dr := math.Abs(a.R - b.R)      //radius difference
	ab := math.Sqrt(dx*dx + dy*dy) //center distance
	if ab <= lr && ab > dr {
		theta1 := math.Atan(dy / dx)
		ef := lr - ab
		ao := a.R - ef/2
		theta2 := math.Acos(ao / a.R)
		theta := theta1 + theta2
		xc := a.X() + a.R*math.Cos(theta)
		yc := a.Y() + a.R*math.Sin(theta)
		ps = append(ps, Pt(xc, yc))
		if ab < lr { //two intersections
			theta3 := math.Acos(ao / a.R)
			theta = theta3 - theta1
			xd := a.X() + a.R*math.Cos(theta)
			yd := a.Y() - a.R*math.Sin(theta)
			ps = append(ps, Pt(xd, yd))
		}
	}
	return
}
