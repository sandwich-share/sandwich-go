$GITVER = git rev-parse --short HEAD

$VERSION = "0.9"

$RELEASE = 0
if ($args.count -gt 0) {
	if ($args[0].CompareTo("release") -eq 0) {
		$RELEASE = 1
	}
}

if ($RELEASE -eq 0) {
	$VERSION += ":"+$GITVER
}

echo "package main`n`nconst VERSION = `"$VERSION`"`n" | Out-File -Encoding "ascii" VERSION.go

go build