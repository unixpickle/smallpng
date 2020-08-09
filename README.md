# smallpng

**smallpng** is a library and CLI tool to compress PNG files.

# How it works

**smallpng** uses a lossy compression algorithm to generate a color palette for a given input image. When a new PNG is encoded with this smaller color palette, the file is smaller since color information has been thrown away. For many purposes, small color palettes result in acceptable quality.

# Results

Here are some examples of images compressed using **smallpng**:

| Input | 256 Colors | 64 Colors |
|-------|------------|-----------|
| ![](examples/image1_input.png) | ![](examples/image1_output_256.png) | ![](examples/image1_output_64.png) |
| ![](examples/image2_input.png) | ![](examples/image2_output_256.png) | ![](examples/image2_output_64.png) |
