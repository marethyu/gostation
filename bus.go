package main

import (
	"fmt"
	"log"
	"os"
)

/*
https://psx-spx.consoledev.net/memorymap/

	KUSEG     KSEG0     KSEG1
	00000000h 80000000h A0000000h  2048K  Main RAM (first 64K reserved for BIOS)
	1F000000h 9F000000h BF000000h  8192K  Expansion Region 1 (ROM/RAM)
	1F800000h 9F800000h    --      1K     Scratchpad (D-Cache used as Fast RAM)
	1F801000h 9F801000h BF801000h  8K     I/O Ports
	1F802000h 9F802000h BF802000h  8K     Expansion Region 2 (I/O Ports)
	1FA00000h 9FA00000h BFA00000h  2048K  Expansion Region 3 (SRAM BIOS region for DTL cards)
	1FC00000h 9FC00000h BFC00000h  512K   BIOS ROM (Kernel) (4096K max)
	      FFFE0000h (in KSEG2)     0.5K   Internal CPU control registers (Cache Control)

Notes:
- KUSEG is a physical memory region and KSEG0, KSEG1 and KSEG2 are virtual memory mirrors.
- Notice that all addresses in KUSEG are smaller than KSEG0, etc.

Size of KUSEG: 80000000h bytes
Size of KSEG0: 20000000h bytes (mirrored to the first 20000000h bytes in KUSEG)
Size of KSEG1: 20000000h bytes (mirrored to the first 20000000h bytes in KUSEG)
Size of KSEG2: 40000000h bytes
*/
type Bus struct {
	Core           *GoStation
	Ram            *Memory
	Bios           *Memory
	ScratchPad     *Memory
	MemoryControl1 *MemoryControl1
	SPU            *Memory
	Peripheral     *Memory /* TODO */
	Timer          *Memory /* TODO */
	Expansion1     *Memory
	Expansion2     *Memory /* TODO implement debug uart */
}

func NewBus(core *GoStation, pathToBios string) *Bus {
	bios, err := os.ReadFile(pathToBios)
	if err != nil {
		log.Fatal("Unable to read BIOS: ", err)
	}

	return &Bus{
		core,
		NewMemory(make([]uint8, 2*1024*1024), 0x00000000, 2*1024*1024),
		NewMemory(bios, 0x1fc00000, 1024*512),
		NewMemory(make([]uint8, 0x400), 0x1f800000, 0x400),
		NewMemoryControl1(),
		NewMemory(make([]uint8, 640), 0x1f801c00, 640),
		NewMemory(make([]uint8, 32), 0x1f801040, 32),
		NewMemory(make([]uint8, 3*16), 0x1f801100, 3*16),
		NewMemory(make([]uint8, 1024*512), 0x1f000000, 1024*512),
		NewMemory(make([]uint8, 128), 0x1f802000, 128),
	}
}

