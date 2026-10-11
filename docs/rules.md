# Rules of the terrain plugin

Version 0.1.0, protocol 1. Publishes the ground of a relief as fields of the world: elevation (m) and slope (m/m).

This reference is generated from what the plugin itself tells the host (the `describe` message), so it always
matches the code. Parameters go under `with:` in a rules file; see go-simulator's
[rules file reference](https://github.com/akrck02/go-simulator/blob/feature/simuladores-genericos/docs/rules-files.md).

| Rule | What it does |
|---|---|
| [`relief`](#relief) | Reads a relief and publishes, at the start, the fields elevation (metres) and slope (rise over run). |

## relief

Reads a relief and publishes, at the start, the fields elevation (metres) and slope (rise over run).

- **Reads:** nothing from other plugins
- **Writes:** `elevation`, `slope`

| Parameter | Type | Default | Meaning |
|---|---|---|---|
| `path` | text |  | relief: relieve.json (with relieve.f32 next to it) or a .npy array of heights in metres |
| `origin` | list |  | .json reliefs: [east, north] of the world's origin; x = east − origin[0], z = origin[1] − north |
| `minimum_x` | number |  | .npy reliefs: x of the first column's cell centres (world units) |
| `minimum_z` | number |  | .npy reliefs: z of the first row's cell centres (world units) |
| `cell_size` | number |  | .npy reliefs: spacing of the cells (world units) |
| `metres_per_unit` | number | `1000` | size of a world unit in metres, for the slope |
