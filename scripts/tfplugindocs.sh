#!/usr/bin/env bash

go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs || exit 1

grep -E '^subcategory: ""$' docs/data-sources/* && echo "missing subcategory" && exit 1

grep -E '^subcategory: ""$' docs/resources//* && echo "missing subcategory" && exit 1

git update-index --refresh || exit 1

git diff-index --quiet HEAD -- || exit 1

