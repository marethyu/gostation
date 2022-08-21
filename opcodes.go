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

func (cpu *CPU) OpIllegal() {
	cpu.cop0.EnterException(EXC_RESERVED_INS, "illegal or reserved instruction")
}

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	cond := GetRange(opcode, 16, 5)
	rs := int(GetRange(opcode, 21, 5))

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
	imm26 := GetRange(opcode, 0, 26)

	cpu.next_pc = (cpu.pc & 0xf0000000) | (imm26 << 2)
	cpu.isBranch = true
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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	a := int32(imm16)
	b := int32(cpu.reg(rs))

	// enter exception if a+b results in overflow or underflow
	if (a > 0 && b > math.MaxInt32-a) ||
		(a < 0 && b < math.MinInt32-a) {
		cpu.cop0.EnterException(EXC_OVERFLOW, "signed overflow encountered in CPU::OpADDI")
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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	imm16 := GetRange(opcode, 0, 16)
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5)) /* the low 16 bits of rt which is assumed to be zero will be filled with imm16 */

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
	imm16 := GetRange(opcode, 0, 16)
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5)) /* the low 16 bits of rt which is assumed to be zero will be filled with imm16 */

	val := cpu.reg(rs) | imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// xori rt,rs,imm        rt = rs XOR (0000h..FFFFh)
