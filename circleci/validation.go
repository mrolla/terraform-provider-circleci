package circleci

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	// https://circleci.com/docs/2.0/env-vars/#injecting-environment-variables-with-the-api
	environmentVariablePrefixRegex = regexp.MustCompile("^[[:alpha:]]")
	environmentVariableCharsRegex  = regexp.MustCompile("^[[:word:]]+$")
)

func validateEnvironmentVariableNameFunc(v interface{}, key string) (warns []string, errs []error) {
	name, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", key)}
	}

	if !environmentVariablePrefixRegex.MatchString(name) {
		errs = append(errs, errors.New("environment variables may only begin with a letter"))
	}

	if !environmentVariableCharsRegex.MatchString(name) {
		errs = append(errs, errors.New("environment variable names may only contain letters (uppercase and lowercase), digits, and underscores"))
	}

	return warns, errs
}
