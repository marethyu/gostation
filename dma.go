package main

import (
	"fmt"
)

const (
	DMA_OFFSET = 0x1f801080
	DMA_SIZE   = 8 * 16
)

type DMA struct {
	Core *GoStationCore

	channel [7]DMAChannel

	/* 1F8010F0h DPCR - DMA Control register */
	control uint32

	/* 1F8010F4h DICR - DMA Interrupt register */
	unknown    uint8 /* bits 0-5 */
	forceIrq   bool  /* bit 15: force set the irq master flag (bit 31) if true */
	dmaIE      uint8 /* bits 16-22: IRQ Enable for DMA0..DMA6 */
	dmaIME     bool  /* bit 23: IRQ Master Enable for DMA0..DMA6 */
	dmaIRQFlag uint8 /* bits 24-30: IRQ Flags for DMA0..DMA6 */
}

func NewDMA(core *GoStationCore) *DMA {
	return &DMA{
		core,
		[7]DMAChannel{},
		0x07654321,
		0,
		false,
		0,
		false,
		0,
	}
}

func (dma *DMA) Contains(address uint32) bool {
	return address >= DMA_OFFSET && address < (DMA_OFFSET+DMA_SIZE)
}

func (dma *DMA) Read32(address uint32) uint32 {
	byte := address & 0x000000ff
	nybble_lo := byte & 0xf
	nybble_hi := byte >> 4

	switch nybble_hi {
	case 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe:
		idx := nybble_hi - 0x8
		return dma.channel[idx].Read32(nybble_lo)
	case 0xf:
		switch nybble_lo {
		case 0x0:
			return dma.control
		case 0x4:
			var dicr uint32 = 0

			PackValue(&dicr, 0, uint32(dma.unknown), 6)
			ModifyBit(&dicr, 15, dma.forceIrq)
			PackValue(&dicr, 16, uint32(dma.dmaIE), 7)
			ModifyBit(&dicr, 23, dma.dmaIME)
			PackValue(&dicr, 24, uint32(dma.dmaIRQFlag), 7)
			ModifyBit(&dicr, 31, dma.IRQMasterFlag())

			return dicr
		default:
			panic(fmt.Sprintf("[DMA::Read32] (reading some register) Invalid address: %x", address))
		}
	default:
		panic(fmt.Sprintf("[DMA::Read32] Invalid address: %x", address))
	}
}

func (dma *DMA) Write32(address uint32, data uint32) {
	byte := address & 0x000000ff
	nybble_lo := byte & 0xf
	nybble_hi := byte >> 4

	switch nybble_hi {
	case 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe:
		idx := nybble_hi - 0x8
		dma.channel[idx].Write32(nybble_lo, data)
	case 0xf:
		switch nybble_lo {
		case 0x0:
			dma.control = data
		case 0x4:
			dma.unknown = uint8(GetValue(data, 0, 6))
			dma.forceIrq = TestBit(data, 15)
			dma.dmaIE = uint8(GetValue(data, 16, 7))
			dma.dmaIME = TestBit(data, 23)
			dma.dmaIRQFlag = uint8(GetValue(data, 24, 7))
		default:
			panic(fmt.Sprintf("[DMA::Write32] (writing to some register) attempt to write %x to invalid address: %x", data, address))
		}
	default:
		panic(fmt.Sprintf("[DMA::Write32] attempt to write %x to invalid address: %x", data, address))
	}
}

// bit 31 of DICR (read only)
func (dma *DMA) IRQMasterFlag() bool {
	// IF b15=1 OR (b23=1 AND (b16-22 AND b24-30)>0) THEN b31=1 ELSE b31=0
	return dma.forceIrq || (dma.dmaIME && (dma.dmaIE&dma.dmaIRQFlag) > 0)
}
