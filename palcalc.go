package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ColorPoint struct {
	color    FloatColor
	segment  int
	count    uint64
	distance float64
}

type PalCalc struct {
	points    []ColorPoint
	centroids []FloatColor

	colors    int
	poinCount uint64

	totalDistance float64
	pointsChanged uint64

	workers     int
	pointRanges [][]ColorPoint

	bestError   float64
	bestPalette Palette
	bestAtt     int
	errors      []float64

	maxSteps   int
	maxAttempt int
}

func swapPoints(left, right *ColorPoint) {
	*left, *right = *right, *left
}

func NewPalCalc(colors int, steps int, attempt int) *PalCalc {
	return &PalCalc{colors: colors, maxSteps: steps, maxAttempt: attempt}
}

func (km *PalCalc) Input(images [][]IntColor, levels int) {
	var cube [256][256][256]uint64

	for _, img := range images {
		for _, data := range img {
			cube[data.R][data.G][data.B]++
		}
	}

	colors_total := uint64(0)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if cube[r][g][b] > 0 {
					colors_total++
				}
			}
		}
	}

	fmt.Printf("\n\nTotal number of pure colors: %d\n", colors_total)
	colors_total *= uint64(levels)
	fmt.Printf("\n\nTotal number of all colors: %d\n", colors_total)
	if colors_total == 0 {
		panic(errors.New("wrong input"))
	}
	if uint64(km.colors) > colors_total {
		km.colors = int(colors_total)
	}
	km.poinCount = colors_total

	km.points = make([]ColorPoint, 0, colors_total)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if cube[r][g][b] > 0 {
					km.points = append(km.points, ColorPoint{
						color:    FloatColor{float64(r) / 255, float64(g) / 255, float64(b) / 255},
						segment:  0,
						count:    cube[r][g][b],
						distance: math.MaxFloat64})
					for l := 1; l <= levels-1; l++ {
						k := float64(l) / float64(levels)
						km.points = append(km.points, ColorPoint{
							color:    FloatColor{float64(r) / 255 * k, float64(g) / 255 * k, float64(b) / 255 * k},
							segment:  0,
							count:    cube[r][g][b],
							distance: math.MaxFloat64})
					}
				}
			}
		}
	}

	km.workers = runtime.NumCPU()
	if km.workers > 1 {
		km.pointRanges = make([][]ColorPoint, km.workers)
		rangeSize := len(km.points) / km.workers
		for i := 0; i < km.workers-1; i++ {
			km.pointRanges[i] = km.points[i*rangeSize : (i+1)*rangeSize]
		}
		km.pointRanges[km.workers-1] = km.points[(km.workers-1)*rangeSize:]
	}
}

func (point *ColorPoint) pointDistance(center *ColorPoint) float64 {
	dist := point.color.Distance(center.color)
	if dist < point.distance {
		point.distance = dist
		return dist
	}
	return point.distance
}

func (km *PalCalc) initCentroids() {
	centInd := 0
	swapPoints(&km.points[0], &km.points[rand.Uint64()%km.poinCount])
	for centInd < km.colors-1 {
		var sum float64 = 0
		for i := uint64(centInd + 1); i < km.poinCount; i++ {
			sum += km.points[i].pointDistance(&km.points[centInd])
		}
		rnd := rand.Float64() * sum
		centInd++
		sum = 0
		next := km.poinCount - 1
		for i := uint64(centInd + 1); i < km.poinCount; i++ {
			sum += km.points[i].distance
			if sum > rnd {
				next = i
				break
			}
		}
		swapPoints(&km.points[centInd], &km.points[next])
	}

	km.centroids = make([]FloatColor, km.colors)
	for i := 0; i < km.colors; i++ {
		km.centroids[i] = km.points[i].color
	}
}

