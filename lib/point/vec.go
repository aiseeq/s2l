package point

import (
	"math"
	"math/cmplx"
)

// Vector directions
type Side int

const (
	E Side = iota
	NE
	N
	NW
	W
	SW
	S
	SE
)

func (a Point) Compas() Side {
	th := cmplx.Phase(complex128(a))
	_, f := math.Modf(1 + th/math.Pi/2 + 1.0/16)
	return Side(f * 8)
}

func (s Side) IsOrthogonal() bool {
	return s == E || s == N || s == W || s == S
}

func (s Side) IsDiagonal() bool {
	return !s.IsOrthogonal()
}
