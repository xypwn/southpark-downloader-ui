package main

import (
	"context"
    "net/url"
    "strconv"

	/*"os"
	"runtime/pprof"*/

	"github.com/xypwn/southpark-downloader-ui/internal/gui"
	"github.com/xypwn/southpark-downloader-ui/internal/logic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	/*f, err := os.Create("profile.prof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()*/

	ctx := context.Background()

	app := app.NewWithID("org.nobrain.southparkdownloaderui")
	window := app.NewWindow("South Park Downloader")

	onError := func(err error) {
		errText := widget.NewLabel("Error: " + err.Error())
		errText.Wrapping = fyne.TextWrapWord
		errText.Alignment = fyne.TextAlignCenter
		copyBtn := widget.NewButtonWithIcon(
			"Copy",
			theme.ContentCopyIcon(),
			func() {
				window.Clipboard().SetContent(err.Error())
			},
		)
		copyBtn.Importance = widget.LowImportance
		issueBtn := widget.NewButtonWithIcon(
			"Automatically file a bug report on GitHub",
			theme.MailComposeIcon(),
			func() {
                issueURL, _ := url.Parse("https://github.com/xypwn/southpark-downloader-ui/issues/new?labels=bug&"+
                    "title="+url.QueryEscape("Internal Error Bug Report [Automatically Created via App]")+"&"+
                    "body="+url.QueryEscape("Version: "+app.Metadata().Version+
                        " build "+strconv.Itoa(app.Metadata().Build)+"\n"+
                        "Error: "+err.Error()),

                )
                app.OpenURL(issueURL)
			},
		)
		issueBtn.Importance = widget.LowImportance
		dialog.ShowCustom("An internal error has occurred", "OK", container.NewVBox(
			errText,
			copyBtn,
            issueBtn,
		), window)
	}

	storage, err := logic.NewStorage(app.Storage().RootURI().Path())
	if err != nil {
		panic(err)
	}
	cfgStor, err := logic.NewStorageItem(storage, "config", logic.NewConfig, func(err error) {
		panic(err)
	})
	if err != nil {
		panic(err)
	}
	cacheStor, err := logic.NewStorageItem(storage, "cache", logic.NewCache, func(err error) {
		panic(err)
	})
	if err != nil {
		panic(err)
	}
	dlInfoStor, err := logic.NewStorageItem(storage, "downloads", logic.NewDownloadsInfo, func(err error) {
		panic(err)
	})
	if err != nil {
		panic(err)
	}

	dls := logic.NewDownloads(cfgStor.NewClient(), onError)

	downloads := gui.NewDownloads(dls)

	logic.ConnectDownloadsToDownloadsInfo(ctx, dls, dlInfoStor, func(err error) {
		panic(err)
	})

	episodesPanel := gui.NewEpisodesPanel(ctx, dls, cacheStor, cfgStor.NewClient(),
		func(title, text string) {
			dialog.ShowInformation(title, text, window)
		},
		onError,
		func(newClipboardContent string) {
			window.Clipboard().SetContent(newClipboardContent)
		},
	)

	appTabs := container.NewAppTabs(
		container.NewTabItemWithIcon(
			"Episodes",
			theme.ListIcon(),
			episodesPanel,
		),
		container.NewTabItemWithIcon(
			"Downloads",
			theme.DownloadIcon(),
			downloads,
		),
		container.NewTabItemWithIcon(
			"Preferences",
			theme.SettingsIcon(),
			gui.NewPreferences(ctx, cfgStor, onError, window),
		),
	)

	window.SetContent(appTabs)

	window.Resize(fyne.NewSize(800, 450))

	window.ShowAndRun()
}
