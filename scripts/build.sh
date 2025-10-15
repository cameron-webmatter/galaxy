#!/bin/sh
VERSION=$(cat VERSION)
go install -ldflags "-X github.com/galaxy/galaxy/pkg/cli.Version=$VERSION" ./cmd/galaxy
