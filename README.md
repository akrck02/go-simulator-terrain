# go-simulator-terrain

![The ground of a made-up island, as the other plugins read it](docs/img/terrain_fields.png)


A [go-simulator](https://github.com/akrck02/go-simulator) plugin that publishes the ground: it reads a relief once and puts
it in the world as the fields `elevation` (metres) and `slope`, so the other plugins read the same
ground from the world instead of each one loading its own file.

```yaml
plugins:
  - {name: terrain, command: [go, run, .], directory: ../go-simulator-terrain}
stages:
  - name: ground
    clock: {from: 0, to: 1, step: 1}
    rules:
      - {rule: relief, with: {path: relieve.json, origin: [6, 4.5]}}
```

Documentation in [docs/](docs/README.md): [reliefs and fields](docs/model.md) and the
[rules reference](docs/rules.md).
