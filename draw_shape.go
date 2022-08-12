package main

import "fmt"

const (
	SHAPE_POLYGON = iota
)

func (gpu *GPU) DrawShape() {
	fmt.Printf("[GPU::DrawShape] shape=%d\n", gpu.fifo.shape)
	fmt.Printf("[GPU::DrawShape] attr=%05b\n", gpu.fifo.attr)

	for i := 0; i < gpu.fifo.nArgs; i++ {
		fmt.Printf("[GPU::DrawShape] arg%d=%x\n", i, gpu.fifo.buffer[i])
	}

	// TODO

	gpu.fifo.Done()
}
