package main

type GoStationCore struct {
	Bus *Bus
	CPU *CPU
}

func NewCore(pathToBios string) *GoStationCore {
	core := GoStationCore{nil, nil}
	core.Bus = NewBus(&core, pathToBios)
	core.CPU = NewCPU(&core)
	return &core
}

func (core *GoStationCore) Step() {
	core.CPU.Step()
}
