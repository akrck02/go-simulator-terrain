# Rules of the terrain plugin

Version 0.2.0, protocol 1. Publishes the ground of a relief as fields of the world, elevation (m) and slope (m/m), and can evolve it: uplift, rivers and hillslopes.

This reference is generated from what the plugin itself tells the host (the `describe` message), so it always
matches the code. Parameters go under `with:` in a rules file; see go-simulator's
[rules file reference](https://github.com/akrck02/go-simulator/blob/feature/simuladores-genericos/docs/rules-files.md).

| Rule | What it does |
|---|---|
| [`relief`](#relief) | Reads a relief and publishes, at the start, the fields elevation (metres) and slope (rise over run). |
| [`erosion`](#erosion) | Evolves the ground the world already has: uplift, rivers that cut valleys (stream power) and hillslopes that round off (diffusion). Publishes elevation, slope, drainage_area (km²) and the rivers, at the start and as the ground changes. |

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

## erosion

Evolves the ground the world already has: uplift, rivers that cut valleys (stream power) and hillslopes that round off (diffusion). Publishes elevation, slope, drainage_area (km²) and the rivers, at the start and as the ground changes.

- **Reads:** fields `elevation`
- **Writes:** `elevation`, `slope`, `drainage_area`, `river`

| Parameter | Type | Default | Meaning |
|---|---|---|---|
| `years_per_unit` | number | `1` | years of erosion each unit of the stage's clock stands for (10000: every step is ten thousand years) |
| `uplift` | number | `0` | metres a year the land rises |
| `erodibility` | number | `1e-05` | K of the stream power law, a year⁻¹ · m^(1−2m) |
| `area_exponent` | number | `0.5` | m of the stream power law (the slope's exponent is 1) |
| `diffusivity` | number | `0.005` | m² a year: how fast hillslopes round off |
| `sea_level` | number | `0` | metres: lower ground is sea and does not change |
| `river_area` | number | `1` | km² drained from which there is a river |
| `river_width` | list |  | [metres, exponent]: width = metres · area(km²)^exponent; default [3, 0.5] |
| `random_receivers` | boolean | `false` | water goes to a lower neighbour drawn by its slope, not always the steepest: no straight valleys along the grid |
| `seed` | number | `1` | seed of random_receivers: the same seed, the same valleys |
| `publish_every` | number | `1` | steps between publications of the changed ground |
| `metres_per_unit` | number | `1000` | size of a world unit in metres |
