package objects

import "math/rand/v2"

const SpawnLimit int = 1000
var getPlayerPosition = func(p *Player) (float64, float64) { return p.X, p.Y }
var getPlayerRadius = func(p *Player) float64 { return p.Radius }
var getSporePosition = func(s *Spore) (float64, float64) { return s.X, s.Y }
var getSporeRadius = func(s *Spore) float64 { return s.Radius }

func SpawnCoords(radius float64, playersToAvoid *SharedCollection[*Player], sporesToAvoid *SharedCollection[*Spore]) (float64, float64) {
	bounds := 5000.0
	const maxTries int = 30

	tries := 0
	for {
		x := bounds * (2*rand.Float64()-1)
		y := bounds * (2*rand.Float64()-1)

		if !isTooClose(x, y, radius, playersToAvoid, getPlayerPosition, getPlayerRadius) && !isTooClose(x, y, radius, sporesToAvoid, getSporePosition, getSporeRadius) {
			return x, y
		}
		tries++
		if tries > maxTries {
			bounds *= 2
			tries = 0
		}
	}
}

func isTooClose[T any](x float64, y float64, radius float64, objects *SharedCollection[T], getPosition func(T) (float64, float64), getRadius func(T) float64) bool {
    // Not too close if there are no objects
    if objects == nil {
        return false
    }

    // Check if any object is too close
    tooClose := false
    objects.ForEach(func(_ uint64, object T) {
        if tooClose {
            return
        }

        objX, objY := getPosition(object)
        objRad := getRadius(object)
        xDst := objX - x
        yDst := objY - y
        dstSq := xDst*xDst + yDst*yDst

        if dstSq <= (radius+objRad)*(radius+objRad) {
            tooClose = true
            return
        }
    })

    return tooClose
}