func (km *PalCalc) calcCentroids() {
	//start := time.Now()
	newCentroids := make([]FloatColor, km.colors)
	sizes := make([]uint64, km.colors)
	for _, point := range km.points {
		sizes[point.segment] += point.count
		c := &newCentroids[point.segment]
		c.R += point.color.R * float64(point.count)
		c.G += point.color.G * float64(point.count)
		c.B += point.color.B * float64(point.count)
	}
	km.totalDistance = 0
	for i := range km.centroids {
		if sizes[i] == 0 {
			continue
		}
		size := float64(sizes[i])
		newCentroids[i].R /= size
		newCentroids[i].G /= size
		newCentroids[i].B /= size
		km.totalDistance += math.Sqrt(newCentroids[i].Distance(km.centroids[i]))
		km.centroids[i] = newCentroids[i]
	}
	//fmt.Printf("Centroids: %s   ", time.Since(start))
}

func (km *PalCalc) calcSegments() {
	var (
		mt sync.Mutex
		wg sync.WaitGroup
	)

	//start := time.Now()
	km.pointsChanged = 0
	for _, task := range km.pointRanges {
		wg.Add(1)
		go func(chunk []ColorPoint) {
			for i := range chunk {
				oldSeg := chunk[i].segment
				newSeg := oldSeg
				minDist := chunk[i].color.Distance(km.centroids[oldSeg])
				for c := range km.centroids {
					dist := chunk[i].color.Distance(km.centroids[c])
					if dist < minDist {
						minDist = dist
						newSeg = c
					}
				}
				if oldSeg != newSeg {
					chunk[i].segment = newSeg
					mt.Lock()
					km.pointsChanged++
					mt.Unlock()
				}
			}
			wg.Done()
		}(task)
	}
	wg.Wait()

	//fmt.Printf("SegmentsMt: %s\n", time.Since(start))
}

func formatTime(dur time.Duration) string {
	var result strings.Builder
	d := dur.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		fmt.Fprintf(&result, "%2d h ", int(h))
	} else {
		fmt.Fprint(&result, "     ")
	}
	if m > 0 {
		fmt.Fprintf(&result, "%2d m ", int(m))
	} else {
		fmt.Fprint(&result, "     ")
	}
	fmt.Fprintf(&result, "%2d s", int(s))
	return result.String()
}

func (km *PalCalc) printState(attempt int, step int, start time.Time) {
	elapsed := time.Since(start)
	remainingSteps := km.maxSteps*km.maxAttempt - step - (attempt-1)*km.maxSteps
	remaining := elapsed * time.Duration(remainingSteps) / time.Duration(step+(attempt-1)*km.maxSteps)
	fmt.Printf("\r Att %2d / %d | Step %4d / %d | Dist %10.5g | Ch %10d | El %s | Rem %s  ",
		attempt,
		km.maxAttempt,
		step,
		km.maxSteps,
		km.totalDistance,
		km.pointsChanged,
		formatTime(elapsed),
		formatTime(remaining))
}

func (km *PalCalc) CalcError() float64 {
	score := float64(0)
	for _, point := range km.points {
		score += math.Sqrt(point.color.Distance(km.centroids[point.segment])) * float64(point.count)
	}
	return score
}

func (km *PalCalc) Run() {
	km.errors = make([]float64, 0, km.maxAttempt)
	startTime := time.Now()
	for a := 1; a < km.maxAttempt+1; a++ {
		km.initCentroids()
		for i := 1; i < km.maxSteps+1; i++ {
			km.calcSegments()
			if km.pointsChanged == 0 {
				km.printState(a, i, startTime)
				break
			}
			km.calcCentroids()
			km.printState(a, i, startTime)
		}
		km.calcSegments()
		colorErr := km.CalcError()
		if a == 1 || colorErr < km.bestError {
			km.bestAtt = a
			km.bestError = colorErr
			km.bestPalette = km.calcPalette()
		}
		km.errors = append(km.errors, colorErr)
	}
	fmt.Printf("\nMost successful attempt is %d\n", km.bestAtt)
	fmt.Print(km.errors)
	fmt.Println()
}

func (km *PalCalc) calcPalette() Palette {
	result := make(Palette, 0, km.colors+1)
	for _, c := range km.centroids {
		result = append(result, c.ToIntColor().Normalized())

	}
	result = append(result, IntColor{0, 0, 0})
	//result.Sort()
	return result
}

func (km *PalCalc) GetPalette() Palette {
	return km.bestPalette
}
