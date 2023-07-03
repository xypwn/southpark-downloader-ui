package gui

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/xypwn/southpark-downloader-ui/internal/logic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Preferences struct {
	widget.BaseWidget
	secDownloads *fyne.Container
	obj          fyne.CanvasObject
}

func NewPreferences(ctx context.Context, cfgStor *logic.StorageItem[*logic.Config], onError func(error), window fyne.Window) *Preferences {
	res := &Preferences{
		secDownloads: container.NewVBox(),
	}
	res.ExtendBaseWidget(res)

	// GUI client
	cfg := cfgStor.NewClient()

	// Output Path
	{
		label := widget.NewLabel("Output Path:")

		var entry *widget.Entry
		{
			entry = widget.NewEntry()
			cfg.Examine(func(c *logic.Config) {
				entry.Text = c.DownloadPath
			})
			pathValidator := func(s string) error {
				uri := storage.NewFileURI(s)
				canList, err := storage.CanList(uri)
				if err != nil {
					return err
				}
				if !canList {
					return errors.New("cannot list URI")
				}
				return nil
			}
			entry.OnChanged = func(s string) {
				if pathValidator(s) == nil {
					cfg.Change(func(c *logic.Config) *logic.Config {
						c.DownloadPath = s
						return c
					})
				}
			}
			entry.Validator = pathValidator
			cfg.AddListener(func(c *logic.Config) {
				entry.SetText(c.DownloadPath)
			})
		}

		var button *widget.Button
		{
			button = widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
				fo := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
					if err != nil {
						onError(err)
						return
					}
					if uri == nil {
						return
					}
					path := uri.Path()
					entry.SetText(path)
				}, window)
				var dlPath string
				cfg.Examine(func(c *logic.Config) {
					dlPath = c.DownloadPath
				})
				uri := storage.NewFileURI(dlPath)
				list, err := storage.ListerForURI(uri)
				if err == nil {
					fo.SetLocation(list)
				}
				fo.Show()
			})
		}

		res.secDownloads.Add(
			container.NewBorder(
				nil,
				nil,
				label,
				button,
				entry,
			),
		)
	}

	// Concurrent Downloads
	{
		min := 1
		max := 8
		label := widget.NewLabel("Concurrent Downloads:")
		opts := make([]string, max-min+1)
		for i := 0; i <= max-min; i++ {
			opts[i] = fmt.Sprint(i + min)
		}
		sel := widget.NewSelect(opts, nil)
		cfg.Examine(func(c *logic.Config) {
			sel.SetSelected(fmt.Sprint(c.ConcurrentDownloads))
		})
		sel.OnChanged = func(s string) {
			v, _ := strconv.Atoi(s)
			cfg.Change(func(c *logic.Config) *logic.Config {
				c.ConcurrentDownloads = v
				return c
			})
		}
		cfg.AddListener(func(c *logic.Config) {
			sel.SetSelected(strconv.Itoa(c.ConcurrentDownloads))
		})
		res.secDownloads.Add(
			container.NewBorder(
				nil,
				nil,
				label,
				nil,
				sel,
			),
		)
	}

	// Maximum Quality
	{
		label := widget.NewLabel("Maximum Quality:")
		defaultQualities := logic.DefaultQualities()
		opts := make([]string, len(defaultQualities))
		for i, v := range defaultQualities {
			opts[i] = v.String()
		}
		sel := widget.NewSelect(opts, nil)
		cfg.Examine(func(c *logic.Config) {
			sel.SetSelected(c.MaximumQuality.String())
		})
		sel.OnChanged = func(s string) {
			for _, quality := range defaultQualities {
				if s == quality.String() {
					cfg.Change(func(c *logic.Config) *logic.Config {
						c.MaximumQuality = quality
						return c
					})
					break
				}
			}
		}
		cfg.AddListener(func(c *logic.Config) {
			sel.SetSelected(c.MaximumQuality.String())
		})
		res.secDownloads.Add(
			container.NewBorder(
				nil,
				nil,
				label,
				nil,
				sel,
			),
		)
	}

	sections := widget.NewAccordion(
		widget.NewAccordionItem("Downloads", res.secDownloads),
	)
	sections.MultiOpen = true
	sections.OpenAll()

	res.obj = container.NewVScroll(sections)

	return res
}

func (p *Preferences) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.obj)
}
