package gui

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"unicode"

	"github.com/xypwn/southpark-downloader-ui/internal/logic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SelectableEntry struct {
	widget.Entry
	OnSelected   func()
	OnUnselected func()
}

func NewSelectableEntry() *SelectableEntry {
	res := &SelectableEntry{
		OnSelected:   func() {},
		OnUnselected: func() {},
	}
	res.ExtendBaseWidget(res)
	return res
}

func (e *SelectableEntry) FocusGained() {
	e.OnSelected()
	e.Entry.FocusGained()
}

func (e *SelectableEntry) FocusLost() {
	e.OnUnselected()
	e.Entry.FocusLost()
}

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

	// File Name
	{
		label := widget.NewLabel("File Name:")

		defaultValue := logic.NewConfig().OutputFilePattern

		var help *fyne.Container
		{
			help = container.NewVBox()
			help.Hide()
		}

		var errLabel *widget.Label
		var errCnt *fyne.Container
		{
			errLabel = widget.NewLabel("")
			errLabel.Wrapping = fyne.TextWrapBreak
			errCnt = container.NewBorder(
				nil,
				nil,
				widget.NewIcon(theme.NewErrorThemedResource(theme.ErrorIcon())),
				nil,
				errLabel,
			)
			errCnt.Hide()
			help.Add(errCnt)
		}

		var example *widget.Label
		{
			example = widget.NewLabel("")
			example.Wrapping = fyne.TextWrapBreak
			help.Add(example)
		}

		var entry *SelectableEntry
		var reset *widget.Button
		{
			reset = widget.NewButtonWithIcon("Reset", theme.ContentUndoIcon(), func() {
				dialog.ShowConfirm("Reset File Name", "Do you really want to reset the file name pattern?", func(ok bool) {
					if !ok {
						return
					}
					entry.SetText(defaultValue)
					help.Hide()
				}, window)
			})
		}
		{
			entry = NewSelectableEntry()
			validator := func(s string) error {
				{
					dollar := false
					for _, c := range s {
						if dollar {
							if !strings.Contains("SELTQ", string(c)) {
								return errors.New("unknown parameter: $" + string(c))
							}
						}
						dollar = c == '$'
					}
					if dollar {
						return errors.New("illegal $ symbol at the end")
					}
				}
				for _, v := range []string{"$S", "$E", "$L", "$T", "$Q"} {
					if strings.Count(s, v) > 1 {
						return errors.New("more than one instance of " + v)
					}
				}
				hasS := strings.Contains(s, "$S") // Season
				hasE := strings.Contains(s, "$E") // Episode
				hasL := strings.Contains(s, "$L") // Language
				hasT := strings.Contains(s, "$T") // Title
				if !hasL {
					return errors.New("missing $L (language)")
				}
				if !((hasS && hasE) || hasT) {
					return errors.New("missing either both $S (season) and $E (episode), or just $T (title)")
				}
				return nil
			}
			entry.OnChanged = func(s string) {
				if s == defaultValue {
					reset.Disable()
				} else {
					reset.Enable()
				}

				toValidFilename := func(s string) (string, bool) {
					changed := false
					var result strings.Builder
					for i := 0; i < len(s); i++ {
						b := s[i]
						if ('a' <= b && b <= 'z') ||
							('A' <= b && b <= 'Z') ||
							('0' <= b && b <= '9') ||
							b == '_' ||
							b == '$' {
							result.WriteByte(b)
						} else {
							result.WriteByte('_')
							changed = true
						}
					}
					return result.String(), changed
				}

				if newS, changed := toValidFilename(s); changed {
					entry.SetText(newS)
					return
				}

				if validator(s) == nil {
					{
						rep := strings.NewReplacer(
							"$S", "01",
							"$E", "01",
							"$L", "English",
							"$T", "Cartman_Gets_An_Anal_Probe",
							"$Q", "Best",
						)
						example.SetText("Example: " + rep.Replace(s) + ".mp4")
					}
					cfg.Change(func(c *logic.Config) *logic.Config {
						c.OutputFilePattern = s
						return c
					})
				}
			}
			entry.Validator = validator
			var init string
			cfg.Examine(func(c *logic.Config) {
				init = c.OutputFilePattern
				if init == "" {
					init = defaultValue
				}
			})
			entry.SetText(init)
		}
		cfg.AddListener(func(c *logic.Config) {
			entry.SetText(c.OutputFilePattern)
		})

		var info *fyne.Container
		{
			legend := widget.NewRichTextFromMarkdown(`- $S: Season Number
- $E: Episode Number
- $L: Language
- $T: Title
- $Q: Quality`)
			legend.Wrapping = fyne.TextWrapWord

			note := widget.NewLabel("You must include $L. Additionally, either both $S and $E, or just $T is required.")
			note.Wrapping = fyne.TextWrapWord

			info = container.NewVBox(
				legend,
				note,
			)

			help.Add(info)
		}

		entry.OnSelected = func() {
			help.Show()
		}
		entry.OnUnselected = func() {
			if entry.Validate() == nil {
				help.Hide()
			}
		}

		entry.SetOnValidationChanged(func(err error) {
			if err != nil {
				e := []byte(err.Error())
				e[0] = byte(unicode.ToUpper(rune(e[0])))
				errLabel.SetText("Error: " + string(e))
				errCnt.Show()
				example.Hide()
			} else {
				errCnt.Hide()
				example.Show()
			}
		})

		helpIndent := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
		helpIndent.SetMinSize(fyne.NewSize(10, 0))

		res.secDownloads.Add(
			container.NewVBox(
				container.NewBorder(
					nil,
					nil,
					label,
					reset,
					entry,
				),
				container.NewBorder(
					nil,
					nil,
					helpIndent,
					nil,
					help,
				),
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
