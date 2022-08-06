package main

import (
	"fmt"
	"log"
	"os"
)

type Bus struct {
	Core *GoStationCore
	Bios Access
}

func NewBus(core *GoStationCore, pathToBios string) *Bus {
	bios, err := os.ReadFile(pathToBios)
	if err != nil {
		log.Fatal("Unable to read BIOS: ", err)
	}

	return &Bus{core, NewMemory(bios, 0xbfc00000, 1024*512)}
}

func (bus *Bus) Read8(address uint32) uint8 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read8(address)
	}

	panic(fmt.Sprintf("[Bus::Read8] Invalid address: %x", address))
}

func (bus *Bus) Read16(address uint32) uint16 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read16(address)
	}

	panic(fmt.Sprintf("[Bus::Read16] Invalid address: %x", address))
}

func (bus *Bus) Read32(address uint32) uint32 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read32(address)
	}

	panic(fmt.Sprintf("[Bus::Read32] Invalid address: %x", address))
}

func (bus *Bus) Write8(address uint32, data uint8) {
	//
}

func (bus *Bus) Write16(address uint32, data uint16) {
	//
}

func (bus *Bus) Write32(address uint32, data uint32) {
	panic(fmt.Sprintf("[Bus::Write32] Can't write data into this address: %x", address))
}
