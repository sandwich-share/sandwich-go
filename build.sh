#!/bin/sh

VERSION=$(git describe --all --always HEAD^)

echo "package main\n\nconst VERSION = \"$VERSION\"\n" > VERSION.go

go build

