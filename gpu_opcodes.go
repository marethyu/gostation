package main

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e1h-draw-mode-setting-aka-texpage

0-3   Texture page X Base   (N*64) (ie. in 64-halfword steps)    ;GPUSTAT.0-3
4     Texture page Y Base   (N*256) (ie. 0 or 256)               ;GPUSTAT.4
5-6   Semi Transparency     (0=B/2+F/2, 1=B+F, 2=B-F, 3=B+F/4)   ;GPUSTAT.5-6
7-8   Texture page colors   (0=4bit, 1=8bit, 2=15bit, 3=Reserved);GPUSTAT.7-8
9     Dither 24bit to 15bit (0=Off/strip LSBs, 1=Dither Enabled) ;GPUSTAT.9
10    Drawing to display area (0=Prohibited, 1=Allowed)          ;GPUSTAT.10
11    Texture Disable (0=Normal, 1=Disable if GP1(09h).Bit0=1)   ;GPUSTAT.15
        (Above might be chipselect for (absent) second VRAM chip?)
12    Textured Rectangle X-Flip   (BIOS does set this bit on power-up...?)
13    Textured Rectangle Y-Flip   (BIOS does set it equal to GPUSTAT.13...?)
14-23 Not used (should be 0)
24-31 Command  (E1h)

*/
func (gpu *GPU) GP0DrawModeSet(data uint32) {
	gpu.txBase = GetValue(data, 0, 4)
	gpu.tyBase = GetValue(data, 4, 1)
	gpu.semiTransparency = uint8(GetValue(data, 5, 2))
	gpu.textureFormat = uint8(GetValue(data, 7, 2))
	gpu.dilthering = TestBit(data, 9)
	gpu.drawToDisplay = TestBit(data, 10)
	gpu.textureDisable = TestBit(data, 11)
	gpu.rectTextureXFlip = TestBit(data, 12)
	gpu.rectTextureYFlip = TestBit(data, 13)
}
