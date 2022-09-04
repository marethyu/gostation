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
	CPU_CYCLES_PER_SEC = 33868800
)

type GoStation struct {
	Bus        *Bus
	CPU        *CPU
	GPU        *GPU
	DMA        *DMA
	CDROM      *CDROM
	Interrupts *Interrupts

	cycles         uint64
	cyclesPerFrame uint64
	log            bool
}

func NewGoStation(pathToBios string) *GoStation {
	gostation := GoStation{}

	gostation.Bus = NewBus(&gostation, pathToBios)
	gostation.CPU = NewCPU(&gostation)
	gostation.GPU = NewGPU(&gostation)
	gostation.DMA = NewDMA(&gostation)
	gostation.CDROM = NewCDROM(&gostation)
	gostation.Interrupts = NewInterrupts(&gostation)

	gostation.cycles = 0
	gostation.cyclesPerFrame = CPU_CYCLES_PER_SEC / 60 // NTSC mode for default

	return &gostation
}

func (gostation *GoStation) LoadExecutable(pathToExe string) {
	exe := NewPSXExe(pathToExe)

	// emulate till pc=80030000h
	for gostation.CPU.pc != 0x80030000 {
		gostation.Step()
	}

	// copy contents of executable into the main ram
	start := exe.Header.TAddr
	size := exe.Header.TSize
	for i := uint32(0); i < size; i += 1 {
		gostation.CPU.Write8(start+i, exe.Data[i])
	}

	gostation.CPU.pc = exe.Header.PC0
	gostation.CPU.next_pc = exe.Header.PC0 + 4

	gostation.CPU.modifyReg(28, exe.Header.GP0)

	if exe.Header.SAddr != 0 {
		gostation.CPU.modifyReg(29, exe.Header.SAddr+exe.Header.SSize)
		gostation.CPU.modifyReg(30, exe.Header.SAddr+exe.Header.SSize)
	}

	fmt.Printf("[GoStation::LoadExecutable] executable successfully loaded; pc is now in %08x\n", gostation.CPU.pc)
}

func (gostation *GoStation) Update() {
	for gostation.Step() {
	}
}

func (gostation *GoStation) Step() bool {
	if gostation.log {
		gostation.CPU.Log(true)
	}

	gostation.CheckBIOSFunctionCalls(false)
	gostation.CPU.Step()
	gostation.GPU.Step(2)
	gostation.CDROM.Step()

	gostation.cycles += 2 // each instruction takes about 2 cycles

	if gostation.cycles == gostation.cyclesPerFrame {
		gostation.cycles = 0
		return false
	}

	return true
}

/*
https://psx-spx.consoledev.net/kernelbios/#bios-function-summary
*/
func (gostation *GoStation) CheckBIOSFunctionCalls(log bool) {
	switch gostation.CPU.pc {
	case 0xa0: /* A function */
		fn := gostation.CPU.reg(9)
		BIOSAFunction(gostation, fn, log)
	case 0xb0: /* B function */
		fn := gostation.CPU.reg(9)
		BIOSBFunction(gostation, fn, log)
	case 0xc0: /* C function */
		fn := gostation.CPU.reg(9)
		if log {
			fmt.Printf("[GoStation::CheckBIOSFunctionCalls] BIOS C(%02Xh)\n", fn)
		}
	}
}
