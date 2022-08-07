package main

import (
	"fmt"
)

const (
	MC1_OFFSET = 0x1f801000
	MC1_SIZE   = 0x24
)

/*
https://psx-spx.consoledev.net/iomap/#memory-control-1

	1F801000h 4    Expansion 1 Base Address (usually 1F000000h)
	1F801004h 4    Expansion 2 Base Address (usually 1F802000h)
	1F801008h 4    Expansion 1 Delay/Size (usually 0013243Fh; 512Kbytes 8bit-bus)
	1F80100Ch 4    Expansion 3 Delay/Size (usually 00003022h; 1 byte)
	1F801010h 4    BIOS ROM    Delay/Size (usually 0013243Fh; 512Kbytes 8bit-bus)
	1F801014h 4    SPU_DELAY   Delay/Size (usually 200931E1h)
	1F801018h 4    CDROM_DELAY Delay/Size (usually 00020843h or 00020943h)
	1F80101Ch 4    Expansion 2 Delay/Size (usually 00070777h; 128-bytes 8bit-bus)
	1F801020h 4    COM_DELAY / COMMON_DELAY (00031125h or 0000132Ch or 00001325h)
*/
type MemoryControl1 struct {
	exp1_base_addr uint32
	exp2_base_addr uint32
	exp1_delay     uint32
	exp3_delay     uint32
	bios_delay     uint32
	spu_delay      uint32
	cdrom_delay    uint32
	exp2_delay     uint32
	common_delay   uint32
}

func NewMemoryControl1() *MemoryControl1 {
	return &MemoryControl1{
		0x1f000000,
		0x1f802000,
		0x0013243f,
		0x00003022,
		0x0013243f,
		0x200931e1,
		0x00020843,
		0x00070777,
		0x00031125,
	}
}

func (mc1 *MemoryControl1) Contains(address uint32) bool {
	return address >= MC1_OFFSET && address < (MC1_OFFSET+MC1_SIZE)
}

func (mc1 *MemoryControl1) Write32(address uint32, data uint32) {
	switch address {
	case 0x1f801000:
		if mc1.exp1_base_addr != data {
			/* cuz the base addresses are fixed */
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad expansion 1 base address: %x", data))
		}
	case 0x1f801004:
		if mc1.exp2_base_addr != data {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad expansion 2 base address: %x", data))
		}
	case 0x1f801008:
		if mc1.exp1_delay != data {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad expansion 1 delay value: %x", data))
		}
	case 0x1f80100c:
		if mc1.exp3_delay != data {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad expansion 3 delay value: %x", data))
		}
	case 0x1f801010:
		if mc1.bios_delay != data {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad BIOS delay value: %x", data))
		}
	case 0x1f801014:
		if mc1.spu_delay != data {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad SPU delay value: %x", data))
		}
	case 0x1f801018:
		if data != 0x00020843 && data != 0x00020943 {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad CDROM delay value: %x", data))
		}
		mc1.cdrom_delay = data
	case 0x1f80101c:
		if mc1.exp2_delay != data {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad expansion 2 delay value: %x", data))
		}
	case 0x1f801020:
		if data != 0x00031125 && data != 0x0000132c && data != 0x00001325 {
			panic(fmt.Sprintf("[MemoryControl1::Write32] Bad common delay value: %x", data))
		}
		mc1.common_delay = data
	default:
		panic(fmt.Sprintf("[MemoryControl1::Write32] Unknown address: %x", address))
	}
}
