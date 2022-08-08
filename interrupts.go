package main

const (
	IC_OFFSET = 0x1f801070
	IC_SIZE   = 8
)

type Interrupts struct {
	Core *GoStationCore

	Status uint32
	Mask   uint32
}

func NewInterrupts(core *GoStationCore) *Interrupts {
	return &Interrupts{
		core,
		0,
		0,
	}
}

func (ic *Interrupts) Contains(address uint32) bool {
	return address >= IC_OFFSET && address < (IC_OFFSET+IC_SIZE)
}

func (ic *Interrupts) Write32(address uint32, data uint32) {
	switch address {
	case 0x1f801070:
		ic.Status = data
	case 0x1f801074:
		ic.Mask = data
	}
}
