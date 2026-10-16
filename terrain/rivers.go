package terrain

import (
	"math"
	"sort"
)

// A river is the line the water follows down from where enough ground drains into one cell to
// keep a channel (a channel head; Montgomery and Dietrich 1988) to the sea, the edge of the grid
// or a bigger river. Its width grows with the area it drains, as hydraulic geometry finds for
// the discharge (Leopold and Maddock 1953): width = coefficient · area^exponent.

// river is one stretch of a river, in cells.
type river struct {
	head  int   // the cell where it starts: it names the river
	cells []int // from the head down; the last one is in the sea, on the edge or on another river
	area  float64
}

// riversOf traces the rivers of the drainage: every channel head is followed down until the
// water reaches the base level or a river traced before. The rivers draining most are traced
// first, so a main river keeps its whole length.
func (land *landscape) riversOf(routes *drainage, smallestArea float64) []river {
	channel := make([]bool, len(routes.area))
	for _, cell := range routes.order {
		channel[cell] = routes.area[cell] >= smallestArea
	}
	fed := make([]bool, len(routes.area))
	for _, cell := range routes.order {
		if channel[cell] {
			fed[routes.receiver[cell]] = true
		}
	}
	var heads []int
	for _, cell := range routes.order {
		if channel[cell] && !fed[cell] {
			heads = append(heads, cell)
		}
	}
	sort.SliceStable(heads, func(first, second int) bool {
		return land.reachedArea(routes, heads[first]) > land.reachedArea(routes, heads[second])
	})
	traced := make([]bool, len(routes.area))
	var rivers []river
	for _, head := range heads {
		stretch := river{head: head, cells: []int{head}}
		for cell := head; !land.baseLevel[cell] && !traced[cell]; {
			traced[cell] = true
			stretch.area = routes.area[cell]
			cell = routes.receiver[cell]
			stretch.cells = append(stretch.cells, cell)
		}
		if len(stretch.cells) >= 3 {
			rivers = append(rivers, stretch)
		}
	}
	return rivers
}

// reachedArea is the area drained where a channel's water reaches the base level.
func (land *landscape) reachedArea(routes *drainage, cell int) float64 {
	for !land.baseLevel[cell] {
		cell = routes.receiver[cell]
	}
	return routes.area[cell]
}

// riverWidth is how wide a river draining an area is, in metres.
func riverWidth(areaSquareKilometres float64, widthLaw [2]float64) float64 {
	return widthLaw[0] * math.Pow(areaSquareKilometres, widthLaw[1])
}

// smooth rounds the corners of a line of cell centres (Chaikin 1974), keeping both ends.
func smooth(points [][2]float32, passes int) [][2]float32 {
	for range passes {
		if len(points) < 3 {
			return points
		}
		rounded := [][2]float32{points[0]}
		for index := 0; index < len(points)-1; index++ {
			from, to := points[index], points[index+1]
			rounded = append(rounded,
				[2]float32{0.75*from[0] + 0.25*to[0], 0.75*from[1] + 0.25*to[1]},
				[2]float32{0.25*from[0] + 0.75*to[0], 0.25*from[1] + 0.75*to[1]})
		}
		points = append(rounded, points[len(points)-1])
	}
	return points
}
