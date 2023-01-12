package versions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	"github.com/pivotal-cf/go-pivnet/v7"

	"github.com/shanman190/pivnet-product-stemcell-resource/versions"
)

var _ = Describe("Versions", func() {
	Describe("Since", func() {
		var (
			allVersions []string
			version     string
		)

		BeforeEach(func() {
			allVersions = []string{}
			version = ""
		})

		Context("when the provided version is the newest", func() {
			BeforeEach(func() {
				allVersions = []string{"1.2.3#abc", "1.3.2#def"}
				version = "1.2.3#abc"
			})

			It("returns the latest version", func() {
				versions, _ := versions.Since(allVersions, version)

				Expect(versions).To(HaveLen(1))
				Expect(versions).To(Equal([]string{"1.2.3#abc"}))
			})
		})

		Context("when provided version is present but not the newest", func() {
			BeforeEach(func() {
				allVersions = []string{"newest version", "middle version", "older version", "last version"}
				version = "older version"
			})

			It("returns new versions", func() {
				versions, _ := versions.Since(allVersions, version)

				Expect(versions).To(Equal([]string{"newest version", "middle version", "older version"}))
			})
		})

		Context("When the version is not present", func() {
			BeforeEach(func() {
				allVersions = []string{"1.2.3#abc", "1.3.2#def"}
				version = "1.3.2"
			})

			It("returns the newest version", func() {
				versions, _ := versions.Since(allVersions, version)

				Expect(versions).To(Equal([]string{"1.2.3#abc"}))
			})
		})
	})

	Describe("SinceRelease", func() {
		var (
			allReleases []pivnet.Release
			version		string
		)

		BeforeEach(func() {
			allReleases = []pivnet.Release{}
			version = ""
		})

		Context("when the provided version is the newest", func() {
			BeforeEach(func() {
				allReleases = []pivnet.Release{
					pivnet.Release{
						ID: 1,
						Version: "1.2.3",
					},
					pivnet.Release {
						ID: 2,
						Version: "1.3.2",
					},
				}
				version = "1.2.3"
			})

			It("returns the latest release", func() {
				releases, _ := versions.SinceRelease(allReleases, version)

				Expect(releases).To(HaveLen(1))
				Expect(releases).To(Equal([]pivnet.Release{
					pivnet.Release{
						ID: 1,
						Version: "1.2.3",
					},
				}))
			})
		})

		Context("when provided version is present but not the newest", func() {
			BeforeEach(func() {
				allReleases = []pivnet.Release{
					pivnet.Release{
						ID: 1,
						Version: "newest version",
					},
					pivnet.Release{
						ID: 2,
						Version: "middle version",
					},
					pivnet.Release{
						ID: 3,
						Version: "older version",
					},
					pivnet.Release{
						ID: 4,
						Version: "last version",
					},
				}
				version = "older version"
			})

			It("returns new releases", func() {
				releases, _ := versions.SinceRelease(allReleases, version)

				Expect(releases).To(Equal([]pivnet.Release{
					pivnet.Release{
						ID: 1,
						Version: "newest version",
					},
					pivnet.Release{
						ID: 2,
						Version: "middle version",
					},
					pivnet.Release{
						ID: 3,
						Version: "older version",
					},
				}))
			})
		})

		Context("when the version is not present", func() {
			BeforeEach(func() {
				allReleases = []pivnet.Release{
					pivnet.Release{
						ID: 1,
						Version: "1.2.3",
					},
					pivnet.Release{
						ID: 2,
						Version: "1.3.2",
					},
				}
				version = "1.3.1"
			})

			It("returns thew newest release", func() {
				releases, _ := versions.SinceRelease(allReleases, version)

				Expect(releases).To(Equal([]pivnet.Release{
					pivnet.Release{
						ID: 1,
						Version: "1.2.3",
					},
				}))
			})
		})
	})

	Describe("Reverse", func() {
		It("returns reversed ordered versions because concourse expects them that way", func() {
			versions, err := versions.Reverse([]string{"v201", "v178", "v120", "v200"})

			Expect(err).NotTo(HaveOccurred())
			Expect(versions).To(Equal([]string{"v200", "v120", "v178", "v201"}))
		})
	})

	Describe("SplitIntoVersionAndFingerprint", func() {
		var (
			input string
		)

		BeforeEach(func() {
			input = "some.version#my-fingerprint"
		})

		It("splits without error", func() {
			version, fingerprint, err := versions.SplitIntoVersionAndFingerprint(input)

			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("some.version"))
			Expect(fingerprint).To(Equal("my-fingerprint"))
		})

		Context("when the input does not contain enough delimiters", func() {
			BeforeEach(func() {
				input = "some.version"
			})

			It("returns error", func() {
				_, _, err := versions.SplitIntoVersionAndFingerprint(input)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the input contains too many delimiters", func() {
			BeforeEach(func() {
				input = "some.version#fingerprint-1#-fingerprint-2"
			})

			It("returns error", func() {
				_, _, err := versions.SplitIntoVersionAndFingerprint(input)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("CombineVersionAndFingerprint", func() {
		var (
			version     string
			fingerprint string
		)

		BeforeEach(func() {
			version = "some.version"
			fingerprint = "my-fingerprint"
		})

		It("combines without error", func() {
			versionWithFingerprint, err := versions.CombineVersionAndFingerprint(version, fingerprint)

			Expect(err).NotTo(HaveOccurred())
			Expect(versionWithFingerprint).To(Equal("some.version#my-fingerprint"))
		})

		Context("when the fingerprint is empty", func() {
			BeforeEach(func() {
				fingerprint = ""
			})

			It("does not include the #", func() {
				versionWithFingerprint, err := versions.CombineVersionAndFingerprint(version, fingerprint)

				Expect(err).NotTo(HaveOccurred())
				Expect(versionWithFingerprint).To(Equal("some.version"))
			})
		})
	})
})