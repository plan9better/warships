package main

import (
	"fmt"
	"math"
)

type Coord struct {
	X int
	Y int
}

type Ship []Coord

type Enemy struct {
	Ships          []Ship
	LastMove       Coord
	LastMoveEffect int
	HasLastMove    bool
}

const (
	hit  = 0
	miss = 1
	sunk = 2
)

func isSame(cord1 Coord, cord2 Coord) bool {
	if cord1.X == cord1.Y && cord2.X == cord2.Y {
		return true
	}
	return false
}

func isAdjacent(cord1 Coord, cord2 Coord) bool {
	if isSame(cord1, cord2) {
		return false
	}

	if math.Abs(float64(cord1.X-cord2.X)) <= 1 && math.Abs(float64(cord1.Y-cord2.Y)) <= 1 {
		return true
	}
	// fmt.Println(math.Abs(float64(cord1.X - cord2.X)))
	// fmt.Println(math.Abs(float64(cord1.Y - cord2.Y)))

	return false
}

func (e *Enemy) LogShot(shot Coord, shotEffect int) {
	e.HasLastMove = true
	e.LastMove = shot
	e.LastMoveEffect = shotEffect
}

func main() {
	var cord1 Coord
	cord1.X = 'A'
	cord1.Y = 2

	var cord2 Coord
	cord2.X = 'A'
	cord2.Y = 3

	fmt.Println(isAdjacent(cord1, cord2))
}
