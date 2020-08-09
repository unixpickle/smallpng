# smallpng

**smallpng** is a library and CLI tool to compress PNG files.

# How it works

**smallpng** uses a lossy compression algorithm to generate a color palette for a given input image. When a new PNG is encoded with this smaller color palette, the file is smaller since color information has been thrown away. For many purposes, small color palettes result in acceptable quality.

# Usage

To compress an image with the `smallpng` command, simply run:

```
$ smallpng input.png output.png
```

Here is the full usage information:

```
Usage: smallpng [flags] <input.png> [output.png]

  -max-cluster-pixels int
        maximum number of pixels to use as data points for clustering (default 100000)
  -max-iters int
        maximum number of clustering iterations (more iterations means better clusters) (default 5)
  -no-palette
        use the original color space, not a palette
  -palette-size int
        number of colors in the color palette (default 256)
```

# Results

Here are some examples of images compressed using **smallpng**:

| Input | 256 Colors | 64 Colors |
|-------|------------|-----------|
| ![](examples/image1_input.png) | ![](examples/image1_output_256.png) | ![](examples/image1_output_64.png) |
| ![](examples/image2_input.png) | ![](examples/image2_output_256.png) | ![](examples/image2_output_64.png) |
