package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
	"github.com/shanman190/pivnet-product-stemcell-resource/validator"
)

var _ = Describe("Check Validator", func() {
	var (
		checkRequest concourse.CheckRequest
		v            *validator.CheckValidator

		apiToken     string
		productSlug  string
		stemcellSlug string
	)

	BeforeEach(func() {
		apiToken = "some-api-token"
		productSlug = "some-productSlug"
		stemcellSlug = "some-stemcellSlug"
	})

	JustBeforeEach(func() {
		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:     apiToken,
				ProductSlug:  productSlug,
				StemcellSlug: stemcellSlug,
			},
		}
		v = validator.NewCheckValidator(checkRequest)
	})

	It("returns without error", func() {
		err := v.Validate()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when neither legacy API token nor UAA refresh token are provided", func() {
		BeforeEach(func() {
			apiToken = ""
		})

		It("returns an error", func() {
			err := v.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("api_token must be provided"))
		})
	})

	Context("when a UAA refresh token or legacy API token is provided", func() {
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
})
