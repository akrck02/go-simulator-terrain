package terrain

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/akrck02/go-simulator/configuration"
	"github.com/akrck02/go-simulator/host"
	"github.com/akrck02/go-simulator/world"
)

// island is a round made-up island: a cone 600 m high and 2 km across, rough enough for water to
// gather in valleys, with the sea around it.
func island(rows, columns int) []float32 {
	heights := make([]float32, rows*columns)
	for row := range rows {
		for column := range columns {
			across, along := float64(column)/float64(columns-1)*2-1, float64(row)/float64(rows-1)*2-1
			distance := math.Hypot(across, along)
			rough := 15 * math.Sin(float64(column)*1.7) * math.Cos(float64(row)*1.3)
			heights[row*columns+column] = float32(600*(0.8-distance) + rough)
		}
	}
	return heights
}

func writeHeights(t *testing.T, path string, rows, columns int, heights []float32) {
	t.Helper()
	header := fmt.Sprintf("{'descr': '<f4', 'fortran_order': False, 'shape': (%d, %d), }", rows, columns)
	for (10+len(header)+1)%64 != 0 {
		header += " "
	}
	header += "\n"
	contents := append([]byte("\x93NUMPY\x01\x00"), byte(len(header)), byte(len(header)>>8))
	contents = append(contents, header...)
	for _, height := range heights {
		contents = binary.LittleEndian.AppendUint32(contents, math.Float32bits(height))
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEveryDropReachesTheBaseLevel(t *testing.T) {
	heights := island(30, 30)
	heights[15*30+15] = -50 // a deep hollow in the middle of the land
	land := newLandscape(30, 30, 100, heights, 0)
	routes := land.drain()
	cellArea, landCells, reached := 100.0*100, 0, 0.0
	for index, isBase := range land.baseLevel {
		if !isBase {
			landCells++
			continue
		}
		reached += routes.area[index] - cellArea
	}
	if math.Abs(reached-float64(landCells)*cellArea) > 1e-6 {
		t.Fatalf("the base level receives %g m², the land is %g m²: water was lost or made", reached, float64(landCells)*cellArea)
	}
	for _, cell := range routes.order {
		steps := 0
		for at := cell; !land.baseLevel[at]; at = routes.receiver[at] {
			if steps++; steps > len(heights) {
				t.Fatalf("the water of cell %d goes round in circles", cell)
			}
		}
	}
}

func TestUpliftAndErosionReachTheirBalance(t *testing.T) {
	// with uplift U and no diffusion, the steady state has S = (U/K)·A^−m in every cell
	// (Whipple and Tucker 1999): the implicit solution must find it
	land := newLandscape(30, 30, 100, island(30, 30), 0)
	settings := erosionSettings{uplift: 1e-3, erodibility: 1e-4, areaExponent: 0.5}
	var routes *drainage
	for range 300 {
		routes = land.evolve(10000, settings)
	}
	for _, cell := range routes.order {
		receiver := routes.receiver[cell]
		distance := land.cellMetres
		if receiver/30 != cell/30 && receiver%30 != cell%30 {
			distance *= math.Sqrt2
		}
		slope := (land.heights[cell] - land.heights[receiver]) / distance
		want := settings.uplift / (settings.erodibility * math.Pow(routes.area[cell], settings.areaExponent))
		if math.Abs(slope-want) > 0.01*want {
			t.Fatalf("cell %d: slope %.5f, the balance of uplift and erosion is %.5f", cell, slope, want)
		}
	}
}

func TestHillslopesRoundOffWithoutOvershooting(t *testing.T) {
	heights := make([]float32, 21*21)
	heights[10*21+10] = 100 // a spike on a flat plain one metre above the sea
	for index := range heights {
		heights[index] += 1
	}
	land := newLandscape(21, 21, 10, heights, 0)
	// a long time in one go: the rule must cut it into stable pieces
	land.evolve(1e6, erosionSettings{diffusivity: 0.01})
	for index, height := range land.heights {
		if height < 1-1e-9 || height > 101 {
			t.Fatalf("cell %d at %g m: diffusion only spreads the spike, between 1 and 101 m", index, height)
		}
	}
	if land.heights[10*21+10] > 50 {
		t.Fatalf("after a million years of creep the spike should have spread: still %g m", land.heights[10*21+10])
	}
}

func TestTheSeaDoesNotChange(t *testing.T) {
	heights := island(20, 20)
	land := newLandscape(20, 20, 100, heights, 0)
	land.evolve(1e5, erosionSettings{uplift: 1e-3, erodibility: 1e-5, areaExponent: 0.5, diffusivity: 0.01})
	for index, height := range heights {
		if land.baseLevel[index] && land.heights[index] != float64(height) {
			t.Fatalf("cell %d of the sea or the edge moved from %g to %g m", index, height, land.heights[index])
		}
	}
}

func TestRiversWidenDownstream(t *testing.T) {
	if got := riverWidth(4, defaultRiverWidth); math.Abs(got-6) > 1e-9 {
		t.Fatalf("a river draining 4 km² is 3 m · 4^0.5 = 6 m wide, not %g", got)
	}
}

func TestTheGroundEvolvesInTheWorld(t *testing.T) {
	folder := t.TempDir()
	writeHeights(t, filepath.Join(folder, "island.npy"), 60, 60, island(60, 60))
	rules, err := configuration.Parse([]byte(`
plugins: [{name: terrain, builtin: true}]
stages:
  - name: ground
    clock: {unit: year, from: 0, to: 5, step: 1}
    rules:
      - rule: relief
        with: {path: island.npy, minimum_x: 0, minimum_z: 0, cell_size: 0.05}
      - rule: erosion
        with: {years_per_unit: 20000, uplift: 0.0005, river_area: 0.05}
`))
	if err != nil {
		t.Fatal(err)
	}
	rules.Directory = folder
	t.Setenv("GO_SIMULATOR_RULES_DIRECTORY", folder)
	engine, err := host.Build(rules, nil, map[string]host.Builtin{"terrain": Plugin().ServeConnection})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	if len(engine.Warnings) > 0 {
		t.Fatalf("the relief and erosion rules should agree on their contracts: %v", engine.Warnings)
	}
	if _, err := engine.Start(); err != nil {
		t.Fatal(err)
	}
	if engine.World.Field(FieldDrainageArea) == nil || len(engine.World.EntitiesOfKind(world.KindRiver)) == 0 {
		t.Fatal("the drainage and the rivers must be in the world before the first step")
	}
	before := append([]float32(nil), engine.World.Field(world.FieldElevation).Values...)
	for !engine.Stage("ground").Finished() {
		if _, err := engine.StepStage("ground"); err != nil {
			t.Fatal(err)
		}
	}
	after := engine.World.Field(world.FieldElevation).Values
	changed := 0
	for index := range after {
		if after[index] != before[index] {
			changed++
		}
	}
	if changed < len(after)/4 {
		t.Fatalf("100 000 years of uplift and erosion should change the land: %d of %d cells changed", changed, len(after))
	}
	for _, river := range engine.World.EntitiesOfKind(world.KindRiver) {
		if len(river.Shape.Line) < 3 || river.Shape.Width <= 0 {
			t.Fatalf("river %s should be a line with a width: %d points, %g wide", river.Identifier, len(river.Shape.Line), river.Shape.Width)
		}
	}
}

func TestDrawnReceiversRepeatAndLoseNoWater(t *testing.T) {
	evolved := func() *landscape {
		land := newLandscape(30, 30, 100, island(30, 30), 0)
		land.random = rand.New(rand.NewPCG(7, 0x7e44a1))
		land.evolve(1e5, erosionSettings{uplift: 1e-3, erodibility: 1e-5, areaExponent: 0.5, diffusivity: 0.01})
		return land
	}
	first, second := evolved(), evolved()
	for index := range first.heights {
		if first.heights[index] != second.heights[index] {
			t.Fatalf("the same seed must give the same ground: cell %d at %g and %g m", index, first.heights[index], second.heights[index])
		}
	}
	routes, landArea, reached := first.drain(), 0.0, 0.0
	for index, isBase := range first.baseLevel {
		if isBase {
			reached += routes.area[index] - 100*100
		} else {
			landArea += 100 * 100
		}
	}
	if math.Abs(reached-landArea) > 1e-6 {
		t.Fatalf("drawn receivers lose or make water: %g m² reach the base level of %g m² of land", reached, landArea)
	}
}
