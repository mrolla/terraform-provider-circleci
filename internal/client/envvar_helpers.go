package client

import (
	"regexp"
)

const _envVarNameValidRE = "^[[:alpha:]]+[[:word:]]*$" // source https://circleci.com/docs/2.0/env-vars/#injecting-environment-variables-with-the-api

var _envVarNameValidCompiledRE *regexp.Regexp

func init() {
	_envVarNameValidCompiledRE = regexp.MustCompile(_envVarNameValidRE)
}

// ValidateEnvVarName check an environment variable name is valid according to https://circleci.com/docs/2.0/env-vars/#injecting-environment-variables-with-the-api
func ValidateEnvVarName(envVarName string) bool {
	return _envVarNameValidCompiledRE.Match([]byte(envVarName))
}
