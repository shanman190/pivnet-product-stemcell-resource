package versions

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/go-pivnet/v7"
)

const (
	fingerprintDelimiter = "#"
)

// Since : slice the string array to return all versions since the one specified. If no version is found, then return just the first one.
func Since(versions []string, since string) ([]string, error) {
	for i, v := range versions {
		if v == since {
			return versions[:i+1], nil
		}
	}

	return versions[:1], nil
}

// SinceRelease : slice the pivnet.Release array to return all versions since the one specified. If no version is found, then return just the first one.
func SinceRelease(versions []pivnet.Release, since string) ([]pivnet.Release, error) {
	for i, v := range versions {
		if v.Version == since {
			return versions[:i+1], nil
		}
	}

	return versions[:1], nil
}

// Reverse : reverses the version array
func Reverse(versions []string) ([]string, error) {
	var reversed []string
	for i := len(versions) - 1; i >= 0; i-- {
		reversed = append(reversed, versions[i])
	}

	return reversed, nil
}

// SplitIntoVersionAndFingerprint : splits a structured version into it's version and fingerprint parts
func SplitIntoVersionAndFingerprint(versionWithFingerprint string) (string, string, error) {
	split := strings.Split(versionWithFingerprint, fingerprintDelimiter)
	if len(split) != 2 {
		return "", "", fmt.Errorf("Invalid version and Fingerprint: %s", versionWithFingerprint)
	}
	return split[0], split[1], nil
}

// CombineVersionAndFingerprint : combine version and fingerprint into the structured version
func CombineVersionAndFingerprint(version string, fingerprint string) (string, error) {
	if fingerprint == "" {
		return version, nil
	}
	return combineVersionAndFingerprint(version, fingerprint), nil
}

func combineVersionAndFingerprint(version string, fingerprint string) string {
	return fmt.Sprintf("%s%s%s", version, fingerprintDelimiter, fingerprint)
}