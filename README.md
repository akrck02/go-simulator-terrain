# go-simulator-terrain

A [go-simulator](../go-simulator) plugin that publishes the ground. It reads a relief once and
puts it in the world as fields, so the other plugins (climate, urban growth, vegetation…) read the
same ground from the world instead of each one loading its own file.

| Rule | What it does | Parameters |
|---|---|---|
| `relief` | Publishes, at the start, the fields `elevation` (metres) and `slope` (rise over run, like `numpy.gradient`) | see below |

Fields follow go-simulator's vocabulary: values at cell centres, rows from the smallest z, columns
from the smallest x, heights in metres.

## Reliefs

Two formats:

```yaml
# go-scene-builder's relief: relieve.json with relieve.f32 next to it, kilometres east and north.
# origin places it: x = east − origin[0], z = origin[1] − north
- rule: relief
  with: {path: relieve.json, origin: [6, 4.5]}

# a NumPy array of heights in metres, rows from the smallest z, placed by its first cell centre
# and the spacing of its cells (world units)
- rule: relief
  with: {path: relief.npy, minimum_x: -1, minimum_z: -1, cell_size: 0.0125}
```

`metres_per_unit` (1000: the scenes are in kilometres) is the size of a world unit, for the
slope. Relative paths start at the rules file's folder.

## Use

```bash
cd ../go-simulator
go run ./cmd/simulate -rules ../go-simulator-terrain/examples/ground.yaml -output results/
```

writes `ground 1 elevation.f32` and `ground 1 slope.f32` (with their JSON headers) in `results/`.
In a rules file, put the terrain plugin first, so its fields are in the world before the plugins
that read them start.

## Tests

```bash
go test ./...
```
