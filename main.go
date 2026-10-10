// Command go-simulator-terrain is a go-simulator plugin that publishes the ground: the elevation
// and the slope of a relief, as fields of the world, so other plugins read them from the world
// instead of from their own files. The host starts it; it talks over standard input and output.
package main

import (
	"log"

	"github.com/akrck02/go-simulator-terrain/terrain"
)

func main() {
	log.SetFlags(0)
	if err := terrain.Plugin().Serve(); err != nil {
		log.Fatal(err)
	}
}
