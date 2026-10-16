package terrain

import (
	"container/heap"
	"math"
	"math/rand/v2"
)

// A landscape evolves under three processes, as in landscape evolution models (Braun and Willett
// 2013; Tucker and Hancock 2010):
//
//	dz/dt = U − K·A^m·S + κ·∇²z
//
// U lifts the land (metres a year), the stream power term K·A^m·S cuts valleys where much water
// gathers (A, the area drained, in m²; S, the slope towards the cell the water goes to) and
// diffusion κ·∇²z rounds the hillslopes (soil creep). The sea and the edges of the grid are the
// base level: they do not change, and every drop of water ends in one of them.

// fillStep is how much higher than the cell it spills into a cell of a filled hollow is, so water
// always finds a way down (metres).
const fillStep = 1e-4

// neighbours are the eight cells around one, and the distance to each in cells.
var neighbours = [8]struct {
	row, column int
	distance    float64
}{
	{-1, -1, math.Sqrt2}, {-1, 0, 1}, {-1, 1, math.Sqrt2}, {0, -1, 1},
	{0, 1, 1}, {1, -1, math.Sqrt2}, {1, 0, 1}, {1, 1, math.Sqrt2},
}

// landscape is the ground as it evolves: heights in metres, in float64 so that changes of a
// fraction of a millimetre a year add up instead of being lost to rounding.
type landscape struct {
	columns, rows int
	cellMetres    float64
	heights       []float64
	baseLevel     []bool     // the sea and the edges of the grid: they stay as they are
	random        *rand.Rand // when set, water goes to a lower neighbour drawn by its slope (see drain)
}

func newLandscape(columns, rows int, cellMetres float64, heights []float32, seaLevel float64) *landscape {
	land := &landscape{columns: columns, rows: rows, cellMetres: cellMetres,
		heights: make([]float64, len(heights)), baseLevel: make([]bool, len(heights))}
	for index, height := range heights {
		row, column := index/columns, index%columns
		land.heights[index] = float64(height)
		land.baseLevel[index] = float64(height) <= seaLevel || row == 0 || column == 0 || row == rows-1 || column == columns-1
	}
	return land
}

// drainage is where the water of every cell goes and how much ground drains through it.
type drainage struct {
	receiver []int     // the cell the water goes to (itself on the base level)
	order    []int     // cells from the lowest to the highest: each one after its receiver
	area     []float64 // m² drained through every cell, its own included
}

// drain routes the water: hollows are filled (priority flood, Barnes et al. 2014) so the water
// of every cell reaches the base level, each cell gives its water to its steepest neighbour on the
// filled surface, and the area drained adds up from the highest cells down. With a random source,
// the water goes instead to a lower neighbour drawn with a chance that grows with the square of
// its slope: always following the steepest of eight directions cuts straight valleys lined up
// with the grid, which a drawn relief shows as stripes.
func (land *landscape) drain() *drainage {
	filled := append([]float64(nil), land.heights...)
	routes := &drainage{receiver: make([]int, len(filled)), area: make([]float64, len(filled))}
	visited := make([]bool, len(filled))
	queue := &cellQueue{heights: filled}
	for index, isBase := range land.baseLevel {
		routes.receiver[index] = index
		if isBase {
			visited[index] = true
			heap.Push(queue, index)
		}
	}
	for queue.Len() > 0 {
		cell := heap.Pop(queue).(int)
		if !land.baseLevel[cell] {
			routes.order = append(routes.order, cell)
		}
		for _, neighbour := range land.neighboursOf(cell) {
			if visited[neighbour] {
				continue
			}
			visited[neighbour] = true
			filled[neighbour] = math.Max(filled[neighbour], filled[cell]+fillStep)
			heap.Push(queue, neighbour)
		}
	}
	for _, cell := range routes.order {
		if land.random != nil {
			routes.receiver[cell] = land.drawnNeighbour(cell, filled)
		} else {
			routes.receiver[cell] = land.steepestNeighbour(cell, filled)
		}
	}
	cellArea := land.cellMetres * land.cellMetres
	for index := range routes.area {
		routes.area[index] = cellArea
	}
	for position := len(routes.order) - 1; position >= 0; position-- {
		cell := routes.order[position]
		routes.area[routes.receiver[cell]] += routes.area[cell]
	}
	return routes
}

// steepestNeighbour is the neighbour the water of a cell runs to on the filled surface. The
// flood always leaves a lower neighbour (the one that reached the cell), so there is one.
func (land *landscape) steepestNeighbour(cell int, filled []float64) int {
	best, steepest := cell, 0.0
	row, column := cell/land.columns, cell%land.columns
	for _, offset := range neighbours {
		otherRow, otherColumn := row+offset.row, column+offset.column
		if otherRow < 0 || otherColumn < 0 || otherRow >= land.rows || otherColumn >= land.columns {
			continue
		}
		other := otherRow*land.columns + otherColumn
		if slope := (filled[cell] - filled[other]) / offset.distance; slope > steepest {
			best, steepest = other, slope
		}
	}
	return best
}

