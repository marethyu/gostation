package main

/*
Timing:

	CPU Clock   =  33.868800MHz (44100Hz*300h)
	Video Clock =  53.222400MHz (44100Hz*300h*11/7)
*/
const (
	CPU_CYCLES_PER_FRAME = 33868800 / 60 /* for now use NTSC mode (refresh display about 60 times per second) */
)

type GoStationCore struct {
	Bus        *Bus
	CPU        *CPU
	GPU        *GPU
	DMA        *DMA
	Interrupts *Interrupts

	pixels []byte
}

func NewGoStation(pathToBios string) *GoStationCore {
	core := GoStationCore{}
	core.Bus = NewBus(&core, pathToBios)
	core.CPU = NewCPU(&core)
	core.GPU = NewGPU(&core)
	core.DMA = NewDMA(&core)
	core.Interrupts = NewInterrupts(&core)
	core.pixels = make([]byte, VRAM_WIDTH*VRAM_HEIGHT*4)
	return &core
}

func (core *GoStationCore) Step() {
	for i := 1; i < CPU_CYCLES_PER_FRAME; i++ {
		// core.CPU.Log(false)
		core.CPU.Step()
	}

	core.UpdateDisplay()
}

func (core *GoStationCore) UpdateDisplay() {
	for y := 0; y < VRAM_HEIGHT; y += 1 {
		for x := 0; x < VRAM_WIDTH; x += 1 {
			offset := y*VRAM_WIDTH*4 + x*4
			pixel := core.GPU.vram.Read16(uint32(x), uint32(y))

			// after bitmask, shift 3 bits right to upgrade rgb values from 5 bit to 8 bit
			r := uint8(pixel&0b11111) << 3
			g := uint8((pixel>>5)&0b11111) << 3
			b := uint8((pixel>>10)&0b11111) << 3

			core.pixels[offset] = b
			core.pixels[offset+1] = g
			core.pixels[offset+2] = r
			core.pixels[offset+3] = 255
		}
	}
}
