<p align="center"><img alt="" src="images/logo.svg" width="120"></p>

# Welcome to go-simulator-terrain 👋

**The terrain plugin gives a simulated world its ground**, and lets rivers carve it over thousands of years: valleys, ridges, coasts.

<p align="center"><img alt="A made-up island seen from above, coloured by height" src="images/hero-light.png" width="720"></p>

It reads a relief (a grid of heights) into the world, and its erosion rule lets the ground change: the earth pushes it up, rivers wear it down, slopes round off. Other plugins (climate, flora, towns) read the ground it leaves.

## Is it for me?

- ✅ You use [go-simulator](https://github.com/akrck02/go-simulator) and want **ground shaped by water and time**.
- ✅ You want to **set it up by writing a rules file**: no programming needed.
- ✅ You are curious about **the science** behind it: the pages explain the model with pictures.

## Where to start

Pick the line that sounds like you.

| If you… | Read | Time |
|---|---|---|
| are **new** and want to see it work | 🏴‍☠️ [The quest](2-the-quest.md): an island shaped by its rivers, in 5 levels | 45 min |
| want to **run its example** | [Getting started](3-getting-started.md) | 10 min |
| want to know **how erosion works** | [Erosion and rivers](5-erosion-and-rivers.md) | 15 min |
| are **looking up** a parameter | [Rules reference](7-rules-reference.md) | — |

## All the pages, in reading order

The menu on the left (or the files in `docs/`) follows this order.

**Start**

- [Overview](1-overview.md): this page
- [The quest](2-the-quest.md)
- [Getting started](3-getting-started.md)

**Understand**

- [Reliefs and fields](4-reliefs-and-fields.md)
- [Erosion and rivers](5-erosion-and-rivers.md)
- [How it works](6-how-it-works.md)

**Reference**

- [Rules reference](7-rules-reference.md)

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
  `relieve.f32` next to it (see [getting started](3-getting-started.md#your-own-relief)).
- go-simulator's 3D viewer draws the scene's own ground, not these fields, so there are no
  screenshots of the viewer here; the pictures are 2D drawings of the published fields (see
  [how the pictures were made](6-how-it-works.md#how-the-pictures-were-made)).

## Words you will meet

| Word | What it means |
|---|---|
| Terminal | The window where you type commands: *Terminal* on macOS and Linux, *PowerShell* on Windows. |
| Repository | A project's folder of files, kept on GitHub. `git clone <address>` copies it to your computer. |
| Plugin | A small program that adds one ability to another: here, one kind of simulation each. |
| Rules file | A text file (YAML) that says which plugins to use, with which numbers, for how long: no programming needed. |
| Field | A grid of numbers laid over the world: the height of every spot, its temperature, its rain. |
| Quest | The tutorial of these docs: short levels, each ending with something that works and an achievement. |

## The family

go-simulator-terrain is one of a family of small projects that work together:

| Project | What it does |
|---|---|
| [go-renderer](https://github.com/akrck02/go-renderer) | draws 3D scenes and flat interfaces |
| [go-ui](https://github.com/akrck02/go-ui) | buttons, panels and whole applications |
| [go-ui-android](https://github.com/akrck02/go-ui-android) | Android apps that look like Android's own |
| [go-sound](https://github.com/akrck02/go-sound) | sound and music |
| [go-scene-builder](https://github.com/akrck02/go-scene-builder) | 3D scenes from a flat map |
| [go-simulator](https://github.com/akrck02/go-simulator) | simulations that change a scene |
| **go-simulator-terrain** (this one) | ground shaped by rivers |
| [go-simulator-climate](https://github.com/akrck02/go-simulator-climate) | weather and biomes |
| [go-simulator-flora](https://github.com/akrck02/go-simulator-flora) | forests that grow |
| [go-simulator-urban](https://github.com/akrck02/go-simulator-urban) | towns that grow |
| [go-simulator-motion](https://github.com/akrck02/go-simulator-motion) | a scene that moves |
| [go-3ds](https://github.com/akrck02/go-3ds) | Go on the Nintendo 3DS |

> ⚠️ All of them are **educational experiments**, written with an AI assistant and pending review by their author: good for learning and playing, not for production.
