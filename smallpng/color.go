package smallpng

import (
	"image"
	"image/color"
	"math/rand"
	"runtime"
	"sync"
)

// DefaultMaxKMeansIters is the default maximum number of
// iterations of the k-means algorithm for clustering.
const DefaultMaxKMeansIters = 5

const (
	maxClusterPixels = 100000
)

// PaletteImage creates a color palette for an image using
// clustering to minimize the discrepency from reduced
// colors.
//
// If maxIters is non-zero, then it limits the number of
// k-means iterations for clustering.
// Otherwise, DefaultMaxKMeansIters is used.
func PaletteImage(img image.Image, maxIters int) *image.Paletted {
	if maxIters == 0 {
		maxIters = DefaultMaxKMeansIters
	}
	bounds := img.Bounds()
	colors := make([]rgbaColor, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			colors = append(colors, newRGBAColor(img.At(x, y)))
		}
	}
	colors = subsampleClusterPixels(colors)

	clusters := newColorClusters(colors, 256)
	loss := clusters.Iterate()
	for i := 0; i < maxIters; i++ {
		newLoss := clusters.Iterate()
		if newLoss >= loss {
			break
		}
		loss = newLoss
	}
	palette := make(color.Palette, 256)
	for i, x := range clusters.Centers {
		palette[i] = x.Color()
	}

	// Prevent nil colors in palette.
	for i := len(clusters.Centers); i < len(palette); i++ {
		palette[i] = palette[0]
	}

	res := image.NewPaletted(bounds, palette)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			res.Set(x, y, img.At(x, y))
		}
	}
	return res
}

func subsampleClusterPixels(colors []rgbaColor) []rgbaColor {
	if len(colors) <= maxClusterPixels {
		return colors
	}
	for i := 0; i < maxClusterPixels; i++ {
		j := rand.Intn(len(colors) - i)
		colors[i], colors[j] = colors[j], colors[i]
	}
	return colors[:maxClusterPixels]
}

type rgbaColor [4]float32

func newRGBAColor(c color.Color) rgbaColor {
	r, g, b, a := c.RGBA()
	return rgbaColor{
		float32(r) / 0xffff,
		float32(g) / 0xffff,
		float32(b) / 0xffff,
		float32(a) / 0xffff,
	}
}

func (r rgbaColor) DistSquared(r1 rgbaColor) float32 {
	var res float32
	for i, x := range r {
		d := x - r1[i]
		res += d * d
	}
	return res
}

func (r rgbaColor) Add(r1 rgbaColor) rgbaColor {
	for i, x := range r1 {
		r[i] += x
	}
	return r
}

func (r rgbaColor) Scale(s float32) rgbaColor {
	for i, x := range r {
		r[i] = x * s
	}
	return r
}

func (r rgbaColor) Color() color.RGBA {
	for i, x := range r {
		if x < 0 {
			r[i] = 0
		} else if x > 1 {
			r[i] = 1
		}
	}
	return color.RGBA{
		R: uint8(r[0] * 255.999),
		G: uint8(r[1] * 255.999),
		B: uint8(r[2] * 255.999),
		A: uint8(r[3] * 255.999),
	}
}

type colorClusters struct {
	Centers   []rgbaColor
	AllColors []rgbaColor
}

func newColorClusters(allColors []rgbaColor, numCenters int) *colorClusters {
	// Optimization for the case where there are enough
	// centers to cover every mode exactly.
	uniqueColors := map[rgbaColor]bool{}
	for _, c := range allColors {
		uniqueColors[c] = true
	}
	if len(uniqueColors) <= numCenters {
		unique := make([]rgbaColor, 0, len(uniqueColors))
		for c := range uniqueColors {
			unique = append(unique, c)
		}
		return &colorClusters{
			Centers:   unique,
			AllColors: allColors,
		}
	}

	return &colorClusters{
		Centers:   kmeansPlusPlusInit(allColors, numCenters),
		AllColors: allColors,
	}
}

// Iterate performs a step of k-means and returns the
// current MSE loss.
// If the MSE loss does not decrease, then the process has
// converged.
func (c *colorClusters) Iterate() float64 {
	centerSum := make([]rgbaColor, len(c.Centers))
	centerCount := make([]int, len(c.Centers))
	totalError := 0.0

	numProcs := runtime.GOMAXPROCS(0)
	var resultLock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < numProcs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			localCenterSum := make([]rgbaColor, len(c.Centers))
			localCenterCount := make([]int, len(c.Centers))
			localTotalError := 0.0
			for i := idx; i < len(c.AllColors); i += numProcs {
				co := c.AllColors[i]
				closestDist := 0.0
				closestIdx := 0
				for i, center := range c.Centers {
					d := float64(co.DistSquared(center))
					if d < closestDist || i == 0 {
						closestDist = d
						closestIdx = i
					}
				}
				localCenterSum[closestIdx] = localCenterSum[closestIdx].Add(co)
				localCenterCount[closestIdx]++
				localTotalError += closestDist
			}
			resultLock.Lock()
			defer resultLock.Unlock()
			for i, c := range localCenterCount {
				centerCount[i] += c
			}
			for i, s := range localCenterSum {
				centerSum[i] = centerSum[i].Add(s)
			}
			totalError += localTotalError
		}(i)
	}
	wg.Wait()

	for i, newCenter := range centerSum {
		count := centerCount[i]
		if count > 0 {
			c.Centers[i] = newCenter.Scale(1 / float32(count))
		}
	}

	return totalError / float64(len(c.AllColors))
}

func kmeansPlusPlusInit(allColors []rgbaColor, numCenters int) []rgbaColor {
	centers := make([]rgbaColor, numCenters)
	centers[0] = allColors[rand.Intn(len(allColors))]
	dists := newCenterDistances(allColors, centers[0])
	for i := 1; i < numCenters; i++ {
		sampleIdx := dists.Sample()
		centers[i] = allColors[sampleIdx]
		dists.Update(centers[i])
	}
	return centers
}

type centerDistances struct {
	AllColors   []rgbaColor
	Distances   []float64
	DistanceSum float64
}

func newCenterDistances(allColors []rgbaColor, center rgbaColor) *centerDistances {
	dists := make([]float64, len(allColors))
	sum := 0.0
	for i, c := range allColors {
		dists[i] = float64(c.DistSquared(center))
		sum += dists[i]
	}
	return &centerDistances{
		AllColors:   allColors,
		Distances:   dists,
		DistanceSum: sum,
	}
}

func (c *centerDistances) Update(newCenter rgbaColor) {
	c.DistanceSum = 0
	for i, co := range c.AllColors {
		d := float64(co.DistSquared(newCenter))
		if d < c.Distances[i] {
			c.Distances[i] = d
		}
		c.DistanceSum += c.Distances[i]
	}
}

func (c *centerDistances) Sample() int {
	sample := rand.Float64() * c.DistanceSum
	idx := len(c.AllColors) - 1
	for i, dist := range c.Distances {
		sample -= dist
		if sample < 0 {
			idx = i
			break
		}
	}
	return idx
}
