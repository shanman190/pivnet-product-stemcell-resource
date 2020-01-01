package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
	"github.com/shanman190/pivnet-product-stemcell-resource/validator"
)

var _ = Describe("In Validator", func() {
	var (
		inRequest      concourse.InRequest
		v              *validator.InValidator

		apiToken        string
		productSlug     string
		stemcellSlug    string
		productVersion  string
		stemcellVersion string
	)

	BeforeEach(func() {
		apiToken = "some-api-token"
		productSlug = "some-productSlug"
		stemcellSlug = "some-stemcellSlug"
		productVersion = "some-product-version"
		stemcellVersion = "some-stemcell-version"
	})

	JustBeforeEach(func() {
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				APIToken:     apiToken,
				ProductSlug:  productSlug,
				StemcellSlug: stemcellSlug,
			},
			Params: concourse.InParams{},
			Version: concourse.Version{
				ProductVersion:  productVersion,
				StemcellVersion: stemcellVersion,
			},
		}

		v = validator.NewInValidator(inRequest)
	})

	It("returns without error", func() {
		err := v.Validate()
		Expect(err).NotTo(HaveOccurred())
	})


	Context("when neither UAA refresh token nor legacy API token are provided", func() {
		BeforeEach(func() {
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("api_token must be provided"))
		})
	})

	Context("when UAA refresh token or legacy API token is provided", func() {
		It("returns without error", func() {
			err := v.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when no product slug is provided", func() {
		BeforeEach(func() {
			productSlug = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(".*product_slug.*provided"))
		})
	})

	Context("when no stemcell slug is provided", func() {
		BeforeEach(func() {
			stemcellSlug = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(".*stemcell_slug.*provided"))
		})
	})

	Context("when no product version is provided", func() {
		BeforeEach(func() {
			productVersion = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(MatchRegexp(".*product_version.*provided"))
		})
	})

	Context("when no stemcell version is provided", func() {
		BeforeEach(func() {
			stemcellVersion = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(MatchRegexp(".*stemcell_version.*provided"))
		})
	})
})
