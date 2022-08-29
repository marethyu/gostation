package main

import (
	"fmt"
)

/*
https://psx-spx.consoledev.net/cdromdrive/#cdrom-test-commands-version-switches-region-chipset-scex
*/
func (cdrom *CDROM) CommandTest() {
	if !cdrom.paramFIFO.Empty() {
		arg := cdrom.paramFIFO.Pop()

		switch arg {
		case 0x20: // 19h,20h --> INT3(yy,mm,dd,ver)
			// 94h,09h,19h,C0h  ;PSX (PU-7)               19 Sep 1994, version vC0 (a)
			cdrom.respFIFO.Reset(4)
			cdrom.respFIFO.Push(0x94)
			cdrom.respFIFO.Push(0x09)
			cdrom.respFIFO.Push(0x19)
			cdrom.respFIFO.Push(0xc0)

			cdrom.irqFlag = RESP_INT3
		default:
			panic(fmt.Sprintf("[CDROM::CommandTest] Unknown argument: %x", arg))
		}
	} else {
		panic("[CDROM::CommandTest] Missing one argument")
	}
}
