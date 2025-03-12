package boxtray

import (
	"fmt"
	qt "github.com/mappu/miqt/qt6"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/config"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"os"
	"time"
)

func Main(cfg config.Config) {
	err := log.Init(cfg.Log)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "level: ", err.Error())
		return
	}
	logger := log.Get("init")
	client, err := capi.NewClient(cfg.Api.Endpoint(), &capi.ClientConfig{
		Timeout: 10 * time.Second,
		Secret:  cfg.Api.Secret,
	})
	logger.Info("Set endpoint", slog.String("endpoint", cfg.Api.Endpoint()))
	if err != nil {
		logger.Error("Create Client", slog.String("error", err.Error()))
		return
	}
	_ = qt.NewQApplication(os.Args)

	tray1 := qt.NewQSystemTrayIcon2(qt.QApplication_Style().StandardIcon(qt.QStyle__SP_ComputerIcon, nil, nil))
	tray1.OnActivated(func(reason qt.QSystemTrayIcon__ActivationReason) {
		if reason != qt.QSystemTrayIcon__Trigger {
			return
		}
		fmt.Println("Main Click")
	})
	menu := qt.NewQMenu2()
	actionOpen := qt.NewQAction2("Click me")
	actionOpen.SetCheckable(true)
	actionOpen.OnTriggered(func() {
		actionOpen.SetChecked(false)
		fmt.Println("Open Clicked")
	})

	subA := qt.NewQAction2("Sub Action")
	subA.SetCheckable(true)
	subA.OnTriggered(func() {
		fmt.Println("sub check")
		// subA.SetChecked(!subA.IsChecked())
		subA.SetChecked(false)
	})
	subMenu := qt.NewQMenu3("Sub Menu")
	subMenu.AddAction(subA)
	menu.AddAction(actionOpen)
	menu.AddMenu(subMenu)
	actionGroup := qt.NewQActionGroup(tray1.QObject)
	actionGroup.SetExclusive(true)
	btn1 := qt.NewQAction2("a1")
	btn1.SetCheckable(true)
	btn2 := qt.NewQAction2("a2")
	btn2.SetCheckable(true)
	btn3 := qt.NewQAction2("a3")
	btn3.SetCheckable(true)
	actionGroup.AddAction(btn1)
	actionGroup.AddAction(btn2)
	actionGroup.AddAction(btn3)

	menu.AddAction(btn1)
	menu.AddAction(btn2)
	menu.AddAction(btn3)

	tray1.SetContextMenu(menu)
	tray1.Show()
	qt.QApplication_Exec()
}
