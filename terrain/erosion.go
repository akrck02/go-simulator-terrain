package terrain

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/akrck02/go-simulator/plugin"
	"github.com/akrck02/go-simulator/protocol"
	"github.com/akrck02/go-simulator/world"
)

// FieldDrainageArea is the field of the ground that drains through every cell, in km² (0 on the
// base level). It is this plugin's name for it: plugins that want it ask for "drainage_area".
const FieldDrainageArea = "drainage_area"

// Defaults of the erosion rule, from the literature (see docs/erosion.md): the rules file sets them
// for its own place.
const (
	defaultErodibility  = 1e-5  // K, a year⁻¹ for m = 0.5: the middle of what rivers show (Stock and Montgomery 1999; Harel et al. 2016)
	defaultAreaExponent = 0.5   // m, with n = 1: concavity m/n ≈ 0.5 (Whipple and Tucker 1999)
	defaultDiffusivity  = 0.005 // κ, m² a year: soil creep on hillslopes (Fernandes and Dietrich 1997; Roering et al. 1999)
	defaultRiverArea    = 1.0   // km²: channel heads drain from hundredths of a km² to about one (Montgomery and Dietrich 1988)
	riverSmoothing      = 2     // passes that round the corners of the cell-by-cell line
)

// defaultRiverWidth is width = 3 m · area(km²)^0.5: the exponent of hydraulic geometry (Leopold and
// Maddock 1953) and a round coefficient, which each place sets in its rules file.
var defaultRiverWidth = [2]float64{3, 0.5}

func erosionRule() plugin.Definition {
	return plugin.Definition{
		Description: protocol.RuleDescription{
			Name: "erosion",
			Summary: "Evolves the ground the world already has: uplift, rivers that cut valleys (stream power) and hillslopes that round off " +
				"(diffusion). Publishes elevation, slope, drainage_area (km²) and the rivers, at the start and as the ground changes.",
			Parameters: []protocol.ParameterDescription{
				{Name: "years_per_unit", Type: "number", Default: 1.0, Summary: "years of erosion each unit of the stage's clock stands for (10000: every step is ten thousand years)"},
				{Name: "uplift", Type: "number", Default: 0.0, Summary: "metres a year the land rises"},
				{Name: "erodibility", Type: "number", Default: defaultErodibility, Summary: "K of the stream power law, a year⁻¹ · m^(1−2m)"},
				{Name: "area_exponent", Type: "number", Default: defaultAreaExponent, Summary: "m of the stream power law (the slope's exponent is 1)"},
				{Name: "diffusivity", Type: "number", Default: defaultDiffusivity, Summary: "m² a year: how fast hillslopes round off"},
				{Name: "sea_level", Type: "number", Default: 0.0, Summary: "metres: lower ground is sea and does not change"},
				{Name: "river_area", Type: "number", Default: defaultRiverArea, Summary: "km² drained from which there is a river"},
				{Name: "river_width", Type: "list", Summary: "[metres, exponent]: width = metres · area(km²)^exponent; default [3, 0.5]"},
				{Name: "random_receivers", Type: "boolean", Default: false, Summary: "water goes to a lower neighbour drawn by its slope, not always the steepest: no straight valleys along the grid"},
				{Name: "seed", Type: "number", Default: 1.0, Summary: "seed of random_receivers: the same seed, the same valleys"},
				{Name: "publish_every", Type: "number", Default: 1.0, Summary: "steps between publications of the changed ground"},
				{Name: "metres_per_unit", Type: "number", Default: defaultMetresPerUnit, Summary: "size of a world unit in metres"},
			},
			Reads:  world.Interest{Fields: []string{world.FieldElevation}},
			Writes: []string{world.FieldElevation, world.FieldSlope, FieldDrainageArea, world.KindRiver},
		},
		New: newErosion,
	}
}

type erosionRuleInstance struct {
	settings      erosionSettings
	yearsPerUnit  float64
	seaLevel      float64
	riverArea     float64 // m²
	riverWidth    [2]float64
	publishEvery  int
	metresPerUnit float64
	random        bool
	seed          uint64

	extent    world.Extent
	land      *landscape
	steps     int
	published map[string]string // river identifier → signature of what was sent
}

func newErosion(parameters plugin.Parameters) (plugin.Rule, error) {
	rule := &erosionRuleInstance{
		settings: erosionSettings{uplift: parameters.Number("uplift", 0), erodibility: parameters.Number("erodibility", defaultErodibility),
			areaExponent: parameters.Number("area_exponent", defaultAreaExponent), diffusivity: parameters.Number("diffusivity", defaultDiffusivity)},
		yearsPerUnit: parameters.Number("years_per_unit", 1), seaLevel: parameters.Number("sea_level", 0),
		riverArea: parameters.Number("river_area", defaultRiverArea) * 1e6, riverWidth: pairOr(parameters["river_width"], defaultRiverWidth),
		publishEvery: int(parameters.Number("publish_every", 1)), metresPerUnit: parameters.Number("metres_per_unit", defaultMetresPerUnit),
		random: parameters.Boolean("random_receivers", false), seed: uint64(parameters.Number("seed", 1)),
		published: map[string]string{},
	}
	switch {
	case rule.yearsPerUnit <= 0:
		return nil, fmt.Errorf("erosion: years_per_unit must be positive")
	case rule.settings.erodibility < 0 || rule.settings.diffusivity < 0:
		return nil, fmt.Errorf("erosion: erodibility and diffusivity cannot be negative")
	case rule.riverArea <= 0 || rule.publishEvery < 1 || rule.metresPerUnit <= 0:
		return nil, fmt.Errorf("erosion: river_area, publish_every and metres_per_unit must be positive")
	}
	return rule, nil
}

