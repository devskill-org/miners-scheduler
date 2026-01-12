// Package main provides an example of using sun calculations for sunrise/sunset times.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/sixdouglas/suncalc"
)

func main() {
	// Get sun position (azimuth and altitude)
	pos := suncalc.GetPosition(time.Now().Add(-time.Hour*2), 56.9496, 24.1052) // NYC
	fmt.Printf("Azimuth: %.2f°, Altitude: %.2f°\n",
		pos.Azimuth*180/math.Pi,
		pos.Altitude*180/math.Pi)

	// Get sunrise/sunset times
	times := suncalc.GetTimes(time.Now(), 24.1052, 24.1052)
	fmt.Println("Sunrise:", times["sunrise"])
	fmt.Println("Sunset:", times["sunset"])
}
