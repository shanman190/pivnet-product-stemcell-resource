package validator

import (
	"fmt"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
)

// InValidator : validates that a in request is valid before processing can continue
type InValidator struct {
	input concourse.InRequest
}

// NewInValidator : Create a new InValidator
func NewInValidator(input concourse.InRequest) *InValidator {
	return &InValidator{
		input: input,
	}
}

// Validate : validate the in request
func (v InValidator) Validate() error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	if v.input.Source.StemcellSlug == "" {
		return fmt.Errorf("%s must be provided", "stemcell_slug")
	}

	if v.input.Version.ProductVersion == "" {
		return fmt.Errorf("%s must be provided", "product_version")
	}

	if v.input.Version.StemcellVersion == "" {
		return fmt.Errorf("%s must be provided", "stemcell_version")
	}

	return nil
}
