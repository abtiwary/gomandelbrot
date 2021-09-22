package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)


func mapToRange(val, in_min, in_max, out_min, out_max float64) float64 {
	return (val - in_min) * (out_max - out_min) / (in_max - in_min) + out_min;
}

func main() {
	const (
		width = 800
		height = 800
		min = -2.84
		max = 2.04
		max_iterations = 200
	)

	img := image.NewRGBA(image.Rectangle{
		image.Point{0, 0},
		image.Point{width, height},
	})

	var i int64
	var j int64
	for i=0; i < width; i++ {
		for j=0; j < height; j++ {
			//fmt.Println(i, j)
			x := mapToRange(float64(i), 0, width, min, max)
			y := mapToRange(float64(j), 0, width, min, max)
			x0 := x
			y0 := y
			iters := 0

			var z int64
			for z=0; z < max_iterations; z++ {
				//fmt.Println("z = ", z)
				x1 := x * x - y * y
				y1 := 2 * x * y

				x = x1 + x0
				y = y1 + y0

				if x + y > 2 {
					break
				}

				iters += 1
			}

			col := mapToRange(float64(iters), 0, max_iterations, 0, 255)
			if iters == max_iterations || col < 20 {
				col = 0
			}

			img.Set(int(i), int(j), color.RGBA{uint8(col), uint8(col), uint8(col), 255})

		}

	}


	f, _ := os.Create("/Users/abtiwary/temp/image.png")
	png.Encode(f, img)


}
