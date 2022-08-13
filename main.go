package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var texture *sdl.Texture
	var err error

	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()

	window, err = sdl.CreateWindow("GOSTATION", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		VRAM_WIDTH, VRAM_HEIGHT, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer renderer.Destroy()

	texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, VRAM_WIDTH, VRAM_HEIGHT)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		return 4
	}
	defer texture.Destroy()

	gopsx := NewGoStation("SCPH1001.BIN")

	var event sdl.Event
	var running bool = true

	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		gopsx.Step()

		texture.Update(nil, gopsx.pixels, int(VRAM_WIDTH)*4)

		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}

	return 0
}

func main() {
	os.Exit(run())
}
