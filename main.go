package main

import (
	"github.com/JustinTimperio/onionsoup/data"
	"github.com/JustinTimperio/onionsoup/menus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const preferenceCurrentMenu = "Home"

var topWindow fyne.Window

func main() {

	a := app.NewWithID("onionsoup")
	a.SetIcon(data.Logo)
	w := a.NewWindow("OnionSoup")
	content := container.NewStack()

	setMenu := func(t menus.Menu) {
		if fyne.CurrentDevice().IsMobile() {
			child := a.NewWindow(t.Title)
			topWindow = child
			child.SetContent(t.View(child))
			child.Show()
			child.SetOnClosed(func() {
				topWindow = w
			})
			return
		}

		content.Objects = []fyne.CanvasObject{t.View(w)}
		content.Refresh()
	}

	menu := container.NewStack(content)
	split := container.NewHSplit(makeNav(setMenu), menu)
	split.Offset = 0.10
	w.SetContent(split)
	w.Resize(fyne.Size{Width: 1024, Height: 768})

	w.ShowAndRun()
}

func makeNav(setWindow func(m menus.Menu)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return menus.MenuOrder
		},
		IsBranch: func(uid string) bool {
			if uid == "" {
				return true
			} else {
				return false
			}
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := menus.Menus[uid]
			if !ok {
				fyne.LogError("Missing menu panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
			obj.(*widget.Label).TextStyle = fyne.TextStyle{}
		},
		OnSelected: func(uid string) {
			if t, ok := menus.Menus[uid]; ok {
				a.Preferences().SetString(preferenceCurrentMenu, uid)
				setWindow(t)
			}
		},
	}

	setWindow(menus.Menus["Home"])
	return container.NewBorder(nil, nil, nil, nil, tree)
}
