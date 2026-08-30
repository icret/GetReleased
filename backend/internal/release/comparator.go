package release

import "github.com/Masterminds/semver/v3"

func IsNewer(current, candidate string) bool {
	if current == candidate {
		return false
	}
	c, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	n, err := semver.NewVersion(candidate)
	if err != nil {
		return false
	}
	return n.GreaterThan(c)
}
