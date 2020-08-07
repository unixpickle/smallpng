package smallpng

import (
	"math"
	"math/rand"
	"testing"
)

func TestConvertXYZToLab(t *testing.T) {
	// Test from https://github.com/cangoektas/xyz-to-lab.
	xyz := [3]float32{0.77, 0.9278, 0.1385}
	expected := [3]float32{97.13824698129729, -21.555908334832285, 94.48248544644461}
	actual := convertXYZToLab(xyz)
	for i, x := range expected {
		a := actual[i]
		if math.Abs(float64(a-x)) > 0.01 {
			t.Errorf("expected %f but got %f", x, a)
		}
	}
}

func TestLabXYZInverses(t *testing.T) {
	for i := 0; i < 10; i++ {
		xyz := [3]float32{rand.Float32(), rand.Float32(), rand.Float32()}
		lab := convertXYZToLab(xyz)
		xyz1 := convertLabToXYZ(lab)
		for i, x := range xyz {
			a := xyz1[i]
			if math.Abs(float64(x-a)) > 1e-4 {
				t.Errorf("component %d: expected %f but got %f", i, x, a)
			}
		}
	}
}

func TestXYZRGBInverses(t *testing.T) {
	for i := 0; i < 10; i++ {
		rgb := [3]float32{rand.Float32(), rand.Float32(), rand.Float32()}
		xyz := convertLinearRGBToXYZ(rgb)
		rgb1 := convertXYZToLinearRGB(xyz)
		for i, x := range rgb {
			a := rgb1[i]
			if math.Abs(float64(x-a)) > 1e-4 {
				t.Errorf("component %d: expected %f but got %f", i, x, a)
			}
		}
	}
}
