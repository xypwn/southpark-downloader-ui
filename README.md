# southpark-downloader-ui
Fully self-contained South Park downloader GUI, written in Go for Linux, Windows, MacOS &amp; Android.

If you have any request or criticism in particular, feel free to open an issue (English is preferred, but German is also fine).

![Preview image](/preview.png)

## Running the app
### Download
[From GitHub Releases](https://github.com/xypwn/southpark-downloader-ui/releases/latest)

### Windows
Just double-click the .exe :)

### Linux (standalone)
#### Graphical file manager (Gnome Nautilus, PCManFM etc.)
Right-click the executable. Under properties, toggle the 'Executable' switch on, **OR** under 'Permissions' -> 'Execute', select 'Everyone'.

Now you can double-click and run :)

#### Terminal
Run `chmod +x <binary file name>`.

Now you can run it with `./<binary file name>`, or using the graphical method.

### Linux (install)
Unzip the file.

Open a terminal in the folder of the unzipped file. Make sure you have `make` installed.

Run `make user-install` for a local install, or `sudo make install` for a system-wide install.

### MacOS
Thanks to @KatzeMau for testing

Open a terminal and run `chmod +x <binary file name>`. This makes it so you can run the file.

Apple doesn't like it if you run programs that aren't certified by Apple.

To run the program, you have to right-click it, then press open (NOT double-click!). It will show a warning and ask you if you really want to run the program. Press confirm.

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
- [ ] Add 'Download All' button to add all episodes of the season to the queue
- [ ] Make Android usable and useful
  - Figure out a way to save files without direct access to SAF
- [ ] Nitpicks
  - [ ] Fix EllipsisLabel text overflow with very large texts
  - [ ] Add word breaking for EllipsisLabel

## Disclaimer
This application is unofficial and is neither endorsed, nor supported by South Park, Paramount, MTV, Comedy Central, or any associated entities. The application interfaces with official sources using their public APIs in a similar way to a web browser when streaming South Park episodes. Please ensure your usage of this application complies with the terms of service of these APIs.

Downloaded files are intended purely for personal use. Redistribution of these files is considered copyright infringement and is against the law. Users are solely responsible for their use of this application and should not engage in illegal activities.

The application is provided "as is", without warranty of any kind, express or implied, as outlined in the LICENSE. I disclaim all liability and responsibility arising from any reliance placed on the application by its users, or by anyone who may be informed of its contents.
