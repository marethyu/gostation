package main

import (
	"fmt"
)

const (
	DMA_OFFSET = 0x1f801080
	DMA_SIZE   = 8 * 16
)

type DMA struct {
	Core *GoStation

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

func NewDMA(core *GoStation) *DMA {
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

func (dma *DMA) DoDMATransfer(port int) {
	// addresses to RAM must be masked
	// the size of RAM is 0x200000 and we want to make sure that addr can fit inside the ram so the mask is 0x200000-1 but
	// the first nybble is 'c' because we want aligned address
	var mask uint32 = 0x1ffffc

	switch dma.channel[port].syncMode {
	case SYNC_LINKED_LIST:
		addr := dma.channel[port].baseAddress & mask

		if port != DMA2_GPU {
			panic("[DMA::DoDMATransfer] I thought the linked list mode only works for ram to gpu?")
		}

		if !dma.channel[port].RAMToDevice {
			panic("[DMA::DoDMATransfer] linked list mode only works for ram to device")
		}

		for {
			// header of a packet
			// high 8 bits defines size
			// low 24 bits defines address to next packet or 0xffffff if last element
			header := dma.Core.Bus.Ram.Read32(addr)

			size := header >> 24

			for size > 0 {
				addr = (addr + 4) & mask
				command := dma.Core.Bus.Ram.Read32(addr)

				dma.Core.GPU.WriteGP0(command)

				size -= 1
			}

			// last element (can't use 0xffffff for some reason)
			if TestBit(header, 23) {
				break
			}

			addr = header & mask
		}
	default: /* block copy */
		addr := int(dma.channel[port].baseAddress)
		increment := 4
		if dma.channel[port].addressDecrement {
			increment = -4
		}

		size := dma.channel[port].TransferSize()
		if size == 0 {
			panic("[DMA::DoDMATransfer] size is zero during block copy?")
		}

		for size > 0 {
			cur_addr := uint32(addr) & mask

			if dma.channel[port].RAMToDevice {
				data := dma.Core.Bus.Ram.Read32(cur_addr)

				switch port {
				case DMA2_GPU:
					dma.Core.GPU.WriteGP0(data)
				default:
					panic(fmt.Sprintf("[DMA::DoDMATransfer] unsupported port (%d) during ram to device block copy", port))
				}
			} else {
				var data uint32
				switch port {
				case DMA6_OTC:
					if size == 1 {
						// last element of the ordering table
						data = 0xffffff
					} else {
						// pointer to previous entry
						data = uint32(addr-4) & 0x1fffff
					}
				default:
					panic(fmt.Sprintf("[DMA::DoDMATransfer] unsupported port (%d) during device to ram block copy", port))
				}

				dma.Core.Bus.Ram.Write32(cur_addr, data)
			}

			addr += increment
			size -= 1
		}
	}

	dma.channel[port].Done()
}

func (dma *DMA) Read32(address uint32) uint32 {
	byte := address & 0x000000ff
	nybble_lo := byte & 0xf
	nybble_hi := byte >> 4

	switch nybble_hi {
	case 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe:
		return dma.channel[nybble_hi-0x8].Read32(nybble_lo)
	case 0xf:
		switch nybble_lo {
		case 0x0:
			return dma.control
		case 0x4:
			var dicr uint32 = 0

			PackRange(&dicr, 0, uint32(dma.unknown), 6)
			ModifyBit(&dicr, 15, dma.forceIrq)
			PackRange(&dicr, 16, uint32(dma.dmaIE), 7)
			ModifyBit(&dicr, 23, dma.dmaIME)
			PackRange(&dicr, 24, uint32(dma.dmaIRQFlag), 7)
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
		port := int(nybble_hi - 0x8)

		dma.channel[port].Write32(nybble_lo, data)

		if dma.channel[port].Active() {
			dma.DoDMATransfer(port)
		}
	case 0xf:
		switch nybble_lo {
		case 0x0:
			dma.control = data
		case 0x4:
			dma.unknown = uint8(GetRange(data, 0, 6))
			dma.forceIrq = TestBit(data, 15)
			dma.dmaIE = uint8(GetRange(data, 16, 7))
			dma.dmaIME = TestBit(data, 23)
			dma.dmaIRQFlag = uint8(GetRange(data, 24, 7))
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
