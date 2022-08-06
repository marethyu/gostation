package main

type Memory struct {
	Data   []uint8
	Offset uint32
	Size   uint32
	Mask   uint32
}

func NewMemory(data []uint8, offset uint32, size uint32) *Memory {
	return &Memory{data, offset, size, size - 1}
}

func (mem *Memory) Contains(address uint32) bool {
	return address >= mem.Offset && address < (mem.Offset+mem.Size)
}

func (mem *Memory) Read8(address uint32) uint8 {
	return mem.Data[(address-mem.Offset)&mem.Mask]
}

func (mem *Memory) Read16(address uint32) uint16 {
	low := uint16(mem.Read8(address))
	high := uint16(mem.Read8(address + 1))

	return (high << 8) | low
}

func (mem *Memory) Read32(address uint32) uint32 {
	byte0 := uint32(mem.Read8(address))
	byte1 := uint32(mem.Read8(address + 1))
	byte2 := uint32(mem.Read8(address + 2))
	byte3 := uint32(mem.Read8(address + 3))

	return (byte3 << 24) | (byte2 << 16) | (byte1 << 8) | byte0
}

func (mem *Memory) Write8(address uint32, data uint8) {
	mem.Data[(address-mem.Offset)&mem.Mask] = data
}

func (mem *Memory) Write16(address uint32, data uint16) {
	low := uint8(data & 0xff)
	high := uint8(data >> 8)

	mem.Write8(address, low)
	mem.Write8(address+1, high)
}

func (mem *Memory) Write32(address uint32, data uint32) {
	byte0 := uint8(data & 0xff)
	byte1 := uint8((data & 0xff00) >> 8)
	byte2 := uint8((data & 0xff0000) >> 16)
	byte3 := uint8((data & 0xff000000) >> 24)

	mem.Write8(address, byte0)
	mem.Write8(address+1, byte1)
	mem.Write8(address+2, byte2)
	mem.Write8(address+3, byte3)
}
