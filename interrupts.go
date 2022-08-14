package main

import (
	"fmt"
)

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

func (ic *Interrupts) Read16(address uint32) uint16 {
	switch address {
	case 0x1f801070:
		return uint16(ic.Status & 0xffff)
	case 0x1f801074:
		return uint16(ic.Mask & 0xffff)
	default:
		panic(fmt.Sprintf("[Interrupts::Read16] Invalid address: %x", address))
	}
}

func (ic *Interrupts) Read32(address uint32) uint32 {
	switch address {
	case 0x1f801070:
		return ic.Status
	case 0x1f801074:
		return ic.Mask
	default:
		return 0 // never reached
	}
}

func (ic *Interrupts) Write16(address uint32, data uint16) {
	switch address {
	case 0x1f801070:
		ic.Status = ic.Status & (uint32(data) | 0xffff0000) // when writing to status, 0=Clear Bit, 1=No change
	case 0x1f801074:
		ic.Mask = (ic.Mask & 0xffff0000) | uint32(data)
	default:
		panic(fmt.Sprintf("[Interrupts::Write16] Invalid address: %x", address))
	}
}

func (ic *Interrupts) Write32(address uint32, data uint32) {
	switch address {
	case 0x1f801070:
		ic.Status = ic.Status & data
	case 0x1f801074:
		ic.Mask = data
	}
}
