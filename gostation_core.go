package main

type GoStationCore struct {
	Bus        *Bus
	CPU        *CPU
	Interrupts *Interrupts
	DMA        *DMA
}

func NewCore(pathToBios string) *GoStationCore {
	core := GoStationCore{nil, nil, nil, nil}
	core.Bus = NewBus(&core, pathToBios)
	core.CPU = NewCPU(&core)
	core.Interrupts = NewInterrupts(&core)
	core.DMA = NewDMA(&core)
	return &core
}

func (core *GoStationCore) Step() {
	// core.CPU.Log(false)
	core.CPU.Step()
}
