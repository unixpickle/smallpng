package smallpng

import (
	"image"
	"image/png"
	"os"
)

type Config struct {
	NoPalette        bool
	PaletteSize      int
	MaxIters         int
	MaxClusterPixels int
}

// CompressImage reads an image from inPath and saves the
// compressed version to outPath.
func CompressImage(inPath, outPath string, c *Config) error {
	if c == nil {
		c = &Config{}
	}
	img, err := ReadImage(inPath)
	if err != nil {
		return err
	}
	if !c.NoPalette {
		img = PaletteImage(img, &PaletteConfig{
			MaxKMeansIters:   c.MaxIters,
			PaletteSize:      c.PaletteSize,
			MaxClusterPixels: c.MaxClusterPixels,
		})
	}
	return WriteImage(outPath, img)
}

func ReadImage(path string) (image.Image, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return png.Decode(r)
}

func WriteImage(path string, img image.Image) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	return enc.Encode(w, img)
}
