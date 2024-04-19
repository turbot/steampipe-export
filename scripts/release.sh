#!/bin/bash

export VERSION=$(go list -m -json github.com/turbot/steampipe-plugin-chaosdynamic | jq --raw-output '.Version | sub("^v"; "")')
echo "VERSION set to $VERSION"
goreleaser release --snapshot --rm-dist --skip=publish --skip=validate