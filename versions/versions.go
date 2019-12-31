package versions

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/go-pivnet/v3"
)

const (
	fingerprintDelimiter = "#"
)

func Since(versions []string, since string) ([]string, error) {
	for i, v := range versions {
		if v == since {
			return versions[:i+1], nil
		}
	}

	return versions[:1], nil
}

func SinceRelease(versions []pivnet.Release, since string) ([]pivnet.Release, error) {
	for i, v := range versions {
		if v.Version == since {
			return versions[:i+1], nil
		}
	}

	return versions[:1], nil
}

func Reverse(versions []string) ([]string, error) {
	var reversed []string
	for i := len(versions) - 1; i >= 0; i-- {
		reversed = append(reversed, versions[i])
	}

	return reversed, nil
}

func SplitIntoVersionAndFingerprint(versionWithFingerprint string) (string, string, error) {
	split := strings.Split(versionWithFingerprint, fingerprintDelimiter)
	if len(split) != 2 {
		return "", "", fmt.Errorf("Invalid version and Fingerprint: %s", versionWithFingerprint)
	}
	return split[0], split[1], nil
}

func CombineVersionAndFingerprint(version string, fingerprint string) (string, error) {
	if fingerprint == "" {
		return version, nil
	}
	return combineVersionAndFingerprint(version, fingerprint), nil
}

func combineVersionAndFingerprint(version string, fingerprint string) string {
	return fmt.Sprintf("%s%s%s", version, fingerprintDelimiter, fingerprint)
}