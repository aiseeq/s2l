package point

import (
	log "bitbucket.org/aisee/minilog"
	"math"
	"testing"
)

func TestPoint_Compas(t *testing.T) {
	for x := float64(0); x < 360; x++ {
		th := x / 360 * math.Pi * 2
		p := Pt(math.Cos(th), math.Sin(th))
		log.Info(x, p, p.Compas())
	}
}
