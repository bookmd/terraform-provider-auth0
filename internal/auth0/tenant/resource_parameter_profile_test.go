package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceParameterProfileSchema(t *testing.T) {
	validate := NewResource().Schema["resource_parameter_profile"].ValidateFunc

	for _, profile := range []string{"audience", "compatibility"} {
		_, errs := validate(profile, "resource_parameter_profile")
		assert.Empty(t, errs, profile)
	}

	_, errs := validate("resource", "resource_parameter_profile")
	assert.NotEmpty(t, errs)
}
