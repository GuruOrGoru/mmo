package objects

import "time"

type Player struct {
	Name          string
	X             float64
	Y             float64
	Radius        float64
	Direction     float64
	Speed         float64
	DbId          int64
	HighScore     int64
	SporeCooldown float64
}

type Spore struct {
	X         float64
	Y         float64
	Radius    float64
	Dropper   *Player
	DroppedAt time.Time
}
