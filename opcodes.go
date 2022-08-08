package main

import (
	"fmt"
	"math"
)

/*
Nice references:
https://psx-spx.consoledev.net/cpuspecifications/
https://ffhacktics.com/wiki/PSX_instruction_set
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000001 | rs   | 00000| <--immediate16bit--> | bltz
// 000001 | rs   | 00001| <--immediate16bit--> | bgez
// 000001 | rs   | 10000| <--immediate16bit--> | bltzal
// 000001 | rs   | 10001| <--immediate16bit--> | bgezal
// bltz   rs,dest     if rs<0   then pc=$+4+(-8000h..+7FFFh)*4
// bgez   rs,dest     if rs>=0  then pc=$+4+(-8000h..+7FFFh)*4
// bltzal rs,dest     if rs<0   then pc=$+4+(..)*4, ra=$+8
// bgezal rs,dest     if rs>=0  then pc=$+4+(..)*4, ra=$+8
func (cpu *CPU) OpBcondZ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	cond := GetValue(opcode, 16, 5)
	rs := int(GetValue(opcode, 21, 5))

	val := int32(cpu.reg(rs))

	var test bool
	var link bool
	switch cond {
	case 0b00000:
		test = val < 0
		link = false
	case 0b00001:
		test = val >= 0
		link = false
	case 0b10000:
		test = val < 0
		link = true
	case 0b10001:
		test = val >= 0
		link = true
	default:
		panic(fmt.Sprintf("[CPU::OpBcondZ] Unknown condition: %x", cond))
	}

	if link {
		cpu.modifyReg(31, cpu.next_pc) // store the return address in ra
	}

	if test {
		cpu.branch(imm16)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00001x | <---------immediate26bit---------> | j/jal
// j      dest        pc=(pc and F0000000h)+(imm26bit*4)
func (cpu *CPU) OpJump(opcode uint32) {
	imm26 := GetValue(opcode, 0, 26)

	cpu.next_pc = (cpu.pc & 0xf0000000) | (imm26 << 2)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00001x | <---------immediate26bit---------> | j/jal
// jal    dest        pc=(pc and F0000000h)+(imm26bit*4),ra=$+8
func (cpu *CPU) OpJAL(opcode uint32) {
	cpu.modifyReg(31, cpu.next_pc) // store the return address in ra
	cpu.OpJump(opcode)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00010x | rs   | rt   | <--immediate16bit--> | beq/bne
// beq    rs,rt,dest  if rs=rt  then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) OpBEQ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	test := cpu.reg(rs) == cpu.reg(rt)
	if test {
		cpu.branch(imm16)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00010x | rs   | rt   | <--immediate16bit--> | beq/bne
// bne    rs,rt,dest  if rs<>rt then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) OpBNE(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	test := cpu.reg(rs) != cpu.reg(rt)
	if test {
		cpu.branch(imm16)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00011x | rs   | N/A  | <--immediate16bit--> | blez/bgtz
// blez   rs,dest     if rs<=0  then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) OpBLEZ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rs := int(GetValue(opcode, 21, 5))

	val := int32(cpu.reg(rs))
	test := val <= 0
	if test {
		cpu.branch(imm16)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00011x | rs   | N/A  | <--immediate16bit--> | blez/bgtz
// bgtz   rs,dest     if rs>0   then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) OpBGTZ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rs := int(GetValue(opcode, 21, 5))

	val := int32(cpu.reg(rs))
	test := val > 0
	if test {
		cpu.branch(imm16)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// addi  rt,rs,imm        rt=rs+(-8000h..+7FFFh) (with ov.trap)
func (cpu *CPU) OpADDI(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	a := int32(imm16)
	b := int32(cpu.reg(rs))

	if a > 0 && b > math.MaxInt32-a {
		panic("[CPU::OpADDI] Signed integer overflow!!!")
	}

	val := uint32(a + b)
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// addiu rt,rs,imm        rt=rs+(-8000h..+7FFFh)
func (cpu *CPU) OpADDIU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) + imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// setlt slti  rt,rs,imm if rs<(-8000h..+7FFFh)  then rt=1 else rt=0 (signed)
func (cpu *CPU) OpSLTI(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := int32(cpu.reg(rs))
	test := val < int32(imm16)
	if test {
		cpu.modifyReg(rt, 1)
	} else {
		cpu.modifyReg(rt, 0)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// setb  sltiu rt,rs,imm if rs<(FFFF8000h..7FFFh) then rt=1 else rt=0(unsigned)
func (cpu *CPU) OpSLTIU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs)
	test := val < imm16
	if test {
		cpu.modifyReg(rt, 1)
	} else {
		cpu.modifyReg(rt, 0)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// andi rt,rs,imm        rt = rs AND (0000h..FFFFh)
func (cpu *CPU) OpANDI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5)) /* the low 16 bits of rt which is assumed to be zero will be filled with imm16 */

	val := cpu.reg(rs) & imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// ori  rt,rs,imm        rt = rs OR  (0000h..FFFFh)
