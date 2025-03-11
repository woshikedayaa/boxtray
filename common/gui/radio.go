package gui

import (
	"fyne.io/systray"
	"sync"
)

type RadioGroup struct {
	root      *systray.MenuItem
	items     []*systray.MenuItem
	selected  int
	onChanged func(int)
	mu        sync.Mutex
}

func NewRadioGroup(title, tooltip string, onChanged func(int)) *RadioGroup {
	return &RadioGroup{
		root:      systray.AddMenuItem(title, tooltip),
		items:     make([]*systray.MenuItem, 0),
		selected:  -1,
		onChanged: onChanged,
		mu:        sync.Mutex{},
	}
}

func (rg *RadioGroup) AddItem(title string, tooltip string) *systray.MenuItem {
	item := rg.root.AddSubMenuItemCheckbox(title, tooltip, true)
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.items = append(rg.items, item)

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
	rg.mu.Lock()
	defer rg.mu.Unlock()
	if rg.selected == index {
		return
	}
	for _, v := range rg.items {
		v.Uncheck()
	}

	rg.items[index].Check()

	rg.selected = index

	if rg.onChanged != nil {
		rg.onChanged(index)
	}
}

func (rg *RadioGroup) GetSelected() int {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	return rg.selected
}

func (rg *RadioGroup) GetSelectedItem() *systray.MenuItem {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	if rg.selected >= 0 && rg.selected < len(rg.items) {
		return rg.items[rg.selected]
	}
	return nil
}

func (rg *RadioGroup) ChangeHook(f func(i int)) {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	rg.onChanged = f
}

func (rg *RadioGroup) Remove() {
	for _, v := range rg.items {
		v.Remove()
	}
	rg.root.Remove()
}
