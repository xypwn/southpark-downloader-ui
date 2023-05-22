# southpark-downloader-ui
[EARLY RELEASE] Fully self-contained South Park downloader GUI, written in Go for Linux, Windows, MacOS &amp; Android.

The code is currently still quite messy, and will definitely need some more work. If you have any request or criticism in particular, feel free to open an issue.

![Preview image](/preview.png)

## Running the binary
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
