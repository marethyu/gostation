package main

import (
	"fmt"
)

const (
	SHAPE_POLYGON = iota
)

/*
1--2    Note: A 4 point polygon is processed internally as two 3 point
|  |    polygons.
3--4    Note: When drawing a polygon the GPU will not draw the right

	most and bottom edge. So a (0,0)-(32,32) rectangle will actually

be drawn as (0,0)-(31,31). Make sure adjoining polygons have the same
coordinates if you want them to touch eachother!. Haven't checked how this
works with 3 point polygons.
*/
func (gpu *GPU) DoRenderPolygon() {
	switch gpu.shape_attr {
	case 0b01100:
		// TODO
	case 0b10000:
		gpu.RenderTrigGouraud() // TODO something wrong here
	case 0b01000:
		gpu.RenderMonochromeQuad()
	case 0b11000:
		gpu.RenderQuadGouraud()
	default:
		panic(fmt.Sprintf("[GPU::DoRenderPolygon] Unknown attribute: %05b\n", gpu.shape_attr))
	}
}

func (gpu *GPU) RenderTrigGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0])
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2])
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4])

	// TODO ignore rightmost and bottom edges

	fmt.Printf("[GPU::RenderTrigGouraud] x1=%d,y1=%d,x2=%d,y2=%d,x3=%d,y3=%d,r=%d,g=%d,b=%d\n", v1.x, v1.y, v2.x, v2.y, v3.x, v3.y, v1.r, v1.g, v1.b)

	// make sure that vertexes are in clockwise order
	gpu.ShadedTriangle(v1, v2, v3)
}

func (gpu *GPU) RenderMonochromeQuad() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour)
	v2 := NewVertex(gpu.fifo.buffer[2], colour)
	v3 := NewVertex(gpu.fifo.buffer[3], colour)
	v4 := NewVertex(gpu.fifo.buffer[4], colour)

	// ignore rightmost and bottom edges
	v2.x -= 1
	v4.x -= 1
	v3.y -= 1
	v4.y -= 1

	fmt.Printf("[GPU::RenderMonochromeQuad] x1=%d,y1=%d,x2=%d,y2=%d,x3=%d,y3=%d,x4=%d,y4=%d,r=%d,g=%d,b=%d\n", v1.x, v1.y, v2.x, v2.y, v3.x, v3.y, v4.x, v4.y, v1.r, v1.g, v1.b)

	// make sure that vertexes are in clockwise order
	gpu.ShadedTriangle(v1, v2, v3)
	gpu.ShadedTriangle(v2, v4, v3)
}

func (gpu *GPU) RenderQuadGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0])
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2])
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4])
	v4 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6])

	// ignore rightmost and bottom edges
	v2.x -= 1
	v4.x -= 1
	v3.y -= 1
	v4.y -= 1

	fmt.Printf("[GPU::RenderQuadGouraud] x1=%d,y1=%d,x2=%d,y2=%d,x3=%d,y3=%d,x4=%d,y4=%d,r=%d,g=%d,b=%d\n", v1.x, v1.y, v2.x, v2.y, v3.x, v3.y, v4.x, v4.y, v1.r, v1.g, v1.b)

	// make sure that vertexes are in clockwise order
	gpu.ShadedTriangle(v1, v2, v3)
	gpu.ShadedTriangle(v2, v4, v3)
}

/* lol self plagiarised from https://github.com/marethyu/mgl/blob/main/mygl.h#L133 */
func (gpu *GPU) ShadedTriangle(v1, v2, v3 *Vertex) {
	// Area of the parallelogram formed by edge vectors
	area := float32(int32(v3.x-v1.x)*int32(v2.y-v1.y) - int32(v3.y-v1.y)*int32(v2.x-v1.x))

	// top left and bottom right points of a bounding box
	xmin := MinOf(v1.x, v2.x, v3.x)
	xmax := MaxOf(v1.x, v2.x, v3.x)
	ymin := MinOf(v1.y, v2.y, v3.y)
	ymax := MaxOf(v1.y, v2.y, v3.y)

	// TODO clipping

	for y := ymin; y <= ymax; y += 1 {
		for x := xmin; x <= xmax; x += 1 {
			// Barycentric weights
			w1 := float32(int32(x-v2.x)*int32(v3.y-v2.y)-int32(y-v2.y)*int32(v3.x-v2.x)) / area
			w2 := float32(int32(x-v3.x)*int32(v1.y-v3.y)-int32(y-v3.y)*int32(v1.x-v3.x)) / area
			w3 := float32(int32(x-v1.x)*int32(v2.y-v1.y)-int32(y-v1.y)*int32(v2.x-v1.x)) / area

			if (w1 >= 0.0) && (w2 >= 0.0) && (w3 >= 0.0) {
				r := uint8(w1*float32(v1.r) + w2*float32(v1.r) + w3*float32(v1.r))
				g := uint8(w1*float32(v1.g) + w2*float32(v1.g) + w3*float32(v1.g))
				b := uint8(w1*float32(v1.b) + w2*float32(v1.b) + w3*float32(v1.b))

				gpu.Pixel(uint32(x), uint32(y), r, g, b)
			}
		}
	}
}

func (gpu *GPU) Pixel(x uint32, y uint32, r, g, b uint8) {
	var colour uint32 = 0

	PackValue(&colour, 0, uint32(r>>3), 5)
	PackValue(&colour, 5, uint32(g>>3), 5)
	PackValue(&colour, 10, uint32(b>>3), 5)

	gpu.vram.Write16(x, y, uint16(colour))
}
