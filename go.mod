module github.com/mrolla/terraform-provider-circleci

go 1.14

require (
	github.com/CircleCI-Public/circleci-cli v0.1.7645
	github.com/hashicorp/terraform-plugin-sdk v1.7.0
	github.com/jszwedko/go-circleci v0.2.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
)

replace github.com/jszwedko/go-circleci v0.2.0 => github.com/tgermain/go-circleci v0.0.0-20181207123242-bfc5b3445bba
