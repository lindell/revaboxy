package revaboxy

import (
	"math"
	"math/rand"
	"testing"
)

func TestNormalVersionProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	vv.add(Version{
		Name:       "test1",
		Percentage: 0.1,
	})
	vv.add(Version{
		Name:       DefaultName,
		Percentage: 0.9,
	})

	if err := vv.valid(); err != nil {
		t.Error("versions should be valid", err)
	}

	if !versionProbabilityWithinRange(vv, "test1", 0.1, 0.01) {
		t.Fatal("version not selected within margin")
	}
}

func TestToMuchPercentageVersionProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	vv.add(Version{
		Name:       "test1",
		Percentage: 0.6,
	})
	vv.add(Version{
		Name:       DefaultName,
		Percentage: 0.6,
	})

	if err := vv.valid(); err == nil {
		t.Error("versions should be invalid", err)
	}
}
func TestDublicateNameVersionProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	vv.add(Version{
		Name:       "test1",
		Percentage: 0.3,
	})
	vv.add(Version{
		Name:       "test1",
		Percentage: 0.6,
	})

	if err := vv.valid(); err == nil {
		t.Error("versions should be invalid", err)
	}
}

func TestDefaultNameRestProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	vv.add(Version{
		Name:       "test1",
		Percentage: 0.1,
	})
	vv.add(Version{
		Name:       "test2",
		Percentage: 0.2,
	})
	vv.add(Version{
		Name:       DefaultName,
		Percentage: 0.3,
	})

	if err := vv.valid(); err != nil {
		t.Error("versions should be valid", err)
	}

	if !versionProbabilityWithinRange(vv, DefaultName, 0.7, 0.01) {
		t.Fatal("version not selected within margin")
	}
}

func versionProbabilityWithinRange(vv *versions, name string, percentage float64, maxDiff float64) bool {
	total := 10000
	ofName := 0
	for i := 0; i < total; i++ {
		if vv.getRandomVersion().Name == name {
			ofName++
		}
	}

	return math.Abs(float64(ofName)/float64(total)-percentage) < maxDiff
}
