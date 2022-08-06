package main

import (
	"fmt"
)

type CPU struct {
	Core *GoStationCore

	/* All registers */
	r  [32]uint32 /* R0 to R31 */
	pc uint32
	hi uint32
	lo uint32
}

func NewCPU(core *GoStationCore) *CPU {
	return &CPU{
		core,
		[32]uint32{0},
		0xbfc00000,
		0,
		0,
	}
}

func (cpu *CPU) Step() {
	opcode := cpu.Core.Bus.Read32(cpu.pc)
	cpu.pc += 4

	cpu.ExecutePrimaryOpcode(opcode)
}

/* Refer to this page https://psx-spx.consoledev.net/cpuspecifications/#cpu-opcode-encoding for all opcodes and its encodings */
func (cpu *CPU) ExecutePrimaryOpcode(opcode uint32) {
	op := GetValue(opcode, 26, 6)

	switch op {
	case 0x0d:
		cpu.OpORI(opcode)
	case 0x0f:
		cpu.OpLUI(opcode)
	case 0x2b:
		cpu.OpStoreWord(opcode)
	default:
		panic(fmt.Sprintf("[CPU::ExecutePrimaryOpcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) modifyReg(i int, v uint32) {
	cpu.r[i] = v
	cpu.r[0] = 0 // R0 is always zero
}

func (cpu *CPU) reg(i int) uint32 {
	return cpu.r[i]
}

/*
			ALL OPCODES IMPLEMENTED BELOW

Nice reference: https://ffhacktics.com/wiki/PSX_instruction_set
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//  6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |

// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
func (cpu *CPU) OpORI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5)) /* the low 16 bits of rt which is assumed to be zero will be filled with imm16 */
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rt) | imm16
	cpu.modifyReg(rs, val)
}

// 001111 | N/A  | rt   | <--immediate16bit--> | lui-imm
func (cpu *CPU) OpLUI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16) /* this value will be placed in the high 16 bits of a 32 bit value */
	rt := int(GetValue(opcode, 16, 5))

	val := imm16 << 16 /* the low 16 bits of a 32 bit value is set to zero */
	cpu.modifyReg(rt, val)
}

// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
func (cpu *CPU) OpStoreWord(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.reg(rt)
	cpu.Core.Bus.Write32(addr, val)
}