// drawnNeighbour is a lower neighbour on the filled surface, drawn with a chance proportional to
// the square of its slope.
func (land *landscape) drawnNeighbour(cell int, filled []float64) int {
	var candidates [8]int
	var weights [8]float64
	count, total := 0, 0.0
	row, column := cell/land.columns, cell%land.columns
	for _, offset := range neighbours {
		otherRow, otherColumn := row+offset.row, column+offset.column
		if otherRow < 0 || otherColumn < 0 || otherRow >= land.rows || otherColumn >= land.columns {
			continue
		}
		other := otherRow*land.columns + otherColumn
		if slope := (filled[cell] - filled[other]) / offset.distance; slope > 0 {
			candidates[count], weights[count] = other, slope*slope
			total += weights[count]
			count++
		}
	}
	draw := land.random.Float64() * total
	for index := range count {
		if draw < weights[index] {
			return candidates[index]
		}
		draw -= weights[index]
	}
	return candidates[count-1]
}

func (land *landscape) neighboursOf(cell int) []int {
	row, column := cell/land.columns, cell%land.columns
	found := make([]int, 0, 8)
	for _, offset := range neighbours {
		otherRow, otherColumn := row+offset.row, column+offset.column
		if otherRow >= 0 && otherColumn >= 0 && otherRow < land.rows && otherColumn < land.columns {
			found = append(found, otherRow*land.columns+otherColumn)
		}
	}
	return found
}

// erosionSettings are the numbers of the model, in metres and years.
type erosionSettings struct {
	uplift       float64 // U, metres a year
	erodibility  float64 // K, a year⁻¹ · m^(1−2m)
	areaExponent float64 // m (the slope's exponent n is 1)
	diffusivity  float64 // κ, m² a year
}

// evolve moves the landscape some years on. Each piece of time recomputes the drainage, lifts the
// land, cuts the valleys and rounds the hillslopes; the pieces are short enough for the diffusion
// to stay stable (κ·Δt/Δx² ≤ 0.2). It returns the drainage of the last piece.
func (land *landscape) evolve(years float64, settings erosionSettings) *drainage {
	pieces := 1
	if settings.diffusivity > 0 {
		pieces = max(1, int(math.Ceil(years*settings.diffusivity/(0.2*land.cellMetres*land.cellMetres))))
	}
	piece := years / float64(pieces)
	var routes *drainage
	for range pieces {
		routes = land.drain()
		land.lift(piece, settings.uplift)
		land.cutValleys(routes, piece, settings)
		land.creep(piece, settings.diffusivity)
	}
	return routes
}

func (land *landscape) lift(years, uplift float64) {
	if uplift == 0 {
		return
	}
	for index, isBase := range land.baseLevel {
		if !isBase {
			land.heights[index] += uplift * years
		}
	}
}

// cutValleys solves the stream power term implicitly, from the lowest cells up, each one after the
// cell its water goes to (Braun and Willett 2013): stable for any length of time, and a cell never
// ends lower than its receiver. A cell lower than its receiver lies in a hollow and is left alone:
// the water pools there.
func (land *landscape) cutValleys(routes *drainage, years float64, settings erosionSettings) {
	if settings.erodibility == 0 {
		return
	}
	for _, cell := range routes.order {
		receiver := routes.receiver[cell]
		if land.heights[cell] <= land.heights[receiver] {
			continue
		}
		distance := land.cellMetres
		if receiver/land.columns != cell/land.columns && receiver%land.columns != cell%land.columns {
			distance *= math.Sqrt2
		}
		strength := settings.erodibility * years * math.Pow(routes.area[cell], settings.areaExponent) / distance
		land.heights[cell] = (land.heights[cell] + strength*land.heights[receiver]) / (1 + strength)
	}
}

// creep rounds the hillslopes: linear diffusion, with the four neighbours of every cell.
func (land *landscape) creep(years, diffusivity float64) {
	if diffusivity == 0 {
		return
	}
	factor := diffusivity * years / (land.cellMetres * land.cellMetres)
	before := append([]float64(nil), land.heights...)
	for index, isBase := range land.baseLevel {
		if isBase {
			continue // the base level has no neighbours missing: edge cells are base level
		}
		around := before[index-1] + before[index+1] + before[index-land.columns] + before[index+land.columns]
		land.heights[index] += factor * (around - 4*before[index])
	}
}

// cellQueue gives the lowest cell first, and the first one in the grid among equals, so runs repeat.
type cellQueue struct {
	cells   []int
	heights []float64
}

func (queue *cellQueue) Len() int { return len(queue.cells) }
func (queue *cellQueue) Less(first, second int) bool {
	firstCell, secondCell := queue.cells[first], queue.cells[second]
	if queue.heights[firstCell] != queue.heights[secondCell] {
		return queue.heights[firstCell] < queue.heights[secondCell]
	}
	return firstCell < secondCell
}
func (queue *cellQueue) Swap(first, second int) {
	queue.cells[first], queue.cells[second] = queue.cells[second], queue.cells[first]
}
func (queue *cellQueue) Push(value any) { queue.cells = append(queue.cells, value.(int)) }
func (queue *cellQueue) Pop() any {
	last := queue.cells[len(queue.cells)-1]
	queue.cells = queue.cells[:len(queue.cells)-1]
	return last
}
