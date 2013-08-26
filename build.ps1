$GITVER = git rev-parse --short HEAD

$VERSION = "1.0"

$RELEASE = 0
$UPDATE = 0

if ($args.count -gt 0) {
	if ($args[0].CompareTo("release") -eq 0) {
		$RELEASE = 1
	} elseif ($args[0].CompareTo("update") -eq 0) {
		$UPDATE = 1
	}
}

if ($RELEASE -eq 0) {
	echo "Building local version"
	$VERSION += ":"+$GITVER
} else {
	echo "Building release version"
}

echo "package util`n`nconst VERSION = `"$VERSION`"`n" | Out-File -Encoding "ascii" util/VERSION.go

if ($UPDATE -eq 0) {
	go build
} else {
	echo "Updating dependencies"
	go get
}


rm util/VERSION.go
