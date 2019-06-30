package revaboxy

import (
	"fmt"
	"math/rand"
)

type versions map[string]*Version

func (vv versions) valid() error {
	if _, ok := vv[DefaultName]; !ok {
		return fmt.Errorf("a version with the name %s needs to exist", DefaultName)
	}

	totalProbability := 0.0
	for _, v := range vv {
		totalProbability += v.Probability
	}
	if totalProbability > 1 {
		return fmt.Errorf("total percentage is more than 1")
	}

	return nil
}

func (vv versions) add(v Version) error {
	if _, ok := vv[v.Name]; ok {
		return fmt.Errorf("dublicate name \"%s\"", v.Name)
	}
	vv[v.Name] = &v

	return nil
}

func (vv versions) get(name string) *Version {
	v := vv[name]
	return v
}

func (vv versions) getRandomVersion() *Version {
	n := rand.Float64()

	addedProbability := 0.0
	for _, v := range vv {
		if n > addedProbability && n < addedProbability+v.Probability {
			return v
		}
		addedProbability += v.Probability
	}

	return vv[DefaultName]
}
