# 📚 go-simulator-terrain documentation

| | Page | What is in it |
|---|---|---|
| 🚀 | [Getting started](getting-started.md) | The folders it needs, running the example, using your own relief, the files a run writes, and the tests. |
| 🏔️ | [Reliefs and fields](model.md) | What it publishes and where the values sit, how the slope is measured, the two relief formats and how each lies on the world, with pictures. |
| 🌊 | [Erosion and rivers](erosion.md) | How the `erosion` rule lets the ground change: uplift, valleys, hillslopes, the rivers it publishes and what a year means. |
| 🔍 | [How it works](how-it-works.md) | How the plugin fits go-simulator, what it does at the start and at every step, and how the pictures were made. |
| 📋 | [Rules reference](rules.md) | The `relief` and `erosion` rules and every parameter, generated from the plugin's own description. |

New here? Start with the [project README](../README.md).

## 🚧 Not done yet

- The `erosion` rule only cuts and rounds: no sediment is laid down (no deltas, alluvial fans or
  filled valleys), uplift is the same everywhere, and every rock gives way alike (one
  `erodibility` for the whole map). Without it, the relief is read and published once.
- No aspect (which way a slope faces) and no catchments; the flow itself only shows as
  `drainage_area` and the rivers.
- `.npy` reliefs must be two-dimensional little-endian float32 or float64 arrays in C order;
  Fortran-ordered arrays are refused.
- The cells are square: a relief with different spacings east and north cannot be read.
- `examples/ground.yaml` does not come with a relief: put your own `relieve.json` and
  `relieve.f32` next to it (see [getting started](getting-started.md#your-own-relief)).
- go-simulator's 3D viewer draws the scene's own ground, not these fields, so there are no
  screenshots of the viewer here; the pictures are 2D drawings of the published fields (see
  [how the pictures were made](how-it-works.md#how-the-pictures-were-made)).
