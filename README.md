# go-simulator-terrain

> **Note:** this project was written with an AI assistant and is pending thorough review. See the [AI disclosure](#ai-disclosure).

![The ground of a made-up island, as the other plugins read it](docs/img/terrain_fields.png)

A [go-simulator](https://github.com/akrck02/go-simulator) plugin that publishes the ground: it reads a relief once and puts
it in the world as the fields `elevation` (metres) and `slope`, so the other plugins read the same
ground from the world instead of each one loading its own file. Its `erosion` rule lets that
ground change: the land rises, rivers cut valleys and hillslopes round off, and the rivers it
finds go into the world too.

```yaml
plugins:
  - {name: terrain, command: [go, run, .], directory: ../go-simulator-terrain}
stages:
  - name: ground
    clock: {from: 0, to: 1, step: 1}
    rules:
      - {rule: relief, with: {path: relieve.json, origin: [6, 4.5]}}
```

Documentation in [docs/](docs/README.md): [reliefs and fields](docs/model.md),
[erosion and rivers](docs/erosion.md) and the [rules reference](docs/rules.md).

## AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction.** I decided what to build and how it should
behave; the code, the tests and the documentation were written by the AI. Every commit made this
way carries a `Co-Authored-By: Claude` line (all of them).

**Status: pending thorough review.** The code compiles and its tests pass, but I have not yet
reviewed it line by line. Treat it as a working draft, not as code I vouch for.
