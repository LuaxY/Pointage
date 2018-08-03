package systray

import (
	"Pointage/icon"
	"github.com/getlantern/systray"
)

func OnReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Pointage")
	systray.SetTooltip("Pointage")
}

func OnExit() {

}
