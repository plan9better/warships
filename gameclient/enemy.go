package gameclient

import (
	"math"
)

type Coord struct {
	X int
	Y int
}

type Ship []Coord

type Player struct {
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

func IsSame(cord1 Coord, cord2 Coord) bool {
	if cord1.X == cord1.Y && cord2.X == cord2.Y {
		return true
	}
	return false
}

func IsAdjacent(cord1 Coord, cord2 Coord) bool {
	if IsSame(cord1, cord2) {
		return false
	}

	if math.Abs(float64(cord1.X-cord2.X)) <= 1 && math.Abs(float64(cord1.Y-cord2.Y)) <= 1 {
		return true
	}
	return false
}

func isOnEdge(coord Coord) bool {
	if coord.Y == 1 || coord.Y == 10 || coord.X == 'A' || coord.Y == 'J' {
		return true
	}
	return false
}

func FindAdjacent(coord Coord) []Coord {
	var res []Coord
	if coord.X == 'A' {
		// For A1

		// B1
		res = append(res, Coord{coord.X + 1, coord.Y})

		// B2
		res = append(res, Coord{coord.X + 1, coord.Y + 1})

		// A2
		res = append(res, Coord{coord.X, coord.Y + 1})
	}
	return nil
}

func (p *Player) LogShot(shot Coord, shotEffect int) {
	p.HasLastMove = true
	p.LastMove = shot
	p.LastMoveEffect = shotEffect
}
