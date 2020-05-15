---
layout: "circleci"
page_title: "CircleCI: circleci_context_environment_variable"
sidebar_current: "docs-resource-circleci-context-environment-variable"
description: |-
  Manages a CircleCI context environment variable.
---

# circleci_context_environment_variable

A CircleCI context is a named collection of environment variables that can be referenced in the configuration for workflows.
Each environment variable is represented by a separate Terraform resource.

## Example Usage

Basic usage:

```hcl
resource "circleci_context" "build" {
  name  = "build"
}

resource "circleci_context_environment_variable" "build" {
  variable   = "TOKEN"
  value      = "secret"
  context_id = circleci_context.build.id
}
```

With `for_each`:

```hcl
resource "circleci_context" "build" {
  name  = "build"
}

resource "circleci_context_environment_variable" "build" {
  for_each = {
    TOKEN_A = "secret"
    TOKEN_B = "secret"
  }

  variable   = each.key
  value      = each.value
  context_id = circleci_context.build.id
}
```

## Argument Reference

The following arguments are supported:

* `variable` - (Required) Name of the environment variable.
* `value` - (Required) The value of the environment variable. A hash of this value will be stored in state in order to detect changes, but the plain text value will not be stored.
* `context_id` - (Required) The context that the environment variable will be added to.
* `organization` - (Optional) Organization where the context is defined.

## Import

Context environment variables can be imported as `$organization/$context/$variable`, where "context" can be either a context name or ID. 
Additionally, you must specify an existing value as `CIRCLECI_ENV_VALUE` so that Terraform can detect whether it needs to update the variable.

For example:

```shell
CIRCLECI_ENV_VALUE=secret terraform import circleci_context_environment_variable.token hashicorp/build/TOKEN
```
