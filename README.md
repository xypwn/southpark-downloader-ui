# southpark-downloader-ui
Fully self-contained South Park downloader GUI, written in Go for Linux, Windows &amp; MacOS.

If you have any request or criticism in particular, feel free to open an issue (English is preferred, but German is also fine).

![Preview image](/preview.png)

## Running the app
### Download
- [Windows (64-bit)](https://github.com/xypwn/southpark-downloader-ui/releases/latest/download/southpark-downloader-ui-windows-amd64.exe)
- [Linux (64-bit)](https://github.com/xypwn/southpark-downloader-ui/releases/latest/download/southpark-downloader-ui-linux-amd64.tar.xz)
- [MacOS (Intel)](https://github.com/xypwn/southpark-downloader-ui/releases/latest/download/southpark-downloader-ui-macos-amd64.tar.gz)
- [MacOS (Apple Silicon)](https://github.com/xypwn/southpark-downloader-ui/releases/latest/download/southpark-downloader-ui-macos-arm64.tar.gz)

### Windows
Open the .exe file.

If it complains about the app being unknown to Microsoft, you can click "More info" -> "Run anyway". The warning should no longer appear after that.

### Linux
Extract the archive.

Open a terminal in the extracted directory that contains the `Makefile`.

Run `sudo make install` (system-wide) or `make user-install` (local).

#### NixOS:

Run `nix run github:xypwn/southpark-downloader-ui` to run the application without downloading it manually.

### MacOS
Extract the archive.

Run the .app using **right click** (two fingers), or it will NOT open.

### From source (advanced users)
You need to install [Golang](https://go.dev/dl/) first

`git clone https://github.com/xypwn/southpark-downloader-ui && cd southpark-downloader-ui`

`go build ./cmd/southpark-downloader-ui`

If there's no error message, you should now have an executable binary called `southpark-downloader-ui` (with a `.exe` at the end for Windows)

## Roadmap
- [X] Write a custom data binding type using generics (fyne is too restrictive)
  - [X] Use it instead of fyne's bindings
- [X] Write tests
  - [X] `pkg/data`
  - [X] `pkg/taskqueue`
- [X] Extract GUI components into internal package & despaghettify
  - [X] Individual episodes
  - [X] Downloads
  - [X] Season selection
  - [X] Preferences
- [X] Extract downloader and cache logic into internal package & despaghettify
  - [X] Make downloads persistent after closing the app
- [X] Allow directly downloading search results & fix search in general
- [X] Add 'Download All' button to add all episodes of the season to the queue
- [ ] Make Android usable and useful
  - Figure out a way to save files without direct access to SAF
- [ ] Nitpicks
  - [ ] Fix EllipsisLabel text overflow with very large texts
  - [ ] Add word breaking for EllipsisLabel

## Disclaimer
This application is unofficial and is neither endorsed, nor supported by South Park, Paramount, MTV, Comedy Central, or any associated entities. The application interfaces with official sources using their public APIs in a similar way to a web browser when streaming South Park episodes. Please ensure your usage of this application complies with the terms of service of these APIs.

Downloaded files are intended purely for personal use. Redistribution of these files is considered copyright infringement and is against the law. Users are solely responsible for their use of this application and should not engage in illegal activities.

The application is provided "as is", without warranty of any kind, express or implied, as outlined in the LICENSE. I disclaim all liability and responsibility arising from any reliance placed on the application by its users, or by anyone who may be informed of its contents.
