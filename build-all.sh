#!/usr/bin/env sh

NAME="$(go list -m)"
SRCDIR="./cmd/southpark-downloader-ui"
BUILDDIR="build"

[ "$1" = "clean" ] && echo "Cleaning $BUILDDIR" && rm -rf "$BUILDDIR" && exit 0

build() {
	# [linux|dragonfly|freebsd|netbsd|openbsd|plan9|solaris|darwin|windows]"
	OS="$1"
	# [arm|arm64|ppc64|ppc64le|mips64|386|amd64]
	ARCH="$2"
	# compiler executable, for example cc or x86_64-w64-mingw32-cc
	CC="$3"

	echo "Building for $OS on $ARCH using $CC"

	[ "$OS" = "windows" ] && EXT=".exe"

	mkdir -p "$BUILDDIR"

	env GOOS="$OS" GOARCH="$ARCH" CC="$CC" CGO_ENABLED=1 go build -ldflags "-s -w" -o "$BUILDDIR/$NAME-$OS-$ARCH$EXT" "$SRCDIR"
}

build_android() {
	echo "Building for Android"

	mkdir -p "$BUILDDIR"

	cd "$SRCDIR"
	./package_android.sh
	cd -

	mv "$SRCDIR/Southpark_Downloader.apk" "$BUILDDIR/$NAME.apk"
}

#build linux 386 cc &
build linux amd64 cc &
#build linux arm cc &
#build linux arm64 cc &
#build darwin amd64 &
#build windows 386 winegcc &
build windows amd64 x86_64-w64-mingw32-cc &
build_android &

wait
