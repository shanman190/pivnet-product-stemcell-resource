package validator

import (
	"fmt"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
)

// CheckValidator : validates that a check request is valid before processing can continue
type CheckValidator struct {
	input concourse.CheckRequest
}

// NewCheckValidator : Create a new CheckValidator
func NewCheckValidator(input concourse.CheckRequest) *CheckValidator {
	return &CheckValidator{
		input: input,
	}
}

// Validate : validate the check request
func (v CheckValidator) Validate() error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	if v.input.Source.StemcellSlug == "" {
		return fmt.Errorf("%s must be provided", "stemcell_slug")
	}
	return nil
}
