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
	return PaletteImageColorSpace(img, maxIters, CIELAB)
}

// PaletteImageColorSpace is like PaletteImage, but it
// allows you to configure which color space the colors
// are clustered in.
//
// Using Lab is more perceptually accurate than RGB.
func PaletteImageColorSpace(img image.Image, maxIters int, cs ColorSpace) *image.Paletted {
	if maxIters == 0 {
		maxIters = DefaultMaxKMeansIters
	}
	bounds := img.Bounds()
	colors := make([]colorVector, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			colors = append(colors, cs.toVector(img.At(x, y)))
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
		palette[i] = cs.toColor(x)
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

func subsampleClusterPixels(colors []colorVector) []colorVector {
	if len(colors) <= maxClusterPixels {
		return colors
	}
	for i := 0; i < maxClusterPixels; i++ {
		j := rand.Intn(len(colors) - i)
		colors[i], colors[j] = colors[j], colors[i]
	}
	return colors[:maxClusterPixels]
}

type colorClusters struct {
	Centers   []colorVector
	AllColors []colorVector
}

func newColorClusters(allColors []colorVector, numCenters int) *colorClusters {
	// Optimization for the case where there are enough
	// centers to cover every mode exactly.
	uniqueColors := map[colorVector]bool{}
	for _, c := range allColors {
		uniqueColors[c] = true
	}
	if len(uniqueColors) <= numCenters {
		unique := make([]colorVector, 0, len(uniqueColors))
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
	centerSum := make([]colorVector, len(c.Centers))
	centerCount := make([]int, len(c.Centers))
	totalError := 0.0

	numProcs := runtime.GOMAXPROCS(0)
	var resultLock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < numProcs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			localCenterSum := make([]colorVector, len(c.Centers))
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

func kmeansPlusPlusInit(allColors []colorVector, numCenters int) []colorVector {
	centers := make([]colorVector, numCenters)
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
	AllColors   []colorVector
	Distances   []float64
	DistanceSum float64
}

func newCenterDistances(allColors []colorVector, center colorVector) *centerDistances {
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

func (c *centerDistances) Update(newCenter colorVector) {
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
