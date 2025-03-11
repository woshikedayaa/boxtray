package gui

import (
	"fmt"
	"fyne.io/systray"
)

type RadioGroup struct {
	items     []*systray.MenuItem
	selected  int
	onChanged func(int)
}

func NewRadioGroup(onChanged func(int)) *RadioGroup {
	return &RadioGroup{
		items:     make([]*systray.MenuItem, 0),
		selected:  -1,
		onChanged: onChanged,
	}
}

func (rg *RadioGroup) AddItem(title string) *systray.MenuItem {
	item := systray.AddMenuItem(title, fmt.Sprintf("Select %s", title))
	rg.items = append(rg.items, item)

	if len(rg.items) == 1 {
		rg.Select(0)
	}

	go func(item *systray.MenuItem, index int) {
		for range item.ClickedCh {
			rg.Select(index)
		}
	}(item, len(rg.items)-1)

	return item
}

func (rg *RadioGroup) Select(index int) {
	if index < 0 || index >= len(rg.items) {
		return
	}

	if rg.selected >= 0 && rg.selected < len(rg.items) {
		rg.items[rg.selected].Uncheck()
	}

	rg.items[index].Check()
	rg.selected = index

	if rg.onChanged != nil {
		rg.onChanged(index)
	}
}

func (rg *RadioGroup) GetSelected() int {
	return rg.selected
}

func (rg *RadioGroup) GetSelectedItem() *systray.MenuItem {
	if rg.selected >= 0 && rg.selected < len(rg.items) {
		return rg.items[rg.selected]
	}
	return nil
}
