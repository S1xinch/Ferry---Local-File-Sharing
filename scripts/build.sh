#!/usr/bin/env bash
# Builds release files into dist/.
#   scripts/build.sh [windows] [mac] [linux]     (default: all three)
# The Mac app is only universal and signed when built on a Mac.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo dev)}"
VERSION="${VERSION#v}"
LDFLAGS="-s -w -X main.version=$VERSION"
TARGETS="${*:-windows mac linux}"
export CGO_ENABLED=0

rm -rf dist && mkdir -p dist

build() { # os arch output [extra ldflags]
	GOOS=$1 GOARCH=$2 go build -trimpath -ldflags "$LDFLAGS ${4:-}" -o "$3" .
}

for target in $TARGETS; do
	case $target in
	windows)
		command -v go-winres >/dev/null || go install github.com/tc-hib/go-winres@latest
		WINRES="$(command -v go-winres || echo "$(go env GOPATH)/bin/go-winres")"
		# Windows wants a 4 part numeric version
		NUM="$(echo "$VERSION" | grep -oE '^[0-9]+(\.[0-9]+){0,2}' || echo 0.0.0)"
		"$WINRES" simply --icon packaging/icons/icon.png --manifest gui \
			--product-name Ferry --file-description "Ferry" \
			--product-version "$NUM" --file-version "$NUM" \
			--original-filename ferry.exe --copyright "S1xinch" --arch amd64,arm64
		# -H windowsgui: no black console window when double-clicked
		build windows amd64 dist/ferry-windows-x64.exe "-H windowsgui"
		build windows arm64 dist/ferry-windows-arm64.exe "-H windowsgui"
		build windows amd64 dist/ferry-cli-windows-x64.exe
		rm -f rsrc_windows_*.syso
		;;
	mac)
		APP=dist/mac/Ferry.app
		mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
		sed "s/__VERSION__/$VERSION/g" packaging/Info.plist >"$APP/Contents/Info.plist"
		cp packaging/icons/icon.icns "$APP/Contents/Resources/icon.icns"
		build darwin arm64 dist/mac/ferry-arm64
		build darwin amd64 dist/mac/ferry-amd64
		if command -v lipo >/dev/null; then
			# One app that runs natively on both Apple Silicon and Intel Macs
			lipo -create -output "$APP/Contents/MacOS/ferry" dist/mac/ferry-arm64 dist/mac/ferry-amd64
			# Ad-hoc signature: required for Apple Silicon to run it at all
			codesign --force --deep --sign - "$APP"
			(cd dist/mac && ditto -c -k --keepParent Ferry.app ../ferry-mac.zip)
		else
			cp dist/mac/ferry-arm64 "$APP/Contents/MacOS/ferry"
			(cd dist/mac && zip -qry ../ferry-mac-unsigned-arm64.zip Ferry.app)
		fi
		cp dist/mac/ferry-arm64 dist/ferry-cli-mac-arm64
		cp dist/mac/ferry-amd64 dist/ferry-cli-mac-intel
		rm -rf dist/mac
		;;
	linux)
		build linux amd64 dist/ferry-linux-x64
		build linux arm64 dist/ferry-linux-arm64
		;;
	*)
		echo "unknown target $target" >&2
		exit 1
		;;
	esac
done

ls -la dist