// Start takes the ground the world has and publishes its drainage and rivers, so the other
// plugins have them before their first step.
func (rule *erosionRuleInstance) Start(context *plugin.Context) error {
	elevation := context.World.Field(world.FieldElevation)
	if elevation == nil {
		return fmt.Errorf("erosion: the world has no elevation; put the relief rule before it")
	}
	rule.extent = elevation.Extent
	cellMetres := (elevation.Extent.MaximumX - elevation.Extent.MinimumX) / float64(elevation.Columns) * rule.metresPerUnit
	rule.land = newLandscape(elevation.Columns, elevation.Rows, cellMetres, elevation.Values, rule.seaLevel)
	if rule.random {
		rule.land.random = rand.New(rand.NewPCG(rule.seed, 0x7e44a1))
	}
	rule.publishDrainage(context, rule.land.drain())
	return nil
}

// Step moves the ground on by the years the step stands for.
func (rule *erosionRuleInstance) Step(context *plugin.Context) error {
	if rule.land == nil {
		if err := rule.Start(context); err != nil {
			return err
		}
	}
	routes := rule.land.evolve(context.Step*rule.yearsPerUnit, rule.settings)
	rule.steps++
	if rule.steps%rule.publishEvery != 0 {
		return nil
	}
	grid := rule.grid()
	context.SetField(grid.Field(world.FieldElevation, grid.Heights))
	context.SetField(grid.Field(world.FieldSlope, grid.Slopes(rule.metresPerUnit)))
	rule.publishDrainage(context, routes)
	return nil
}

// grid is the landscape as a relief on the world, in float32 like every field.
func (rule *erosionRuleInstance) grid() *Grid {
	heights := make([]float32, len(rule.land.heights))
	for index, height := range rule.land.heights {
		heights[index] = float32(height)
	}
	return &Grid{Columns: rule.land.columns, Rows: rule.land.rows, Extent: rule.extent, Heights: heights}
}

func (rule *erosionRuleInstance) publishDrainage(context *plugin.Context, routes *drainage) {
	areas := make([]float32, len(routes.area))
	for index, area := range routes.area {
		if !rule.land.baseLevel[index] {
			areas[index] = float32(area / 1e6)
		}
	}
	context.SetField(rule.grid().Field(FieldDrainageArea, areas))
	rule.publishRivers(context, rule.land.riversOf(routes, rule.riverArea))
}

// publishRivers sends the rivers that are new or changed and removes the ones that are gone.
func (rule *erosionRuleInstance) publishRivers(context *plugin.Context, rivers []river) {
	wanted := map[string]bool{}
	for _, stretch := range rivers {
		entity := rule.riverEntity(stretch)
		identifier := string(entity.Identifier)
		wanted[identifier] = true
		signature := fmt.Sprint(entity.Shape.Line, entity.Shape.Width, entity.Properties)
		previous, sent := rule.published[identifier]
		switch {
		case !sent:
			context.Add(entity)
		case previous != signature:
			context.Modify(entity)
		default:
			continue
		}
		rule.published[identifier] = signature
	}
	var gone []string
	for identifier := range rule.published {
		if !wanted[identifier] {
			gone = append(gone, identifier)
		}
	}
	slices.Sort(gone)
	for _, identifier := range gone {
		context.Remove(world.Identifier(identifier))
		delete(rule.published, identifier)
	}
}

func (rule *erosionRuleInstance) riverEntity(stretch river) *world.Entity {
	cell := (rule.extent.MaximumX - rule.extent.MinimumX) / float64(rule.land.columns)
	line := make([][2]float32, 0, len(stretch.cells))
	for _, index := range stretch.cells {
		row, column := index/rule.land.columns, index%rule.land.columns
		line = append(line, [2]float32{float32(rule.extent.MinimumX + (float64(column)+0.5)*cell), float32(rule.extent.MinimumZ + (float64(row)+0.5)*cell)})
	}
	areaSquareKilometres := stretch.area / 1e6
	width := riverWidth(areaSquareKilometres, rule.riverWidth) / rule.metresPerUnit
	return &world.Entity{Identifier: world.Identifier(fmt.Sprintf("terrain-river-%d", stretch.head)), Kind: world.KindRiver,
		Shape:      &world.Shape{Line: smooth(line, riverSmoothing), Width: float32(width), Batch: "terrain_rivers"},
		Properties: map[string]any{"width": width, FieldDrainageArea: areaSquareKilometres}}
}

// pairOr reads two numbers from a parameter, or gives the fallback.
func pairOr(value any, fallback [2]float64) [2]float64 {
	list, isList := value.([]any)
	if !isList || len(list) != 2 {
		return fallback
	}
	first, firstIsNumber := list[0].(float64)
	second, secondIsNumber := list[1].(float64)
	if !firstIsNumber || !secondIsNumber {
		return fallback
	}
	return [2]float64{first, second}
}
