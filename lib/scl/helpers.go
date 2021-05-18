package scl

func TimeToLoop(minutes, seconds int) int {
	return int(float64(minutes*60+seconds) * FPS)
}
