package terrain

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/akrck02/go-simulator/configuration"
	"github.com/akrck02/go-simulator/host"
	"github.com/akrck02/go-simulator/world"
)

// ramp is a 3 × 4 relief rising 10 m per cell eastwards and 1 m per row: heights[row][column].
func ramp(row, column int) float32 { return float32(10*column + row) }

func writeNumpy(t *testing.T, path string, rows, columns int) {
	t.Helper()
	header := fmt.Sprintf("{'descr': '<f4', 'fortran_order': False, 'shape': (%d, %d), }", rows, columns)
	for (10+len(header)+1)%64 != 0 {
		header += " "
	}
	header += "\n"
	contents := append([]byte("\x93NUMPY\x01\x00"), byte(len(header)), byte(len(header)>>8))
	contents = append(contents, header...)
	for row := range rows {
		for column := range columns {
			contents = binary.LittleEndian.AppendUint32(contents, math.Float32bits(ramp(row, column)))
		}
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeRelief(t *testing.T, folder string, rows, columns int) string {
	t.Helper()
	header := fmt.Sprintf(`{"columns": %d, "rows": %d, "cell": 0.5, "west": 10, "north": 20, "units": "km"}`, columns, rows)
	var values []byte
	for row := range rows {
		for column := range columns {
			values = binary.LittleEndian.AppendUint32(values, math.Float32bits(ramp(row, column)))
		}
	}
	path := filepath.Join(folder, "relieve.json")
	if err := os.WriteFile(path, []byte(header), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "relieve.f32"), values, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestANumpyReliefLiesOnItsCellCentres(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relief.npy")
	writeNumpy(t, path, 3, 4)
	grid, err := Load(Source{Path: path, MinimumX: -1, MinimumZ: 2, CellSize: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	want := world.Extent{MinimumX: -1.25, MaximumX: 0.75, MinimumZ: 1.75, MaximumZ: 3.25}
	if grid.Columns != 4 || grid.Rows != 3 || grid.Extent != want {
		t.Errorf("grid %d × %d over %+v, want 4 × 3 over %+v", grid.Columns, grid.Rows, grid.Extent, want)
	}
	if grid.Heights[2*4+3] != ramp(2, 3) {
		t.Errorf("last height = %g, want %g", grid.Heights[11], ramp(2, 3))
	}
}

func TestAJSONReliefIsPlacedByItsOrigin(t *testing.T) {
	path := writeRelief(t, t.TempDir(), 3, 4)
	grid, err := Load(Source{Path: path, Origin: [2]float64{11, 19}})
	if err != nil {
		t.Fatal(err)
	}
	// west 10, east 12, north 20, south 18.5: x = east − 11, z = 19 − north
	want := world.Extent{MinimumX: -1, MaximumX: 1, MinimumZ: -1, MaximumZ: 0.5}
	if grid.Extent != want {
		t.Errorf("extent %+v, want %+v", grid.Extent, want)
	}
}

func TestSlopesAreRiseOverRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "relief.npy")
	writeNumpy(t, path, 3, 4)
	grid, err := Load(Source{Path: path, CellSize: 0.01}) // 10 m cells
	if err != nil {
		t.Fatal(err)
	}
	slopes := grid.Slopes(1000)
	want := math.Hypot(10.0/10, 1.0/10) // 10 m east and 1 m along z per 10 m cell
	for index, slope := range slopes {
		if math.Abs(float64(slope)-want) > 1e-6 {
			t.Fatalf("slope of cell %d = %g, want %g everywhere on a plane", index, slope, want)
		}
	}
}

func TestTheGroundReachesTheWorldBeforeTheFirstStep(t *testing.T) {
	folder := t.TempDir()
	writeNumpy(t, filepath.Join(folder, "relief.npy"), 3, 4)
	rules, err := configuration.Parse([]byte(`
plugins: [{name: terrain, builtin: true}]
stages:
  - name: ground
    clock: {from: 0, to: 1, step: 1}
    rules:
      - rule: relief
        with: {path: relief.npy, minimum_x: 0, minimum_z: 0, cell_size: 0.01}
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
	if _, err := engine.Start(); err != nil {
		t.Fatal(err)
	}
	elevation, slope := engine.World.Field(world.FieldElevation), engine.World.Field(world.FieldSlope)
	if elevation == nil || slope == nil {
		t.Fatal("the terrain must publish elevation and slope at the start")
	}
	if elevation.At(3, 2) != ramp(2, 3) {
		t.Errorf("elevation at column 3, row 2 = %g, want %g", elevation.At(3, 2), ramp(2, 3))
	}
}
