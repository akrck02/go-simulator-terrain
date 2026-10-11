# Reliefs and fields

Almost every simulator needs to know the shape of the land: the climate to know where it is cold,
the forests to know where it is too steep, the town to know where it can build. Instead of each one
loading its own copy of the relief (and maybe disagreeing about it), the terrain plugin reads it
once and puts it in the world for everyone.

![A made-up island: its elevation and its slope, as the terrain plugin publishes them](img/terrain_fields.png)

*The example island as the other plugins see it: the height of every cell, and how steep it is.*

Put it first in the rules file: plugins start in order, so the ones after it find the ground
already there.

## What it publishes

At the start, two fields of the [vocabulary](https://github.com/akrck02/go-3d-simulator/blob/feature/simuladores-genericos/docs/vocabulary.md):

| Field | Unit | How |
|---|---|---|
| `elevation` | metres | the relief's heights, untouched |
| `slope` | rise over run | the gradient like `numpy.gradient`: central differences inside, one-sided at the edges |

Values sit at cell centres, rows from the smallest z and columns from the smallest x. The ground
does not change with time, so later steps publish nothing.

## Formats

**go-scene-builder's relief**: `relieve.json`, `{"columns", "rows", "cell", "west", "north",
"units"}`, with the heights in metres as little-endian float32 in `relieve.f32` next to it, row by
row from the north. Its coordinates are kilometres east and north; `origin` places it on the
world: `x = east − origin[0]`, `z = origin[1] − north`. Rows go from the north, which is the
smallest z, so no row is flipped.

```yaml
- rule: relief
  with: {path: relieve.json, origin: [6, 4.5]}
```

**NumPy array**: a two-dimensional `.npy` of float32 or float64 heights in metres, rows from the
smallest z, placed by the centre of its first cell and the spacing of its cells, in world units.

```yaml
- rule: relief
  with: {path: relief.npy, minimum_x: -1, minimum_z: -1, cell_size: 0.0125}
```

`metres_per_unit` (1000 for scenes in kilometres) turns the cell size into metres for the slope.

## Exactness

The field's extent is float64, so a plugin that rebuilds the grid from the field gets the same
cell size and cell centres as the file. This matters for simulations that are sensitive to their
inputs: a tenth of a millimetre of difference in the heights can grow a different town.

## Use and tests

```bash
cd ../go-simulator
go run ./cmd/simulate -rules ../go-simulator-terrain/examples/ground.yaml -output results/
cd ../go-simulator-terrain && go test ./...
```

The tests check both formats on small synthetic reliefs, the slope of a plane, and that the fields
reach the world before the first step.
