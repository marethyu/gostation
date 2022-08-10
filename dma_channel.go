package main

import (
	"fmt"
)

/*
1F80108xh DMA0 channel 0  MDECin  (RAM to MDEC)
1F80109xh DMA1 channel 1  MDECout (MDEC to RAM)
1F8010Axh DMA2 channel 2  GPU (lists + image data)
1F8010Bxh DMA3 channel 3  CDROM   (CDROM to RAM)
1F8010Cxh DMA4 channel 4  SPU
1F8010Dxh DMA5 channel 5  PIO (Expansion Port)
1F8010Exh DMA6 channel 6  OTC (reverse clear OT) (GPU related)
*/
const (
	DMA0_MDECin = iota
	DMA1_MDECout
	DMA2_GPU
	DMA3_CDROM
	DMA4_SPU
	DMA5_PIO
	DMA6_OTC
)

const (
	SYNC_ALL_AT_ONCE = iota
	SYNC_REQUEST
	SYNC_LINKED_LIST
)

type DMAChannel struct {
	/* 1F801080h+N*10h - D#_MADR - DMA base address (Channel 0..6) (R/W) */
	baseAddress uint32 /* only bits 0-23 are used */

	/* 1F801084h+N*10h - D#_BCR - DMA Block Control (Channel 0..6) (R/W)

	For SyncMode=0 (ie. for OTC and CDROM):
		0-15  BC    Number of words (0001h..FFFFh) (or 0=10000h words)
		16-31 0     Not used (usually 0 for OTC, or 1 ("one block") for CDROM)
	For SyncMode=1 (ie. for MDEC, SPU, and GPU-vram-data):
	    0-15  BS    Blocksize (words) ;for GPU/SPU max 10h, for MDEC max 20h
	    16-31 BA    Amount of blocks  ;ie. total length = BS*BA words
	*/
	blockSize   uint16 /* or number of words */
	blockAmount uint16

	/* 1F801088h+N*10h - D#_CHCR - DMA Channel Control (Channel 0..6) (R/W) */
	RAMToDevice      bool  /* bit 0: 0=device to RAM, 1=from RAM to device */
	addressDecrement bool  /* bit 1: 1=decrement address (-4 bytes); 0=increment address (+4 bytes) */
	choppingEnable   bool  /* bit 8: if 1 run CPU during DMA gaps */
	syncMode         uint8 /* bits 9-10: transfer syncronization mode */
	choppingDMAWind  uint8 /* bits 16-18: Chopping DMA Window Size (1 SHL N words) */
	choppingCPUWind  uint8 /* bits 20-22: Chopping CPU Window Size (1 SHL N clks) */
	start            bool  /* bit 24: start/busy 0=Stopped/Completed, 1=Start/Enable/Busy */
	trigger          bool  /* bit 28: start/trigger 0=Normal, 1=Manual Start; use for SyncMode=0 */
	unknown          uint8 /* bits 29-30 */
}

func NewDMAChannel() *DMAChannel {
	return &DMAChannel{
		0,
		0,
		0,
		false,
		false,
		false,
		0,
		0,
		0,
		false,
		false,
		0,
	}
}

func (channel *DMAChannel) Read32(offset uint32) uint32 {
	switch offset {
	case 0x0:
		return channel.baseAddress
	case 0x4:
		return (uint32(channel.blockAmount) << 16) | uint32(channel.blockSize)
	case 0x8:
		var control uint32 = 0

		ModifyBit(&control, 0, channel.RAMToDevice)
		ModifyBit(&control, 1, channel.addressDecrement)
		ModifyBit(&control, 8, channel.choppingEnable)
		control |= uint32(channel.syncMode&0b11) << 9
		control |= uint32(channel.choppingDMAWind&0b111) << 16
		control |= uint32(channel.choppingCPUWind&0b111) << 20
		ModifyBit(&control, 24, channel.start)
		ModifyBit(&control, 28, channel.trigger)
		control |= uint32(channel.unknown&0b11) << 29

		return control
	default:
		panic(fmt.Sprintf("[DMAChannel::Read32] Invalid offset: %x", offset))
	}
}

func (channel *DMAChannel) Write32(offset uint32, data uint32) {
	switch offset {
	case 0x0:
		channel.baseAddress = data & 0xffffff
	case 0x4:
		channel.blockSize = uint16(data & 0xffff)
		channel.blockAmount = uint16(data >> 16)
	case 0x8:
		channel.RAMToDevice = TestBit(data, 0)
		channel.addressDecrement = TestBit(data, 1)
		channel.choppingEnable = TestBit(data, 8)
		channel.syncMode = uint8(GetValue(data, 9, 2))
		channel.choppingDMAWind = uint8(GetValue(data, 16, 3))
		channel.choppingCPUWind = uint8(GetValue(data, 20, 3))
		channel.start = TestBit(data, 24)
		channel.trigger = TestBit(data, 28)
		channel.unknown = uint8(GetValue(data, 29, 2))
	default:
		panic(fmt.Sprintf("[DMAChannel::Write32] Attempt to write %x to invalid offset: %x", data, offset))
	}
}
