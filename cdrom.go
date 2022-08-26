package main

import (
	"fmt"
)

const (
	CDROM_OFFSET = 0x1f801800
	CDROM_SIZE   = 4
)

type CDROM struct {
	Core *GoStation

	/* 1F801800h - Index/Status Register */
	index int  /* 0-1 Index   Port 1F801801h-1F801803h index (0..3 = Index0..Index3)   (R/W) */
	busy  bool /* 7   BUSYSTS Command/parameter transmission busy  (1=Busy) */

	paramFIFO []uint8
	respFIFO  []uint8
	dataFIFO  []uint8
}

func NewCDROM(core *GoStation) *CDROM {
	return &CDROM{
		core,
		0,
		false,
		[]uint8{},
		[]uint8{},
		[]uint8{},
	}
}

func (cdrom *CDROM) Contains(address uint32) bool {
	return address >= CDROM_OFFSET && address < (CDROM_OFFSET+CDROM_SIZE)
}

func (cdrom *CDROM) Step() {
	cdrom.Core.Interrupts.Request(IRQ_CDROM)
}

func (cdrom *CDROM) Read8(address uint32) uint8 {
	switch address {
	case 0x1f801800:
		var status uint32 = 0

		status |= uint32(cdrom.index)
		ModifyBit(&status, 2, false) // hardcoded for now
		ModifyBit(&status, 3, len(cdrom.paramFIFO) == 0)
		ModifyBit(&status, 4, len(cdrom.paramFIFO) < 16)
		ModifyBit(&status, 5, len(cdrom.respFIFO) > 0)
		ModifyBit(&status, 6, len(cdrom.dataFIFO) > 0)
		ModifyBit(&status, 7, cdrom.busy)

		return uint8(status)
	case 0x1f801801:
		return 0
	case 0x1f801802:
		return 0
	case 0x1f801803:
		return 0
	default:
		panic(fmt.Sprintf("[CDROM::Read8] Invalid address: %x", address))
	}
}

func (cdrom *CDROM) Write8(address uint32, data uint8) {
	switch address {
	case 0x1f801800:
		cdrom.index = int(data & 0b11)
	case 0x1f801801:
	case 0x1f801802:
	case 0x1f801803:
	}
}
