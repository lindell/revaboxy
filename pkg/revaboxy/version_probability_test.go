package revaboxy

import (
	"math"
	"math/rand"
	"testing"
)

func TestNormalVersionProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	err := vv.add(Version{
		Name:        "test1",
		Probability: 0.1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = vv.add(Version{
		Name:        DefaultName,
		Probability: 0.9,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := vv.valid(); err != nil {
		t.Error("versions should be valid", err)
	}

	if !versionProbabilityWithinRange(vv, "test1", 0.1, 0.01) {
		t.Fatal("version not selected within margin")
	}
}

func TestToMuchProbabilityVersionProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	err := vv.add(Version{
		Name:        "test1",
		Probability: 0.6,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = vv.add(Version{
		Name:        DefaultName,
		Probability: 0.6,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := vv.valid(); err == nil {
		t.Error("versions should be invalid", err)
	}
}
func TestDublicateNameVersionProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	err := vv.add(Version{
		Name:        "test1",
		Probability: 0.3,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = vv.add(Version{
		Name:        "test1",
		Probability: 0.6,
	})

	if err == nil { // NB
		t.Fatal("versions should be invalid")
	}
}

func TestDefaultNameRestProbability(t *testing.T) {
	rand.Seed(1)

	vv := &versions{}
	err := vv.add(Version{
		Name:        "test1",
		Probability: 0.1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = vv.add(Version{
		Name:        "test2",
		Probability: 0.2,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = vv.add(Version{
		Name:        DefaultName,
		Probability: 0.3,
	})
	if err != nil {
		t.Fatal(err)
	}

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
