package smallpng

import "math"

// ConvertLinearRGBToXYZ converts a linear RGB color to an
// XYZ color.
func ConvertLinearRGBToXYZ(rgb [3]float32) [3]float32 {
	return [3]float32{
		0.41239080*rgb[0] + 0.35758434*rgb[1] + 0.18048079*rgb[2],
		0.21263901*rgb[0] + 0.71516868*rgb[1] + 0.07219232*rgb[2],
		0.01933082*rgb[0] + 0.11919478*rgb[1] + 0.95053215*rgb[2],
	}
}

// ConvertXYZToLinearRGB converts an XYZ color to a linear
// RGB color.
func ConvertXYZToLinearRGB(xyz [3]float32) [3]float32 {
	return [3]float32{
		3.24096994*xyz[0] - 1.53738318*xyz[1] - 0.49861076*xyz[2],
		-0.96924364*xyz[0] + 1.8759675*xyz[1] + 0.04155506*xyz[2],
		0.05563008*xyz[0] - 0.20397696*xyz[1] + 1.05697151*xyz[2],
	}
}

// ConvertSRGBToLinearRGB converts an sRGB color to a
// linear RGB color.
func ConvertSRGBToLinearRGB(srgb [3]float32) [3]float32 {
	res := [3]float32{}
	for i, x := range srgb {
		res[i] = gammaExpand(x)
	}
	return res
}

// ConvertLinearRGBToSRGB converts a linear RGB color to
// an sRGB color.
func ConvertLinearRGBToSRGB(rgb [3]float32) [3]float32 {
	res := [3]float32{}
	for i, x := range rgb {
		res[i] = gammaCompress(x)
	}
	return res
}

func gammaCompress(u float32) float32 {
	if u <= 0.0031308 {
		return 12.92 * u
	} else {
		return 1.055*float32(math.Pow(float64(u), 1/2.4)) - 0.055
	}
}

func gammaExpand(u float32) float32 {
	if u <= 0.04045 {
		return u / 12.92
	} else {
		return float32(math.Pow((float64(u)+0.055)/1.055, 2.4))
	}
}
