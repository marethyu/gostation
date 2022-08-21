package main

import (
	"fmt"
	"os"
	"unsafe"

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

	texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_BGR555, sdl.TEXTUREACCESS_STREAMING, VRAM_WIDTH, VRAM_HEIGHT)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
		return 4
	}
	defer texture.Destroy()

	gopsx := NewGoStation("roms/SCPH1001.BIN")
	gopsx.LoadExecutable("roms/tests/psxtest_cpu/psxtest_cpu.exe")
	// gopsx.LoadExecutable("roms/tests/PSX/HelloWorld/16BPP/HelloWorld16BPP.exe")
	// gopsx.LoadExecutable("roms/tests/PSX/GPU/16BPP/RenderTexturePolygon/CLUT4BPP/RenderTexturePolygonCLUT4BPP.exe")
	// gopsx.LoadExecutable("roms/tests/PSX/ImageLoad/ImageLoad.exe")
	// gopsx.LoadExecutable("roms/tests/PSX/GPU/16BPP/RenderPolygon/RenderPolygon16BPP.exe")

	var event sdl.Event
	var running bool = true

	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		gopsx.Update()

		texture.Update(nil, unsafe.Pointer(&gopsx.GPU.vram.buffer[0]), VRAM_WIDTH*2)

		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}

	return 0
}

func main() {
	os.Exit(run())
}
