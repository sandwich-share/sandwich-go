#!/bin/sh

VERSION=$(git rev-parse --short HEAD)

echo "package main\n\nconst VERSION = \"$VERSION\"\n" > VERSION.go

go build