func (cpu *CPU) OpXORI(opcode uint32) {
	imm16 := GetRange(opcode, 0, 16)
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rs) ^ imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001111 | N/A  | rt   | <--immediate16bit--> | lui-imm
// lui  rt,imm            rt = (0000h..FFFFh) SHL 16
func (cpu *CPU) OpLUI(opcode uint32) {
	imm16 := GetRange(opcode, 0, 16) /* this value will be placed in the high 16 bits of a 32 bit value */
	rt := int(GetRange(opcode, 16, 5))

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
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := SignExtendedByte(cpu.Core.Bus.Read8(addr))

	cpu.loadDelaySlotInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lh  rt,imm(rs)    rt=[imm+rs]  ;halfword sign-extended
func (cpu *CPU) OpLoadHWord(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	// addresses must be 16 bit aligned
	if addr%2 != 0 {
		cpu.cop0.EnterException(EXC_ADDR_ERROR_LOAD, "unaligned address during lh")
	} else {
		val := SignExtendedHWord(cpu.Core.Bus.Read16(addr))
		cpu.loadDelaySlotInit(rt, val)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lwl   rt,imm(rs)     load left  bits of rt from memory (usually imm+3)
// IMPORTANT NOTE: "left" refers to the *most* significant part not least significant part
// see also CPU::OpLoadWordRight
func (cpu *CPU) OpLoadWordLeft(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	var val uint32
	if cpu.pending_load && cpu.pending_r == rt {
		// Bypass delay slot if needed
		val = cpu.pending_val
	} else {
		val = cpu.reg(rt)
	}

	mask := ^uint32(0b11) // bitmask to strip of lower two bits of the address to get aligned address
	aligned_word := cpu.Core.Bus.Read32(addr & mask)

	// yea bro it's unintuitive compared to lwr but lwl instructions are frequently paired with offset (imm) of 3 for some obscure reason
	switch addr % 4 {
	/* in this case the lwr instruction already put the right 3 bytes of word located at aligned_address+1 on right 3 bytes of val, we need to fill the first byte.
	for example, we want to fill $r11 with ┃ 1 ┆ 2 ┆ 3 ┆ 4 ┃ (┃ least significant --> most significant ┃).

		memory:
		     ├ A             ┤
	     ┃ 0 ┆ 1 ┆ 2 ┆ 3 ┃ 4 ┆ 5 ┆ 6 ┆ 7 ┃

		 $r7=A
		 after lwr $r11,0($r7): $r11=┃ 1 ┆ 2 ┆ 3 ┆ ? ┃

		 let B=A+3
		 memory:
		                 ├ B             ┤
	     ┃ 0 ┆ 1 ┆ 2 ┆ 3 ┃ 4 ┆ 5 ┆ 6 ┆ 7 ┃

		 lwl will fill the ? part in $r11 with the first byte (least significant part) of word located at B which is 4
	*/
	case 0:
		val = (val & 0x00ffffff) | (aligned_word << 24) // the least significant byte of aligned word (bytes 4-7 in example) is placed in the most significant byte of val
	/* in this case the lwr instruction already put the right 2 bytes of word located at aligned_address+2 on right 2 bytes of val, we need to fill the first two bytes.
	for example, we want to fill $r11 with ┃ 2 ┆ 3 ┆ 4 ┆ 5 ┃.

		memory:
		         ├ A             ┤
	     ┃ 0 ┆ 1 ┆ 2 ┆ 3 ┃ 4 ┆ 5 ┆ 6 ┆ 7 ┃

		 $r7=A
		 after lwr $r11,0($r7): $r11=┃ 2 ┆ 3 ┆ ? ┆ ? ┃

		 let B=A+3
		 memory:
		                 ├ B             ┤
	     ┃ 0 ┆ 1 ┆ 2 ┆ 3 ┃ 4 ┆ 5 ┆ 6 ┆ 7 ┃

		 lwl will fill the ┆ ? ┆ ? ┆ part in $r11 with the first two bytes (least significant part) of word located at B which is ┆ 4 ┆ 5 ┆
	*/
	case 1:
		val = (val & 0x0000ffff) | (aligned_word << 16)
	/* ya should get the idea how it works at dis point */
	case 2:
		val = (val & 0x000000ff) | (aligned_word << 8)
	case 3:
		val = (val & 0x00000000) | (aligned_word << 0)
	}

	cpu.loadDelaySlotInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lw  rt,imm(rs)    rt=[imm+rs]  ;word
func (cpu *CPU) OpLoadWord(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	// addresses must be 32 bit aligned
	if addr%4 != 0 {
		cpu.cop0.EnterException(EXC_ADDR_ERROR_LOAD, "unaligned address during lw")
	} else {
		val := cpu.Core.Bus.Read32(addr)
		cpu.loadDelaySlotInit(rt, val)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lbu rt,imm(rs)    rt=[imm+rs]  ;byte zero-extended
func (cpu *CPU) OpLoadByteU(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := uint32(cpu.Core.Bus.Read8(addr)) // zero extended

	cpu.loadDelaySlotInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lhu rt,imm(rs)    rt=[imm+rs]  ;halfword zero-extended
func (cpu *CPU) OpLoadHWordU(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	// addresses must be 16 bit aligned
	if addr%2 != 0 {
		cpu.cop0.EnterException(EXC_ADDR_ERROR_LOAD, "unaligned address during lhu")
	} else {
		val := cpu.Core.Bus.Read16(addr)
		cpu.loadDelaySlotInit(rt, uint32(val))
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lwr   rt,imm(rs)     load right bits of rt from memory (usually imm+0)
func (cpu *CPU) OpLoadWordRight(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	var val uint32
	if cpu.pending_load && cpu.pending_r == rt {
		// Bypass delay slot if needed
		val = cpu.pending_val
	} else {
		val = cpu.reg(rt)
	}

	mask := ^uint32(0b11) // bitmask to strip of lower two bits of the address to get aligned address
	aligned_word := cpu.Core.Bus.Read32(addr & mask)

	switch addr % 4 {
	case 0:
		val = (val & 0x00000000) | (aligned_word >> 0)
	case 1:
		val = (val & 0xff000000) | (aligned_word >> 8)
	case 2:
		val = (val & 0xffff0000) | (aligned_word >> 16)
	case 3:
		val = (val & 0xffffff00) | (aligned_word >> 24)
	}

	cpu.loadDelaySlotInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sb  rt,imm(rs)    [imm+rs]=(rt AND FFh)   ;store 8bit
func (cpu *CPU) OpStoreByte(opcode uint32) {
	if cpu.cop0.CacheIsolated() {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	if cpu.cop0.CacheIsolated() {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	// addresses must be 16 bit aligned
	if addr%2 != 0 {
		cpu.cop0.EnterException(EXC_ADDR_ERROR_STORE, "unaligned address during sh")
	} else {
		val := uint16(cpu.reg(rt))
		cpu.Core.Bus.Write16(addr, val)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// swl   rt,imm(rs)     store left  bits of rt to memory (usually imm+3)
// see also CPU::OpStoreWordRight
func (cpu *CPU) OpStoreWordLeft(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.reg(rt)

	mask := ^uint32(0b11) // bitmask to strip of lower two bits of the address to get aligned address
	aligned_addr := addr & mask
	aligned_word := cpu.Core.Bus.Read32(aligned_addr)

	switch addr % 4 {
	case 0:
		val = (aligned_word & 0xffffff00) | (val >> 24) // the most significant byte of val is put in the least significant byte of aligned word
	case 1:
		val = (aligned_word & 0xffff0000) | (val >> 16) // yea u get the idea
	case 2:
		val = (aligned_word & 0xff000000) | (val >> 8)
	case 3:
		val = (aligned_word & 0x00000000) | (val >> 0)
	}

	cpu.Core.Bus.Write32(aligned_addr, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sw  rt,imm(rs)    [imm+rs]=rt             ;store 32bit
func (cpu *CPU) OpStoreWord(opcode uint32) {
	if cpu.cop0.CacheIsolated() {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16

	// addresses must be 32 bit aligned
	if addr%4 != 0 {
		cpu.cop0.EnterException(EXC_ADDR_ERROR_STORE, "unaligned address during sw")
	} else {
		val := cpu.reg(rt)
		cpu.Core.Bus.Write32(addr, val)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// swr   rt,imm(rs)     store right bits of rt to memory (usually imm+0)
func (cpu *CPU) OpStoreWordRight(opcode uint32) {
	imm16 := SignExtendedWord(GetRange(opcode, 0, 16))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.reg(rt)

	mask := ^uint32(0b11) // bitmask to strip of lower two bits of the address to get aligned address
	aligned_addr := addr & mask
	aligned_word := cpu.Core.Bus.Read32(aligned_addr)

	switch addr % 4 {
	case 0:
		val = (aligned_word & 0x00000000) | (val << 0)
	case 1:
		val = (aligned_word & 0x000000ff) | (val << 8) // the most significant 3 bytes of aligned word gets filled with the least significant 3 bytes of val
	case 2:
		val = (aligned_word & 0x0000ffff) | (val << 16)
	case 3:
		val = (aligned_word & 0x00ffffff) | (val << 24)
	}

	cpu.Core.Bus.Write32(aligned_addr, val)
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
	imm5 := GetRange(opcode, 6, 5)
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))

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
	imm5 := GetRange(opcode, 6, 5)
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))

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
	imm5 := GetRange(opcode, 6, 5)
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))

	val := int32(cpu.reg(rt)) >> imm5
	cpu.modifyReg(rd, uint32(val))
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 0001xx | shift-reg
// sllv rd,rt,rs          rd = rt SHL (rs AND 1Fh)
func (cpu *CPU) OpSLLV(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rt) << (cpu.reg(rs) & 0x1f)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 0001xx | shift-reg
// srlv rd,rt,rs          rd = rt SHR (rs AND 1Fh)
func (cpu *CPU) OpSRLV(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rt) >> (cpu.reg(rs) & 0x1f)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 0001xx | shift-reg
// srav rd,rt,rs          rd = rt SAR (rs AND 1Fh)
func (cpu *CPU) OpSRAV(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := int32(cpu.reg(rt)) >> (cpu.reg(rs) & 0x1f)
	cpu.modifyReg(rd, uint32(val))
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 001000 | jr
// jr     rs          pc=rs
func (cpu *CPU) OpJR(opcode uint32) {
	rs := int(GetRange(opcode, 21, 5))

	cpu.next_pc = cpu.reg(rs)
	cpu.isBranch = true
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | rd   | N/A  | 001001 | jalr
// jalr (rd,)rs(,rd)  pc=rs, rd=$+8
func (cpu *CPU) OpJALR(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rs := int(GetRange(opcode, 21, 5))

	cpu.modifyReg(rd, cpu.next_pc) // store the return address in rd
	cpu.next_pc = cpu.reg(rs)
	cpu.isBranch = true
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | <-----comment20bit------> | 00110x | sys/brk
// syscall  imm20        generates a system call exception
func (cpu *CPU) OpSYS(opcode uint32) {
	// comment := GetRange(opcode, 6, 20)

	cpu.cop0.EnterException(EXC_SYSCALL, fmt.Sprintf("system call: %s", cpu.identifySystemCall()))
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | <-----comment20bit------> | 00110x | sys/brk
// break    imm20        generates a breakpoint exception
func (cpu *CPU) OpBRK(opcode uint32) {
	// comment := GetRange(opcode, 6, 20)

	cpu.cop0.EnterException(EXC_BREAK, "breakpoint")
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | N/A  | rd   | N/A  | 0100x0 | mfhi/mflo
// mfhi   rd              rd=hi  ;move from hi
func (cpu *CPU) OpMFHI(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))

	cpu.modifyReg(rd, cpu.hi)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 0100x1 | mthi/mtlo
// mthi   rs              hi=rs  ;move to hi
func (cpu *CPU) OpMTHI(opcode uint32) {
	rs := int(GetRange(opcode, 21, 5))

	cpu.hi = cpu.reg(rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | N/A  | rd   | N/A  | 0100x0 | mfhi/mflo
// mflo   rd              rd=lo  ;move from lo
func (cpu *CPU) OpMFLO(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))

	cpu.modifyReg(rd, cpu.lo)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 0100x1 | mthi/mtlo
// mtlo   rs              lo=rs  ;move to lo
func (cpu *CPU) OpMTLO(opcode uint32) {
	rs := int(GetRange(opcode, 21, 5))

	cpu.lo = cpu.reg(rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// mult   rs,rt           hi:lo = rs*rt (signed)
func (cpu *CPU) OpMULT(opcode uint32) {
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	a := int64(int32(cpu.reg(rs)))
	b := int64(int32(cpu.reg(rt)))
	val := uint64(a * b)

	cpu.hi = uint32(val >> 32)
	cpu.lo = uint32(val & 0xffffffff)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// multu  rs,rt           hi:lo = rs*rt (unsigned)
func (cpu *CPU) OpMULTU(opcode uint32) {
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	a := uint64(cpu.reg(rs))
	b := uint64(cpu.reg(rt))
	val := a * b

	cpu.hi = uint32(val >> 32)
	cpu.lo = uint32(val & 0xffffffff)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// div    rs,rt           lo = rs/rt, hi=rs mod rt (signed)
// TODO timing
func (cpu *CPU) OpDIV(opcode uint32) {
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	a := int32(cpu.reg(rs))
	b := int32(cpu.reg(rt))

	// enter exception if a+b results in overflow or underflow
	if (a > 0 && b > math.MaxInt32-a) ||
		(a < 0 && b < math.MinInt32-a) {
		cpu.cop0.EnterException(EXC_OVERFLOW, "signed overflow encountered in CPU::OpADD")
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
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rs) + cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// sub   rd,rs,rt         rd=rs-rt (with overflow trap)
func (cpu *CPU) OpSUB(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	a := int32(cpu.reg(rs))
	b := int32(cpu.reg(rt))

	if (b < 0 && a > math.MaxInt32+b) ||
		(b > 0 && a < math.MinInt32+b) {
		cpu.cop0.EnterException(EXC_OVERFLOW, "signed overflow encountered in CPU::OpSUB")
	}

	val := uint32(a - b)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// subu  rd,rs,rt         rd=rs-rt
func (cpu *CPU) OpSUBU(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rs) | cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// xor  rd,rs,rt         rd = rs XOR rt
func (cpu *CPU) OpXOR(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rs) ^ cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// nor  rd,rs,rt         rd = FFFFFFFFh XOR (rs OR rt)
func (cpu *CPU) OpNOR(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	val := cpu.reg(rs) | cpu.reg(rt)
	val = ^val
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// setlt slt   rd,rs,rt  if rs<rt then rd=1 else rd=0 (signed)
func (cpu *CPU) OpSLT(opcode uint32) {
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

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
	rd := int(GetRange(opcode, 11, 5))
	rt := int(GetRange(opcode, 16, 5))
	rs := int(GetRange(opcode, 21, 5))

	test := cpu.reg(rs) < cpu.reg(rt)
	if test {
		cpu.modifyReg(rd, 1)
	} else {
		cpu.modifyReg(rd, 0)
	}
}

/*
Coprocessor opcodes are implemented here
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0000| rt   | rd   | N/A  | 000000 | MFCn rt,rd_dat  ;rt = dat
// mfc# rt,rd       ;rt = cop#datRd ;data regs
func (cpu *CPU) OpMFC0(opcode uint32) {
	rd := GetRange(opcode, 11, 5)
	rt := int(GetRange(opcode, 16, 5))

	val := cpu.cop0.GetRegister(rd)
	cpu.loadDelaySlotInit(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0100| rt   | rd   | N/A  | 000000 | MTCn rt,rd_dat  ;dat = rt
// mtc# rt,rd       ;cop#datRd = rt ;data regs
func (cpu *CPU) OpMTC0(opcode uint32) {
	rd := GetRange(opcode, 11, 5)
	rt := int(GetRange(opcode, 16, 5))

	val := cpu.reg(rt)
	cpu.cop0.ModifyRegister(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 010000 |1|0000| N/A  | N/A  | N/A  | 010000 | COP0 10h  ;=RFE
// rfe
func (cpu *CPU) OpRFE(opcode uint32) {
	if GetRange(opcode, 0, 6) != 0b010000 {
		panic(fmt.Sprintf("[CPU::OpRFE] Unknown opcode: %x", opcode))
	}

	cpu.cop0.LeaveException()
}

func (cpu *CPU) OpLWC0(opcode uint32) {
	cpu.cop0.EnterException(EXC_COP_UNUSABLE, "lwc0 is not supported")
}

func (cpu *CPU) OpLWC1(opcode uint32) {
	cpu.cop0.EnterException(EXC_COP_UNUSABLE, "lwc1 is not supported")
}

func (cpu *CPU) OpLWC2(opcode uint32) {
	panic("[CPU::OpLWC2] GTE is not implemented yet!")
}

func (cpu *CPU) OpLWC3(opcode uint32) {
	cpu.cop0.EnterException(EXC_COP_UNUSABLE, "lwc3 is not supported")
}

func (cpu *CPU) OpSWC0(opcode uint32) {
	cpu.cop0.EnterException(EXC_COP_UNUSABLE, "swc0 is not supported")
}

func (cpu *CPU) OpSWC1(opcode uint32) {
	cpu.cop0.EnterException(EXC_COP_UNUSABLE, "swc1 is not supported")
}

func (cpu *CPU) OpSWC2(opcode uint32) {
	panic("[CPU::OpSWC2] GTE is not implemented yet!")
}

func (cpu *CPU) OpSWC3(opcode uint32) {
	cpu.cop0.EnterException(EXC_COP_UNUSABLE, "swc3 is not supported")
}
