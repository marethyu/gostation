package main

type Access interface {
	Contains(address uint32) bool
	Read8(address uint32) uint8
	Read16(address uint32) uint16
	Read32(address uint32) uint32
	Write8(address uint32, data uint8)
	Write16(address uint32, data uint16)
	Write32(address uint32, data uint32)
}