func (bus *Bus) Read8(address uint32) uint8 {
	if bus.Bios.Contains(address) {
		return bus.Bios.Read8(address)
	}

	if bus.Ram.Contains(address) {
		return bus.Ram.Read8(address)
	}

	if bus.ScratchPad.Contains(address) {
		return bus.ScratchPad.Read8(address)
	}

	if bus.SPU.Contains(address) {
		return bus.SPU.Read8(address)
	}

	if bus.Peripheral.Contains(address) {
		return bus.Peripheral.Read8(address)
	}

	if bus.Core.CDROM.Contains(address) {
		return bus.Core.CDROM.Read8(address)
	}

	if bus.Expansion1.Contains(address) {
		return bus.Expansion1.Read8(address)
	}

	if bus.Expansion2.Contains(address) {
		return bus.Expansion2.Read8(address)
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

	if bus.ScratchPad.Contains(address) {
		return bus.ScratchPad.Read16(address)
	}

	if bus.SPU.Contains(address) {
		return bus.SPU.Read16(address)
	}

	if bus.Peripheral.Contains(address) {
		if address == 0x1f801044 {
			return 3
		}

		return bus.Peripheral.Read16(address)
	}

	if bus.Timer.Contains(address) {
		if address == 0x1f801120 {
			return 0x000016b0
		}

		return bus.Timer.Read16(address)
	}

	if bus.Expansion1.Contains(address) {
		return bus.Expansion1.Read16(address)
	}

	if bus.Expansion2.Contains(address) {
		return bus.Expansion2.Read16(address)
	}

	if bus.Core.Interrupts.Contains(address) {
		return bus.Core.Interrupts.Read16(address)
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

	if bus.ScratchPad.Contains(address) {
		return bus.ScratchPad.Read32(address)
	}

	if bus.MemoryControl1.Contains(address) {
		return bus.MemoryControl1.Read32(address)
	}

	if bus.SPU.Contains(address) {
		return bus.SPU.Read32(address)
	}

	if bus.Peripheral.Contains(address) {
		return bus.Peripheral.Read32(address)
	}

	if bus.Timer.Contains(address) {
		return bus.Timer.Read32(address)
	}

	if bus.Core.DMA.Contains(address) {
		return bus.Core.DMA.Read32(address)
	}

	if bus.Core.GPU.Contains(address) {
		return bus.Core.GPU.Read32(address)
	}

	if bus.Expansion1.Contains(address) {
		return bus.Expansion1.Read32(address)
	}

	if bus.Expansion2.Contains(address) {
		return bus.Expansion2.Read32(address)
	}

	if bus.Core.Interrupts.Contains(address) {
		return bus.Core.Interrupts.Read32(address)
	}

	panic(fmt.Sprintf("[Bus::Read32] Invalid address: %x", address))
}

/*
TODO what to do if data width changed? e.g. writing 32 bit val then suddenly 16 bit val. treat as 32 bit and pad zeroes?
*/

func (bus *Bus) Write8(address uint32, data uint8) {
	if bus.Ram.Contains(address) {
		bus.Ram.Write8(address, data)
		return
	}

	if bus.ScratchPad.Contains(address) {
		bus.ScratchPad.Write8(address, data)
		return
	}

	if bus.SPU.Contains(address) {
		bus.SPU.Write8(address, data)
		return
	}

	if bus.Peripheral.Contains(address) {
		bus.Peripheral.Write8(address, data)
		return
	}

	if bus.Core.CDROM.Contains(address) {
		bus.Core.CDROM.Write8(address, data)
		return
	}

	if bus.Expansion1.Contains(address) {
		bus.Expansion1.Write8(address, data)
		return
	}

	if bus.Expansion2.Contains(address) {
		if address == 0x1F802041 {
			fmt.Printf("BIOS Trace: %x\n", data)
		}

		bus.Expansion2.Write8(address, data)
		return
	}

	panic(fmt.Sprintf("[Bus::Write8] Can't write data %x into this address: %x", data, address))
}

func (bus *Bus) Write16(address uint32, data uint16) {
	if bus.Ram.Contains(address) {
		bus.Ram.Write16(address, data)
		return
	}

	if bus.ScratchPad.Contains(address) {
		bus.ScratchPad.Write16(address, data)
		return
	}

	if bus.SPU.Contains(address) {
		bus.SPU.Write16(address, data)
		return
	}

	if bus.Peripheral.Contains(address) {
		bus.Peripheral.Write16(address, data)
		return
	}

	if bus.Timer.Contains(address) {
		bus.Timer.Write16(address, data)
		return
	}

	if bus.Expansion1.Contains(address) {
		bus.Expansion1.Write16(address, data)
		return
	}

	if bus.Expansion2.Contains(address) {
		bus.Expansion2.Write16(address, data)
		return
	}

	if bus.Core.Interrupts.Contains(address) {
		bus.Core.Interrupts.Write16(address, data)
		return
	}

	panic(fmt.Sprintf("[Bus::Write16] Can't write data %x into this address: %x", data, address))
}

func (bus *Bus) Write32(address uint32, data uint32) {
	if bus.Ram.Contains(address) {
		bus.Ram.Write32(address, data)
		return
	}

	if bus.ScratchPad.Contains(address) {
		bus.ScratchPad.Write32(address, data)
		return
	}

	if bus.MemoryControl1.Contains(address) {
		bus.MemoryControl1.Write32(address, data)
		return
	}

	if bus.SPU.Contains(address) {
		bus.SPU.Write32(address, data)
		return
	}

	if bus.Peripheral.Contains(address) {
		bus.Peripheral.Write32(address, data)
		return
	}

	if bus.Timer.Contains(address) {
		bus.Timer.Write32(address, data)
		return
	}

	if bus.Core.DMA.Contains(address) {
		bus.Core.DMA.Write32(address, data)
		return
	}

	if bus.Core.GPU.Contains(address) {
		bus.Core.GPU.Write32(address, data)
		return
	}

	if bus.Expansion1.Contains(address) {
		bus.Expansion1.Write32(address, data)
		return
	}

	if bus.Expansion2.Contains(address) {
		bus.Expansion2.Write32(address, data)
		return
	}

	if bus.Core.Interrupts.Contains(address) {
		bus.Core.Interrupts.Write32(address, data)
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
