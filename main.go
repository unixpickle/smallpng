package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/smallpng/smallpng"
)

func main() {
	var config smallpng.Config
	flag.BoolVar(&config.NoPalette, "no-palette", false,
		"use the original color space, not a palette")
	flag.IntVar(&config.PaletteSize, "palette-size", smallpng.DefaultPaletteSize,
		"number of colors in the color palette")
	flag.IntVar(&config.MaxClusterPixels, "max-cluster-pixels", smallpng.DefaultMaxClusterPixels,
		"maximum number of pixels to use as data points for clustering")
	flag.IntVar(&config.MaxIters, "max-iters", smallpng.DefaultMaxKMeansIters,
		"maximum number of clustering iterations (more iterations means better clusters)")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[flags] <input.png> [output.png]")
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

	inStats, err := os.Stat(inputPath)
	essentials.Must(err)
	essentials.Must(smallpng.CompressImage(inputPath, outputPath, &config))
	outStats, err := os.Stat(outputPath)
	essentials.Must(err)

	fracReduction := float64(inStats.Size()-outStats.Size()) / float64(inStats.Size())
	fmt.Printf(
		"%s -> %s (%.1f%% reduction)",
		humanize.Bytes(uint64(inStats.Size())),
		humanize.Bytes(uint64(outStats.Size())),
		fracReduction*100,
	)
	fmt.Println()
}
