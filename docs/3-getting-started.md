# 🚀 Getting started

## The folders it needs

The plugin is developed next to Go Simulator and the libraries it uses: `go.mod` replaces them
with the folders beside this one. Clone them side by side:

```text
go-simulator/            the host: runs the rules files
go-simulator-terrain/    this plugin
go-scene-builder/        reads relieve.json reliefs
go-renderer/             used by go-simulator's viewer
```

## Running the example

`examples/ground.yaml` starts the plugin and runs one stage whose only rule is `relief`:

```yaml
plugins:
  - name: terrain
    command: [go, run, .]
    directory: ..

stages:
  - name: ground
    clock: {from: 0, to: 1, step: 1}
    snapshots: [1]
    rules:
      - rule: relief
        with: {path: relieve.json, origin: [0, 0]}
```

Run it from Go Simulator without a window:

```bash
cd ../go-simulator
go run ./cmd/simulate -rules ../go-simulator-terrain/examples/ground.yaml -output results/
```

Relative paths start at the rules file's folder, so `relieve.json` is looked for in `examples/`.

## Your own relief

The example does not come with a relief. Put one next to it, either way:

- **Go Scene Builder's relief**: `relieve.json` and `relieve.f32` side by side, placed on the world
  with `origin`.
- **A NumPy array**: a `.npy` of heights in metres, placed with `minimum_x`, `minimum_z` and
  `cell_size`.

```yaml
- rule: relief
  with: {path: relieve.json, origin: [6, 4.5]}
- rule: relief
  with: {path: relief.npy, minimum_x: -1, minimum_z: -1, cell_size: 0.0125}
```

(One rule or the other, not both.) [Reliefs and fields](4-reliefs-and-fields.md#formats) explains every key, and
the [rules reference](7-rules-reference.md) lists them all.

## What a run writes

At every time a stage lists in `snapshots`, `simulate` writes each field as a grid: a JSON header
and its values as little-endian float32, row by row, the relief format of Go Scene Builder. For the
example, in `results/`:

| File | Content |
|---|---|
| `ground 1 elevation.json` + `.f32` | the heights in metres, the relief's own values |
| `ground 1 slope.json` + `.f32` | the slope of every cell, rise over run |
| `ground 1.geojson` | the entities with a ground shape (none from this plugin) |
| `events.jsonl` | the events (none from this plugin) |

The header gives `columns`, `rows`, `cell` and the `extent` in world units, so the files can be
read back as a relief or drawn as a map, as the pictures in these pages were.

## The tests

```bash
go vet ./...
go test ./...
```

The tests check both formats on small synthetic reliefs, the slope of a plane, and that the fields
reach the world before the first step. For erosion they check that no water is lost or made on the
way to the sea, that hollows drain, that uplift and erosion settle where the stream power law says
(S = (U/K)·A^−m), that creep never overshoots however long the step, that the sea does not move,
that the same seed gives the same valleys, that every rule describes its contracts, and that the
ground, its drainage and its rivers reach the world and change over the steps.

`examples/eroding-island.yaml` runs the erosion on a made-up island that comes with it
(`island.npy`), a million years in 100 steps:

```bash
cd ../go-simulator
go run ./cmd/simulate -rules ../go-simulator-terrain/examples/eroding-island.yaml -output results/
```
