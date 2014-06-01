package main

import (
	"fmt"
	"github.com/glendc/cgreader"
	"math"
	"strings"
)

type Vector struct {
	x, y int
	icon string
}

func (v Vector) GetMapCoordinates() string {
	return fmt.Sprintf("%d;%d", v.x, v.y)
}

func (v Vector) GetMapIcon() string {
	return v.icon
}

type Ragnarok struct {
	thor, dimensions       Vector
	energy, turn, maxTurns int
	giants                 []Vector
}

var WAIT string = "WAIT"
var STRIKE string = "STRIKE"

func GetDirection(x, y int) <-chan int {
	ch := make(chan int)
	go func() {
		difference := x - y
		switch {
		case difference < 0:
			ch <- -1
		case difference > 0:
			ch <- 1
		default:
			ch <- 0
		}
		close(ch)
	}()
	return ch
}

func GetDirectionLetter(a, b string, v int) string {
	switch v {
	default:
		return ""
	case -1:
		return a
	case 1:
		return b
	}
}

func (ragnarok *Ragnarok) IsPositionAvailable(x, y int) bool {
	for i := range ragnarok.giants {
		if x == ragnarok.giants[i].x && y == ragnarok.giants[i].y {
			return false
		}
	}
	return true
}

func (ragnarok *Ragnarok) RemoveGiant(x, y int) {
	i := 0
	for ; i < len(ragnarok.giants); i++ {
		if x == ragnarok.giants[i].x && y == ragnarok.giants[i].y {
			ragnarok.giants = append(ragnarok.giants[:i], ragnarok.giants[i+1:]...)
			return
		}
	}
}

func GetADL(x, y int) (int, int) {
	switch {
	default:
		return x, y + x
	case x == 0:
		return y * -1, y
	case x == y:
		return 0, y
	}
}

func GetADR(x, y int) (int, int) {
	switch {
	default:
		return x, y - x
	case x == 0:
		return y, y
	case x != y:
		return 0, y
	}
}

func (ragnarok *Ragnarok) MoveGiant(giant, target *Vector) {
	channel_a := GetDirection(target.x, giant.x)
	channel_b := GetDirection(target.y, giant.y)

	dx, dy := <-channel_a, <-channel_b
	x, y := giant.x+dx, giant.y+dy

	for i := 0; i < 2 && !ragnarok.IsPositionAvailable(x, y); i++ {
		if i == 0 {
			dx, dy = GetADL(dx, dy)
		} else {
			dx, dy = GetADR(dx, dy)
		}
		x, y = giant.x+dx, giant.y+dy
	}

	if ragnarok.IsPositionAvailable(x, y) {
		giant.x, giant.y = x, y
	}
}

func (ragnarok *Ragnarok) MoveGiants() {
	for i := range ragnarok.giants {
		ragnarok.MoveGiant(&ragnarok.giants[i], &ragnarok.thor)
	}
}

var THOR_X, THOR_Y, ENERGY, N int

func (ragnarok *Ragnarok) ParseInitialData(ch <-chan string) {
	fmt.Sscanf(
		<-ch,
		"%d %d %d \n",
		&ragnarok.dimensions.x,
		&ragnarok.dimensions.y,
		&ragnarok.maxTurns)

	fmt.Sscanf(
		<-ch,
		"%d %d %d %d \n",
		&ragnarok.energy,
		&ragnarok.thor.x,
		&ragnarok.thor.y,
		&N)

	ENERGY = ragnarok.energy
	THOR_X, THOR_Y = ragnarok.thor.x, ragnarok.thor.y

	ragnarok.giants = make([]Vector, N)

	for i := range ragnarok.giants {
		fmt.Sscanf(
			<-ch,
			"%d %d \n",
			&ragnarok.giants[i].x,
			&ragnarok.giants[i].y)
		ragnarok.giants[i].icon = "G"
	}

	ragnarok.thor.icon = "H"
}

func (ragnarok *Ragnarok) GetInput() (ch chan string) {
	ch = make(chan string)
	go func() {
		ch <- fmt.Sprintf("%d %d", ragnarok.energy, len(ragnarok.giants))
		for _, giant := range ragnarok.giants {
			ch <- fmt.Sprintf("%d %d", giant.x, giant.y)
		}
	}()
	return
}

