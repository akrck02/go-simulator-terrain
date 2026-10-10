// Package terrain holds the rules of the terrain plugin: it reads a relief and publishes the
// ground as fields of the world (see world.FieldElevation and world.FieldSlope).
package terrain

import (
	"fmt"
	"math"
	"strings"

	"github.com/akrck02/go-scene-builder/relief"
	"github.com/akrck02/go-simulator/world"
)

// Grid is a relief laid on the world: heights in metres at cell centres, rows from the smallest z
// and columns from the smallest x, like a world field.
type Grid struct {
	Columns, Rows int
	Extent        world.Extent
	Heights       []float32
}

// Source says where a relief is and how it lies on the world. A .json relief (go-scene-builder's
// format, kilometres east and north) is placed with Origin: x = east − Origin[0], z = Origin[1] −
// north. A .npy relief (heights in metres, rows from the smallest z) is placed with the centre of
// its first cell and the spacing of its cells, in world units.
type Source struct {
	Path     string
	Origin   [2]float64
	MinimumX float64
	MinimumZ float64
	CellSize float64
}

// Load reads the relief a source names.
func Load(source Source) (*Grid, error) {
	switch {
	case strings.HasSuffix(source.Path, ".json"):
		return loadRelief(source)
	case strings.HasSuffix(source.Path, ".npy"):
		return loadNumpyRelief(source)
	}
	return nil, fmt.Errorf("relief %s: expected a .json relief (with its .f32) or a .npy array", source.Path)
}

// loadRelief reads a go-scene-builder relief; its rows go from the north, which is the smallest z.
func loadRelief(source Source) (*Grid, error) {
	read, err := relief.Load(source.Path)
	if err != nil {
		return nil, err
	}
	extent := world.Extent{
		MinimumX: read.West - source.Origin[0], MaximumX: read.East() - source.Origin[0],
		MinimumZ: source.Origin[1] - read.North, MaximumZ: source.Origin[1] - read.South(),
	}
	return &Grid{Columns: read.Columns, Rows: read.Rows, Extent: extent, Heights: read.Heights}, nil
}

func loadNumpyRelief(source Source) (*Grid, error) {
	if source.CellSize <= 0 {
		return nil, fmt.Errorf("relief %s: cell_size must be positive", source.Path)
	}
	heights, rows, columns, err := readNumpy(source.Path)
	if err != nil {
		return nil, err
	}
	half := source.CellSize / 2
	extent := world.Extent{
		MinimumX: source.MinimumX - half, MaximumX: source.MinimumX - half + float64(columns)*source.CellSize,
		MinimumZ: source.MinimumZ - half, MaximumZ: source.MinimumZ - half + float64(rows)*source.CellSize,
	}
	return &Grid{Columns: columns, Rows: rows, Extent: extent, Heights: heights}, nil
}

// CellSize is the spacing of the cells in world units.
func (grid *Grid) CellSize() float64 {
	return (grid.Extent.MaximumX - grid.Extent.MinimumX) / float64(grid.Columns)
}

// Slopes returns the steepness of every cell, rise over run, with the gradient taken like
// numpy.gradient: central differences inside, one-sided at the edges.
func (grid *Grid) Slopes(metresPerUnit float64) []float32 {
	spacing := grid.CellSize() * metresPerUnit
	slopes := make([]float32, len(grid.Heights))
	for row := range grid.Rows {
		for column := range grid.Columns {
			index := row*grid.Columns + column
			alongX := difference(grid.Heights, index, column, grid.Columns, 1) / spacing
			alongZ := difference(grid.Heights, index, row, grid.Rows, grid.Columns) / spacing
			slopes[index] = float32(math.Hypot(alongX, alongZ))
		}
	}
	return slopes
}

func difference(values []float32, index, position, count, stride int) float64 {
	switch {
	case count < 2:
		return 0
	case position == 0:
		return float64(values[index+stride] - values[index])
	case position == count-1:
		return float64(values[index] - values[index-stride])
	}
	return float64(values[index+stride]-values[index-stride]) / 2
}

// Field returns some values of the grid as a world field.
func (grid *Grid) Field(name string, values []float32) *world.FieldPatch {
	extent := grid.Extent
	return &world.FieldPatch{Name: name, Columns: grid.Columns, Rows: grid.Rows, Extent: &extent, Values: values}
}
