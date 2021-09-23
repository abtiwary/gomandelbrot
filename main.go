package main

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Point struct {
	X     float64
	Y     float64
	Red   uint8
	Green uint8
	Blue  uint8
}

type Settings struct {
	Width         float64
	Height        float64
	Min           float64
	Max           float64
	MaxIterations int64
	Center        Point
}

type MandelbrotImage struct {
	mu        sync.Mutex
	RGBAImage *image.RGBA
}

func NewMandelbrotImage(xTopLeft, yTopLeft, xBottomRight, yBottomRight int) *MandelbrotImage {
	return &MandelbrotImage{
		RGBAImage: image.NewRGBA(image.Rectangle{
			image.Point{xTopLeft, yTopLeft},
			image.Point{xBottomRight, yBottomRight},
		}),
	}
}

func (mi *MandelbrotImage) DrawPoint(point Point) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	mi.RGBAImage.Set(int(point.X),
		int(point.Y),
		color.RGBA{
			point.Red,
			point.Green,
			point.Blue,
			255,
		},
	)
}

func (mi *MandelbrotImage) WriteImage(w io.Writer) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	err := png.Encode(w, mi.RGBAImage)
	if err != nil {
		return errors.Wrap(err, "error writing mandelbrot image")
	}
	return nil
}

func mapToRange(val, in_min, in_max, out_min, out_max float64) float64 {
	return (val-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

func imageWriter(mi *MandelbrotImage, jobs chan Point) {
	for {
		select {
		case pt := <-jobs:
			{
				mi.DrawPoint(pt)
			}

		default:
			{
			}
		}
	}
}

func mandelbrotWorker(wg *sync.WaitGroup, count *uint64, pt Point, jobs chan Point, settings Settings) {
	defer wg.Done()

	i := pt.X
	j := pt.Y
	//fmt.Println(i, j)

	x := mapToRange(float64(i), 0, settings.Width, settings.Min, settings.Max)
	y := mapToRange(float64(j), 0, settings.Height, settings.Min, settings.Max)

	x = x - settings.Center.X
	y = y - settings.Center.Y

	x0 := x
	y0 := y

	var iters int64
	var z int64
	for z = 0; z < settings.MaxIterations; z++ {
		//fmt.Println("z = ", z)
		x1 := x*x - y*y
		y1 := 2 * x * y
		x = x1 + x0
		y = y1 + y0

		if x+y > 2 {
			break
		}
		iters += 1
	}

	col := mapToRange(float64(iters), 0, float64(settings.MaxIterations), 0, 255)
	if iters == settings.MaxIterations || col < 20 {
		col = 0
	}

	red := mapToRange(col*col, 0, 255*255, 0, 255)
	green := mapToRange(col/2, 0, 255/2, 0, 255)
	blue := mapToRange(math.Sqrt(col), 0, math.Sqrt(255), 0, 255)
	outpt := Point{
		X:     i,
		Y:     j,
		Red:   uint8(red),
		Green: uint8(green),
		Blue:  uint8(blue),
	}
	jobs <- outpt
	atomic.AddUint64(count, 1)
	return
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	/*
		settings := Settings{
			Width:         800,
			Height:        800,
			Min:           -2.84,
			Max:           2.04,
			MaxIterationsStart: 200,
			MaxIterationsThreshold: 2000,
			Center: Point{
				X: 0.0015,
				Y: -0.80,
			},
		}
	*/

	settings := Settings{
		Width:         800,
		Height:        800,
		Min:           -1.00,
		Max:           1.00,
		MaxIterations: 200,
		Center: Point{
			X: 0.5,
			Y: 0.0,
		},
	}

	var wg sync.WaitGroup
	var jobCount uint64

	jobs := make(chan Point)

	mandelbrotImg := NewMandelbrotImage(0, 0, int(settings.Width), int(settings.Height))
	go imageWriter(mandelbrotImg, jobs)

	var i int64
	var j int64
	for i = 0; i < int64(settings.Width); i++ {
		for j = 0; j < int64(settings.Height); j++ {
			pt := Point{
				X: float64(i),
				Y: float64(j),
			}
			wg.Add(1)
			go mandelbrotWorker(&wg, &jobCount, pt, jobs, settings)
		}
	}

	wg.Wait()

	for {
		//fmt.Println(jobCount)
		if jobCount >= (uint64(settings.Width)*uint64(settings.Height) - 1) {
			close(jobs)
			break
		}
		time.Sleep(100)
	}

	outfile := "/Users/abtiwary/temp/image.png"
	f, _ := os.Create(outfile)
	defer f.Close()
	err := mandelbrotImg.WriteImage(f)
	if err != nil {
		log.WithError(err).Infof("could not write image to file at %s", outfile)
	}
}
