package terrain

import (
	"fmt"

	"github.com/akrck02/go-simulator/plugin"
	"github.com/akrck02/go-simulator/protocol"
	"github.com/akrck02/go-simulator/world"
)

// Version of the plugin.
const Version = "0.1.0"

// defaultMetresPerUnit is the size of a world unit in metres: the scenes are in kilometres.
const defaultMetresPerUnit = 1000

// Plugin returns the terrain plugin.
func Plugin() *plugin.Plugin {
	return &plugin.Plugin{
		Name:    "terrain",
		Version: Version,
		Summary: "Publishes the ground of a relief as fields of the world: elevation (m) and slope (m/m).",
		Rules:   []plugin.Definition{reliefRule()},
	}
}

func reliefRule() plugin.Definition {
	return plugin.Definition{
		Description: protocol.RuleDescription{
			Name:    "relief",
			Summary: "Reads a relief and publishes, at the start, the fields elevation (metres) and slope (rise over run).",
			Parameters: []protocol.ParameterDescription{
				{Name: "path", Type: "text", Summary: "relief: relieve.json (with relieve.f32 next to it) or a .npy array of heights in metres"},
				{Name: "origin", Type: "list", Summary: ".json reliefs: [east, north] of the world's origin; x = east − origin[0], z = origin[1] − north"},
				{Name: "minimum_x", Type: "number", Summary: ".npy reliefs: x of the first column's cell centres (world units)"},
				{Name: "minimum_z", Type: "number", Summary: ".npy reliefs: z of the first row's cell centres (world units)"},
				{Name: "cell_size", Type: "number", Summary: ".npy reliefs: spacing of the cells (world units)"},
				{Name: "metres_per_unit", Type: "number", Default: defaultMetresPerUnit, Summary: "size of a world unit in metres, for the slope"},
			},
			Writes: []string{world.FieldElevation, world.FieldSlope},
		},
		New: newRelief,
	}
}

type reliefRuleInstance struct {
	source        Source
	metresPerUnit float64
	published     bool
}

func newRelief(parameters plugin.Parameters) (plugin.Rule, error) {
	written := parameters.Text("path", "")
	if written == "" {
		return nil, fmt.Errorf("relief needs a path")
	}
	origin := parameters.Vector("origin", [3]float32{})
	source := Source{
		Path:     plugin.Path(written),
		Origin:   [2]float64{float64(origin[0]), float64(origin[1])},
		MinimumX: parameters.Number("minimum_x", 0),
		MinimumZ: parameters.Number("minimum_z", 0),
		CellSize: parameters.Number("cell_size", 0),
	}
	metresPerUnit := parameters.Number("metres_per_unit", defaultMetresPerUnit)
	if metresPerUnit <= 0 {
		return nil, fmt.Errorf("metres_per_unit must be positive")
	}
	return &reliefRuleInstance{source: source, metresPerUnit: metresPerUnit}, nil
}

// Start publishes the ground, so every plugin has it before its first step.
func (rule *reliefRuleInstance) Start(context *plugin.Context) error { return rule.publish(context) }

// Step publishes the ground if Start did not; the ground does not change with time.
func (rule *reliefRuleInstance) Step(context *plugin.Context) error { return rule.publish(context) }

func (rule *reliefRuleInstance) publish(context *plugin.Context) error {
	if rule.published {
		return nil
	}
	rule.published = true
	grid, err := Load(rule.source)
	if err != nil {
		return err
	}
	context.SetField(grid.Field(world.FieldElevation, grid.Heights))
	context.SetField(grid.Field(world.FieldSlope, grid.Slopes(rule.metresPerUnit)))
	return nil
}