func Sqrt(x int) int {
	return int(math.Sqrt(float64(x)))
}

func Pow(x int) int {
	return int(math.Pow(float64(x), 2.0))
}

type Position struct {
	x, y int
}

func (ragnarok *Ragnarok) Update(ch <-chan string) string {
	fmt.Sscanf(<-ch, "%d %d", &ENERGY, &N)

	giants := make([]Position, N)
	for i := 0; i < N; i++ {
		fmt.Sscanf(<-ch, "%d %d", &giants[i].x, &giants[i].y)
	}

	x, y := THOR_X, THOR_Y
	td, id := 9999, 0
	dc := 0

	for i, giant := range giants {
		if giant.y > y {
			dc |= 1
		} else if giant.y < y {
			dc |= 2
		}

		if giant.x > x {
			dc |= 4
		} else if giant.x < x {
			dc |= 8
		}

		dx, dy := giant.x-x, giant.y-y
		d := Sqrt(Pow(dx) + Pow(dy))

		if d < 3 {
			return STRIKE
		}

		if d < td {
			id = i
			td = d
		}

	}

	if dc == 15 {
		return WAIT
	}

	chx := GetDirection(giants[id].x, x)
	chy := GetDirection(giants[id].y, y)

	dx, dy := <-chx, <-chy
	THOR_X, THOR_Y = THOR_X+dx, THOR_Y+dy

	return GetDirectionLetter("N", "S", dy) + GetDirectionLetter("W", "E", dx)
}

func (ragnarok *Ragnarok) SetOutput(output string) string {
	ragnarok.MoveGiants()

	var hotspots []Vector
	if output == STRIKE {
		for i := 0; i < 9; i++ {
			x, y := 0, 1
			for u := 0; u < 2; u++ {
				rx, ry := ragnarok.thor.x+(x*i), ragnarok.thor.y+(y*i)
				lx, ly := ragnarok.thor.x-(x*i), ragnarok.thor.y-(y*i)

				ragnarok.RemoveGiant(lx, ly)
				ragnarok.RemoveGiant(rx, ry)

				hotspots = append(hotspots, Vector{lx, ly, "X"})
				hotspots = append(hotspots, Vector{rx, ry, "X"})

				x, y = GetADR(GetADR(x, y))
			}
		}
		ragnarok.energy -= 1
	} else if output != WAIT {
		if strings.Contains(output, "N") {
			ragnarok.thor.y -= 1
		} else if strings.Contains(output, "S") {
			ragnarok.thor.y += 1
		}

		if strings.Contains(output, "E") {
			ragnarok.thor.x += 1
		} else if strings.Contains(output, "W") {
			ragnarok.thor.x -= 1
		}
	}

	ragnarok.turn++

	hotspots = append(hotspots, ragnarok.thor)
	hotspots = append(hotspots, ragnarok.giants...)

	map_info := make([]cgreader.MapObject, len(hotspots))
	for i, v := range hotspots {
		map_info[i] = cgreader.MapObject(v)
	}

	cgreader.DrawMap(
		ragnarok.dimensions.x,
		ragnarok.dimensions.y,
		".",
		map_info...)

	return fmt.Sprintf(
		"Turn = %d\nAmount of Giants = %d\nThor = (%d,%d)\nEnergy = %d",
		ragnarok.turn,
		len(ragnarok.giants),
		ragnarok.thor.x,
		ragnarok.thor.y,
		ragnarok.energy)
}

func (ragnarok *Ragnarok) LoseConditionCheck() bool {
	if ragnarok.energy <= 0 || ragnarok.turn >= ragnarok.maxTurns {
		return true
	}

	x, y := ragnarok.thor.x, ragnarok.thor.y
	dx, dy := ragnarok.dimensions.x, ragnarok.dimensions.y

	for _, giant := range ragnarok.giants {
		if giant.x == x && giant.y == y {
			return true
		}
	}

	if x < 0 || x >= dx || y < 0 || y >= dy {
		return true
	}

	return false
}

func (ragnarok *Ragnarok) WinConditionCheck() bool {
	return len(ragnarok.giants) == 0
}

func main() {
	cgreader.RunTargetProgram("../../input/ragnarok_giants_1.txt", true, &Ragnarok{})
}
