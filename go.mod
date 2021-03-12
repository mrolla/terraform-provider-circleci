module github.com/mrolla/terraform-provider-circleci

go 1.14

require (
	github.com/CircleCI-Public/circleci-cli v0.1.15108
	github.com/google/uuid v1.1.1
	github.com/hashicorp/terraform-plugin-sdk v1.7.0
	github.com/jszwedko/go-circleci v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
)

replace github.com/jszwedko/go-circleci v0.2.0 => github.com/tgermain/go-circleci v0.0.0-20181207123242-bfc5b3445bba
