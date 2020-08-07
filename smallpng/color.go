package smallpng

import (
	"image/color"
	"math"
)

type ColorSpace int

const (
	RGB ColorSpace = iota
	CIELAB
)

const labAlphaScale = 128.0

func (c ColorSpace) toColor(v colorVector) color.Color {
	switch c {
	case RGB:
		return color.RGBA{
			R: uint8(v[0] * 255.999),
			G: uint8(v[1] * 255.999),
			B: uint8(v[2] * 255.999),
			A: uint8(v[3] * 255.999),
		}
	case CIELAB:
		xyz := convertLabToXYZ([3]float32{v[0], v[1], v[2]})
		lrgb := convertXYZToLinearRGB(xyz)
		rgb := convertLinearRGBToSRGB(lrgb)
		return RGB.toColor(colorVector{rgb[0], rgb[1], rgb[2], v[3] / labAlphaScale})
	default:
		panic("unknown color space")
	}
}

func (c ColorSpace) toVector(co color.Color) colorVector {
	rInt, gInt, bInt, aInt := co.RGBA()

	r := float32(rInt) / 0xffff
	g := float32(gInt) / 0xffff
	b := float32(bInt) / 0xffff
	a := float32(aInt) / 0xffff

	switch c {
	case RGB:
		return colorVector{r, g, b, a}
	case CIELAB:
		lrgb := convertSRGBToLinearRGB([3]float32{r, g, b})
		xyz := convertLinearRGBToXYZ(lrgb)
		lab := convertXYZToLab(xyz)
		return colorVector{lab[0], lab[1], lab[2], a * labAlphaScale}
	default:
		panic("unknown color space")
	}
}

type colorVector [4]float32

func (c colorVector) DistSquared(c1 colorVector) float32 {
	var res float32
	for i, x := range c {
		d := x - c1[i]
		res += d * d
	}
	return res
}

func (c colorVector) Add(c1 colorVector) colorVector {
	for i, x := range c1 {
		c[i] += x
	}
	return c
}

func (c colorVector) Scale(s float32) colorVector {
	for i, x := range c {
		c[i] = x * s
	}
	return c
}

// convertXYZToLab creates a CIELAB color from an XYZ
// color.
func convertXYZToLab(xyz [3]float32) [3]float32 {
	x := xyz[0] * 100
	y := xyz[1] * 100
	z := xyz[2] * 100

	// Illuminant D65
	const xn float32 = 95.0489
	const yn float32 = 100.0
	const zn float32 = 108.8840

	f := func(t float32) float32 {
		const delta float32 = 6.0 / 29.0
		if t > delta*delta*delta {
			return float32(math.Cbrt(float64(t)))
		} else {
			return t/(3*delta*delta) + 4.0/29.0
		}
	}

	fx := f(x / xn)
	fy := f(y / yn)
	fz := f(z / zn)
	return [3]float32{
		116.0*fy - 16.0,
		500.0 * (fx - fy),
		200.0 * (fy - fz),
	}
}

// convertLabToXYZ creates an XYZ color from a CIELAB
// color.
func convertLabToXYZ(lab [3]float32) [3]float32 {
	// Illuminant D65
	const xn float32 = 95.0489
	const yn float32 = 100.0
	const zn float32 = 108.8840

	l := lab[0]
	a := lab[1]
	b := lab[2]

	f := func(t float32) float32 {
		const delta float32 = 6.0 / 29.0
		if t > delta {
			return t * t * t
		} else {
			return 3 * delta * delta * (t - 4.0/29.0)
		}
	}

	return [3]float32{
		xn * f((l+16.0)/116.0+a/500.0) / 100.0,
		yn * f((l+16.0)/116.0) / 100.0,
		zn * f((l+16.0)/116.0-b/200.0) / 100.0,
	}
}

// convertLinearRGBToXYZ converts a linear RGB color to an
// XYZ color.
func convertLinearRGBToXYZ(rgb [3]float32) [3]float32 {
	return [3]float32{
		0.41239080*rgb[0] + 0.35758434*rgb[1] + 0.18048079*rgb[2],
		0.21263901*rgb[0] + 0.71516868*rgb[1] + 0.07219232*rgb[2],
		0.01933082*rgb[0] + 0.11919478*rgb[1] + 0.95053215*rgb[2],
	}
}

// convertXYZToLinearRGB converts an XYZ color to a linear
// RGB color.
func convertXYZToLinearRGB(xyz [3]float32) [3]float32 {
	return [3]float32{
		3.24096994*xyz[0] - 1.53738318*xyz[1] - 0.49861076*xyz[2],
		-0.96924364*xyz[0] + 1.8759675*xyz[1] + 0.04155506*xyz[2],
		0.05563008*xyz[0] - 0.20397696*xyz[1] + 1.05697151*xyz[2],
	}
}

// convertSRGBToLinearRGB converts an sRGB color to a
// linear RGB color.
func convertSRGBToLinearRGB(srgb [3]float32) [3]float32 {
	res := [3]float32{}
	for i, x := range srgb {
		res[i] = gammaExpand(x)
	}
	return res
}

// convertLinearRGBToSRGB converts a linear RGB color to
// an sRGB color.
func convertLinearRGBToSRGB(rgb [3]float32) [3]float32 {
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
