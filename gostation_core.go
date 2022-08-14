package main

import (
	"fmt"
)

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
	cycles uint64
}

func NewGoStation(pathToBios string) *GoStationCore {
	core := GoStationCore{}
	core.Bus = NewBus(&core, pathToBios)
	core.CPU = NewCPU(&core)
	core.GPU = NewGPU(&core)
	core.DMA = NewDMA(&core)
	core.Interrupts = NewInterrupts(&core)
	core.pixels = make([]byte, VRAM_WIDTH*VRAM_HEIGHT*4)
	core.cycles = 0
	return &core
}

func (core *GoStationCore) LoadExecutable(pathToExe string) {
	exe := NewPSXExe(pathToExe)

	// emulate till pc=80030000h
	for core.CPU.pc != 0x80030000 {
		core.Step()
	}

	// copy contents of executable into the main ram
	start := exe.Header.TAddr
	size := exe.Header.TSize
	for i := uint32(0); i < size; i += 1 {
		core.Bus.Write8(start+i, exe.Data[i])
	}

	core.CPU.pc = exe.Header.PC0
	core.CPU.next_pc = exe.Header.PC0 + 4

	// TODO other fields like gp0?

	fmt.Printf("[GoStationCore::LoadExecutable] executable successfully loaded; pc is now in %08x\n", core.CPU.pc)
}

func (core *GoStationCore) UpdateDisplay() {
	for y := 0; y < VRAM_HEIGHT; y += 1 {
		for x := 0; x < VRAM_WIDTH; x += 1 {
			offset := y*VRAM_WIDTH*4 + x*4
			pixel := core.GPU.vram.Read16(uint32(x), uint32(y))

			// after bitmask, shift 3 bits right to upgrade rgb values from 5 bit to 8 bit
			r := uint8(GetRange(uint32(pixel), 0, 5) << 3)
			g := uint8(GetRange(uint32(pixel), 5, 5) << 3)
			b := uint8(GetRange(uint32(pixel), 10, 5) << 3)

			core.pixels[offset] = b
			core.pixels[offset+1] = g
			core.pixels[offset+2] = r
			core.pixels[offset+3] = 255
		}
	}
}

func (core *GoStationCore) Update() {
	for core.cycles < CPU_CYCLES_PER_FRAME {
		core.Step()
	}
	core.cycles = 0

	core.UpdateDisplay()
}

func (core *GoStationCore) Step() {
	// core.CPU.Log(false)
	core.CPU.Step()
	core.cycles += 2
}
