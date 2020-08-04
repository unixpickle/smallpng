package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/smallpng/smallpng"
)

func main() {
	var noPalette bool
	var maxIters int
	flag.BoolVar(&noPalette, "no-palette", false, "use the original color space, not a palette")
	flag.IntVar(&maxIters, "max-iters", smallpng.DefaultMaxKMeansIters,
		"maximum number of clustering iterations (more iterations means better clusters)")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ", os.Args[0], " [flags] <input.png> [output.png]")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	flag.Parse()

	if len(flag.Args()) != 1 && len(flag.Args()) != 2 {
		flag.Usage()
	}

	inputPath := flag.Args()[0]
	outputPath := inputPath
	if len(flag.Args()) == 2 {
		outputPath = flag.Args()[1]
	}

	img := ReadImage(inputPath)
	if !noPalette {
		img = smallpng.PaletteImage(img, maxIters)
	}
	WriteImage(outputPath, img)
}

func ReadImage(path string) image.Image {
	r, err := os.Open(path)
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return img
}

func WriteImage(path string, img image.Image) {
	w, err := os.Create(path)
	essentials.Must(err)
	defer w.Close()
	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	essentials.Must(enc.Encode(w, img))
}
