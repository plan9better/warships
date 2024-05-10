package gameclient

import "testing"

func Test_isAdjacent(t *testing.T) {
	type args struct {
		cord1 Coord
		cord2 Coord
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{name: "adj", args: args{cord1: Coord{X: 'A', Y: 2}, cord2: Coord{X: 'A', Y: 3}}, want: true},
		{name: "adj", args: args{cord1: Coord{X: 'B', Y: 10}, cord2: Coord{X: 'B', Y: 9}}, want: true},
		{name: "adj", args: args{cord1: Coord{X: 'C', Y: 6}, cord2: Coord{X: 'B', Y: 5}}, want: true},
		{name: "adj", args: args{cord1: Coord{X: 'A', Y: 2}, cord2: Coord{X: 'A', Y: 3}}, want: true},
		{name: "not adj", args: args{cord1: Coord{X: 'A', Y: 1}, cord2: Coord{X: 'C', Y: 1}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAdjacent(tt.args.cord1, tt.args.cord2); got != tt.want {
				t.Errorf("isAdjacent() = %v, want %v", got, tt.want)
			}
		})
	}
}
