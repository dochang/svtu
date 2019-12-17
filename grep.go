package svtu

import (
	"github.com/blang/semver"
)

func RemoveInvalidSemver(input []string) (output []semver.Version) {
	for _, line := range input {
		v, err := semver.Parse(line)
		if err != nil {
			continue
		}
		output = append(output, v)
	}
	return
}

func RemoveOutOfRanges(svRanges []semver.Range, input []semver.Version) (output []semver.Version) {
	for _, v := range input {
		ok := true
		for _, r := range svRanges {
			if !(r(v)) {
				ok = false
				break
			}
		}
		if ok {
			output = append(output, v)
		}
	}
	return
}
