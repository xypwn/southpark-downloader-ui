#!/usr/bin/env sh
if ! uname -s | grep -i "linux" > /dev/null; then
	echo "This script is supposed to be run on Linux!"
	exit 1
fi

export ANDROID_NDK_HOME=/opt/android-ndk
export MACOS_SDK="$PWD/MacOSX10.15.sdk"

APPID="org.nobrain.southparkdownloaderui"
NAME="southpark-downloader-ui"
ICON="./cmd/southpark-downloader-ui/Icon.png"
SRC="./cmd/southpark-downloader-ui"

cross_compile() {
	fyne-cross "$@" \
		-app-id "$APPID" \
		-name "$NAME" \
		-icon "$ICON" \
		"$SRC"
}

compile() {
	cd "$SRC"
	fyne package "$@" \
		--release \
		--appID "$APPID" \
		--name "$NAME" \
		--icon "../../$ICON"
	cd -
}

compile || exit 1
compile --target android || exit 1
cross_compile windows || exit 1
cross_compile darwin -arch amd64 -macosx-sdk-path "$MACOS_SDK" || exit 1
cross_compile darwin -arch arm64 -macosx-sdk-path "$MACOS_SDK" || exit 1

mkdir -p build
mv "$SRC/southpark-downloader-ui.tar.xz" "build/$NAME-linux-install.tar.xz"
mv "$SRC/southpark_downloader_ui.apk" "build/$(echo "$NAME" | tr - _)_android.apk"
mv "fyne-cross/bin/windows-amd64/southpark-downloader-ui.exe" "build/$NAME-windows.exe"
tar -C "fyne-cross/dist/darwin-amd64/" -czf "build/$NAME-macos-x64.tar.gz" "southpark-downloader-ui.app" 
tar -C "fyne-cross/dist/darwin-arm64/" -czf "build/$NAME-macos-arm64.tar.gz" "southpark-downloader-ui-apple-silicon.app"

TMPDIR="$(mktemp -d)"
tar -xf "build/$NAME-linux-install.tar.xz" -C "$TMPDIR" "usr/local/bin/southpark-downloader-ui"
mv "$TMPDIR/usr/local/bin/southpark-downloader-ui" "build/$NAME-linux-standalone"
rm -rf "$TMPDIR"
