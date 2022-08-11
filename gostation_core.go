package main

type GoStationCore struct {
	Bus        *Bus
	CPU        *CPU
	GPU        *GPU
	DMA        *DMA
	Interrupts *Interrupts
}

func NewCore(pathToBios string) *GoStationCore {
	core := GoStationCore{}
	core.Bus = NewBus(&core, pathToBios)
	core.CPU = NewCPU(&core)
	core.GPU = NewGPU(&core)
	core.DMA = NewDMA(&core)
	core.Interrupts = NewInterrupts(&core)
	return &core
}

func (core *GoStationCore) Step() {
	// core.CPU.Log(false)
	core.CPU.Step()
}
