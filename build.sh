#!/bin/sh

VERSION=0.9

LOCAL_BUILD="true"

if [ $# -gt 0 ]; then
	if [ $1 = "release" ]
	then
		LOCAL_BUILD=-1
	fi
fi

if [ $LOCAL_BUILD = "true" ]
then
	VERSION=$VERSION:$(git rev-parse --short HEAD)
	echo "Building local version"
else
	echo "Building release version"
fi

echo "package main\n\nconst VERSION = \"$VERSION\"\n" > VERSION.go

go build

rm VERSION.go

