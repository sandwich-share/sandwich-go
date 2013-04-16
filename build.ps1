$VERSION = git rev-parse --short HEAD

echo "package main`n`nconst VERSION = `"$VERSION`"`n" | Out-File -Encoding "ascii" VERSION.go

go build