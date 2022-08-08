package main

type GoStationCore struct {
	Bus        *Bus
	CPU        *CPU
	Interrupts *Interrupts
}

func NewCore(pathToBios string) *GoStationCore {
	core := GoStationCore{nil, nil, nil}
	core.Bus = NewBus(&core, pathToBios)
	core.CPU = NewCPU(&core)
	core.Interrupts = NewInterrupts(&core)
	return &core
}

func (core *GoStationCore) Step() {
	core.CPU.Log()
	core.CPU.Step()
}
