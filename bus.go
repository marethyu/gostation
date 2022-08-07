package main

import (
	"fmt"
	"log"
	"os"
)

type Bus struct {
	Core           *GoStationCore
	Ram            Access
	Bios           Access
	MemoryControl1 *MemoryControl1
}

func NewBus(core *GoStationCore, pathToBios string) *Bus {
	bios, err := os.ReadFile(pathToBios)
	if err != nil {
		log.Fatal("Unable to read BIOS: ", err)
	}

	return &Bus{
		core,
		NewMemory(make([]uint8, 2*1024*1024), 0xa0000000, 2*1024*1024),
		NewMemory(bios, 0xbfc00000, 1024*512),
		NewMemoryControl1(),
	}
}

func (bus *Bus) Read8(address uint32) uint8 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read8(address)
	}

	if bus.Ram.Contains(address) {
		return bus.Ram.Read8(address)
	}

	panic(fmt.Sprintf("[Bus::Read8] Invalid address: %x", address))
}

func (bus *Bus) Read16(address uint32) uint16 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read16(address)
	}

	if bus.Ram.Contains(address) {
		return bus.Ram.Read16(address)
	}

	panic(fmt.Sprintf("[Bus::Read16] Invalid address: %x", address))
}

func (bus *Bus) Read32(address uint32) uint32 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read32(address)
	}

	if bus.Ram.Contains(address) {
		return bus.Ram.Read32(address)
	}

	panic(fmt.Sprintf("[Bus::Read32] Invalid address: %x", address))
}

func (bus *Bus) Write8(address uint32, data uint8) {
	if bus.Ram.Contains(address) {
		bus.Ram.Write8(address, data)
		return
	}

	panic(fmt.Sprintf("[Bus::Write8] Can't write data %x into this address: %x", data, address))
}

func (bus *Bus) Write16(address uint32, data uint16) {
	if bus.Ram.Contains(address) {
		bus.Ram.Write16(address, data)
		return
	}

	panic(fmt.Sprintf("[Bus::Write16] Can't write data %x into this address: %x", data, address))
}

func (bus *Bus) Write32(address uint32, data uint32) {
	if bus.MemoryControl1.Contains(address) {
		bus.MemoryControl1.Write32(address, data)
		return
	}

	if bus.Ram.Contains(address) {
		bus.Ram.Write32(address, data)
		return
	}

	if address == 0x1f801060 {
		// Ignore the RAM_SIZE register for now
		return
	}

	if address == 0xfffe0130 {
		// Ignore the cache kontrol register for now
		return
	}

	panic(fmt.Sprintf("[Bus::Write32] Can't write data %x into this address: %x", data, address))
}
