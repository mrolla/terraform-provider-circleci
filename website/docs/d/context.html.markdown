---
layout: "circleci"
page_title: "CircleCI: circleci_context"
sidebar_current: "docs-datasource-circleci-context"
description: |-
  Get information about a CircleCI context.
---

# Data Source: circleci_context

Use this data source to get information about a CircleCI context.

~> **NOTE:** This resource uses the CircleCI GraphQL API, which is not officially supported.

## Example Usage

```hcl
data "circleci_context" "build" {
  name = "build"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the context.
* `organization` - (Optional) Organization where the context is defined.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the context.
