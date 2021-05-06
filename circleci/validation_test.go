package circleci

import "testing"

func TestValidateEnvironmentVariableName(t *testing.T) {
	cases := []struct {
		Name  string
		Error bool
	}{
		{
			Name: "valid",
		},
		{
			Name: "VALID",
		},
		{
			Name: "VALID_UNDERSCORE_",
		},
		{
			Name: "VALID_DIGIT_1",
		},
		{
			Name:  "invalid-dashed",
			Error: true,
		},
		{
			Name:  "1_invalid_leading_digit",
			Error: true,
		},
		{
			Name:  "_invalid_leading_underscore",
			Error: true,
		},
	}

	for _, tc := range cases {
		var name interface{} = tc.Name
		_, errors := validateEnvironmentVariableNameFunc(name, "")

		if tc.Error != (len(errors) != 0) {
			if tc.Error {
				t.Fatalf("expected error, got none (%s)", tc.Name)
			} else {
				t.Fatalf("unexpected error(s): %s (%s)", errors, tc.Name)
			}
		}
	}
}