func (cpu *CPU) OpORI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5)) /* the low 16 bits of rt which is assumed to be zero will be filled with imm16 */

	val := cpu.reg(rs) | imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001111 | N/A  | rt   | <--immediate16bit--> | lui-imm
// lui  rt,imm            rt = (0000h..FFFFh) SHL 16
func (cpu *CPU) OpLUI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16) /* this value will be placed in the high 16 bits of a 32 bit value */
	rt := int(GetValue(opcode, 16, 5))

	val := imm16 << 16 /* the low 16 bits of a 32 bit value is set to zero */
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lb  rt,imm(rs)    rt=[imm+rs]  ;byte sign-extended
func (cpu *CPU) OpLoadByte(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := SignExtendedByte(cpu.Core.Bus.Read8(addr))

	cpu.loadDelayInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lw  rt,imm(rs)    rt=[imm+rs]  ;word
func (cpu *CPU) OpLoadWord(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.Core.Bus.Read32(addr)

	cpu.loadDelayInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lbu rt,imm(rs)    rt=[imm+rs]  ;byte zero-extended
func (cpu *CPU) OpLoadByteU(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := uint32(cpu.Core.Bus.Read8(addr)) // zero extended

	cpu.loadDelayInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sb  rt,imm(rs)    [imm+rs]=(rt AND FFh)   ;store 8bit
func (cpu *CPU) OpStoreByte(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := uint8(cpu.reg(rt))
	cpu.Core.Bus.Write8(addr, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sh  rt,imm(rs)    [imm+rs]=(rt AND FFFFh) ;store 16bit
func (cpu *CPU) OpStoreHWord(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := uint16(cpu.reg(rt))
	cpu.Core.Bus.Write16(addr, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sw  rt,imm(rs)    [imm+rs]=rt             ;store 32bit
func (cpu *CPU) OpStoreWord(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.reg(rt)
	cpu.Core.Bus.Write32(addr, val)
}

/*
Secondary opcodes are implemented here
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | rt   | rd   | imm5 | 0000xx | shift-imm
// sll  rd,rt,imm         rd = rt SHL (00h..1Fh)
func (cpu *CPU) OpSLL(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	val := cpu.reg(rt) << imm5
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | rt   | rd   | imm5 | 0000xx | shift-imm
// srl  rd,rt,imm         rd = rt SHR (00h..1Fh)
func (cpu *CPU) OpSRL(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	val := cpu.reg(rt) >> imm5
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | rt   | rd   | imm5 | 0000xx | shift-imm
// sra  rd,rt,imm         rd = rt SAR (00h..1Fh)
func (cpu *CPU) OpSRA(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	val := int32(cpu.reg(rt)) >> imm5
	cpu.modifyReg(rd, uint32(val))
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 001000 | jr
// jr     rs          pc=rs
func (cpu *CPU) OpJR(opcode uint32) {
	rs := int(GetValue(opcode, 21, 5))

	cpu.next_pc = cpu.reg(rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | rd   | N/A  | 001001 | jalr
// jalr (rd,)rs(,rd)  pc=rs, rd=$+8
func (cpu *CPU) OpJALR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rs := int(GetValue(opcode, 21, 5))

	cpu.modifyReg(rd, cpu.next_pc) // store the return address in rd
	cpu.next_pc = cpu.reg(rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | N/A  | rd   | N/A  | 0100x0 | mfhi/mflo
// mfhi   rd              rd=hi  ;move from hi
func (cpu *CPU) OpMFHI(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))

	cpu.modifyReg(rd, cpu.hi)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | N/A  | rd   | N/A  | 0100x0 | mfhi/mflo
// mflo   rd              rd=lo  ;move from lo
func (cpu *CPU) OpMFLO(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))

	cpu.modifyReg(rd, cpu.lo)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// div    rs,rt           lo = rs/rt, hi=rs mod rt (signed)
// TODO timing
func (cpu *CPU) OpDIV(opcode uint32) {
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	a := int32(cpu.reg(rs))
	b := int32(cpu.reg(rt))

	if b == 0 {
		cpu.hi = uint32(a)

		if a >= 0 {
			cpu.lo = 0xffffffff
		} else {
			cpu.lo = 1
		}
	} else if uint32(a) == 0x80000000 && b == -1 {
		cpu.hi = 0
		cpu.lo = 0x80000000
	} else {
		cpu.hi = uint32(a % b)
		cpu.lo = uint32(a / b)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// divu   rs,rt           lo = rs/rt, hi=rs mod rt (unsigned)
// TODO timing
func (cpu *CPU) OpDIVU(opcode uint32) {
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	a := cpu.reg(rs)
	b := cpu.reg(rt)

	if b == 0 {
		cpu.hi = a
		cpu.lo = 0xffffffff
	} else {
		cpu.hi = a % b
		cpu.lo = a / b
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// add   rd,rs,rt         rd=rs+rt (with overflow trap)
func (cpu *CPU) OpADD(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	a := int32(cpu.reg(rs))
	b := int32(cpu.reg(rt))

	if a > 0 && b > math.MaxInt32-a {
		panic("[CPU::OpADD] Signed integer overflow!!!")
	}

	val := uint32(a + b)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// addu  rd,rs,rt         rd=rs+rt
func (cpu *CPU) OpADDU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) + cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// subu  rd,rs,rt         rd=rs-rt
func (cpu *CPU) OpSUBU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) - cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// and  rd,rs,rt         rd = rs AND rt
func (cpu *CPU) OpAND(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) & cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// or   rd,rs,rt         rd = rs OR  rt
func (cpu *CPU) OpOR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) | cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// setlt slt   rd,rs,rt  if rs<rt then rd=1 else rd=0 (signed)
func (cpu *CPU) OpSLT(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	test := int32(cpu.reg(rs)) < int32(cpu.reg(rt))
	if test {
		cpu.modifyReg(rd, 1)
	} else {
		cpu.modifyReg(rd, 0)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// setb  sltu  rd,rs,rt  if rs<rt then rd=1 else rd=0 (unsigned)
func (cpu *CPU) OpSLTU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	test := cpu.reg(rs) < cpu.reg(rt)
	if test {
		cpu.modifyReg(rd, 1)
	} else {
		cpu.modifyReg(rd, 0)
	}
}

/*
COP0 opcodes are implemented here
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0000| rt   | rd   | N/A  | 000000 | MFCn rt,rd_dat  ;rt = dat
// mfc# rt,rd       ;rt = cop#datRd ;data regs
func (cpu *CPU) OpMFC0(opcode uint32) {
	rd := GetValue(opcode, 11, 5)
	rt := int(GetValue(opcode, 16, 5))

	var val uint32
	switch rd {
	case 12:
		val = cpu.sr
	case 13:
		panic("[CPU::OpMFC0] The cause register is not implemented yet!")
	default:
		panic(fmt.Sprintf("[CPU::OpMFC0] Unsupported COP0 register: %x", rd))
	}

	cpu.loadDelayInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0100| rt   | rd   | N/A  | 000000 | MTCn rt,rd_dat  ;dat = rt
// mtc# rt,rd       ;cop#datRd = rt ;data regs
func (cpu *CPU) OpMTC0(opcode uint32) {
	rd := GetValue(opcode, 11, 5)
	rt := int(GetValue(opcode, 16, 5))

	val := cpu.reg(rt)
	switch rd {
	case 3, 5, 6, 7, 9, 11:
		if val != 0 {
			panic("[CPU::OpMTC0] The breakpoint registers are implemented yet!")
		}
	case 12:
		cpu.sr = val
	case 13:
		if val != 0 {
			panic("[CPU::OpMTC0] The cause register is not implemented yet!")
		}
	default:
		panic(fmt.Sprintf("[CPU::OpMTC0] Unsupported COP0 register: %x", rd))
	}
}
