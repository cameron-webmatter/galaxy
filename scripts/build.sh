#!/bin/sh
VERSION=$(cat VERSION)
go install -ldflags "-X github.com/cameron-webmatter/galaxy/pkg/cli.Version=$VERSION" ./cmd/galaxy
