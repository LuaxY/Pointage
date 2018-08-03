package main

import (
	"Pointage/adp"
	"Pointage/config"
	"Pointage/msgbox"
	syst "Pointage/systray"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/jasonlvhit/gocron"
	"github.com/lxn/win"
	"github.com/sqweek/dialog"
	"gopkg.in/matryer/try.v1"
	"log"
	"os/exec"
	"time"
)

var cfg = config.Get()

func main() {
	go activity()
	go ADP()

	systray.Run(syst.OnReady, syst.OnExit)
}

func ADP() {
	location, err := adp.Preload()

	if err != nil {
		log.Fatal(err)
	}

	name, err := adp.Login(cfg.ADP.Username, cfg.ADP.Password, location)

	if err != nil {
		log.Fatal(err)
	}

	mInfo := systray.AddMenuItem("Utilisateur: "+name, "")
	mInfo.Disable()

	systray.AddSeparator()

	mPointage := systray.AddMenuItem("Pointer", "")
	mRefresh := systray.AddMenuItem("Rafraichir", "")

	systray.AddSeparator()

	for i := 0; i < 4; i++ {
		menusHistory = append(menusHistory, systray.AddMenuItem("Pas encore pointer", ""))
	}

	if cfg.Friday.Shutdown {
		systray.AddSeparator()
		mShutdown := systray.AddMenuItem("Arrêt du PC vendredi à "+cfg.Friday.At, "")
		mShutdown.Disable()
	}

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Fermer", "Fermer")

	schedul()
	refreshHistory()

	go func() {
		for {
			select {
			case <-mPointage.ClickedCh:
				pointer(false)
			case <-mRefresh.ClickedCh:
				refreshHistory()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	go startSchedul()
}

var format = "15:04"

func pointer(auto bool) {
	now, _ := time.Parse(format, time.Now().Format(format))

	hours := []string{"09:00", "12:30", "14:00", "18:00"}

	T, _ := time.Parse(format, "00:00")

	for index, menu := range menusHistory {
		if menu.Checked() {
			continue
		}

		T, _ = time.Parse(format, hours[index])
		break
	}

	hourBefore := T.Add(-1*time.Hour + -1*time.Second)
	hourAfter := T.Add(1*time.Hour + 1*time.Second)

	if now.After(hourBefore) && now.Before(hourAfter) {
		if !auto || time.Since(lastActivity).Minutes() <= 5 {
			ok := dialog.Message("Etes vous sûr de vouloir pointer pour %s ?", now.Format(format)).Title("Pointage").YesNo()

			if !ok {
				return
			}
		}

		err := try.Do(func(attempt int) (bool, error) {
			refreshHistory()

			if err := adp.Pointage(); err != nil {
				return attempt < 3, err
			}

			log.Print("pointage")

			refreshHistory()

			return false, nil
		})

		if err != nil {
			log.Print(err)
			return
		}
	} else if !auto {
		info := fmt.Sprintf("Le prochain pointage doit être fait entre %s et %s.\nIl est actuellment %s.", T.Add(-1*time.Hour).Format(format), T.Add(1*time.Hour).Format(format), now.Format(format))
		msgbox.Display("Trop tôt !", "Il est trop tôt pour pointer >:(\n\n"+info, msgbox.MB_OK|msgbox.MB_ICONWARNING)
	}
}

var menusHistory []*systray.MenuItem

func refreshHistory() {
	for _, menu := range menusHistory {
		menu.Disable()
		menu.Uncheck()
	}

	var err error
	var hours []string

	err = try.Do(func(attempt int) (bool, error) {
		hours, err = adp.History()

		if err != nil {
			return attempt < 3, err
		}

		return false, nil
	})

	if err != nil {
		log.Print(err)
		return
	}

	for index, hour := range hours {
		menusHistory[index].SetTitle(hour)
		menusHistory[index].Check()
	}
}

func schedul() {
	// Reset every new day
	gocron.Every(1).Day().At("00:01").Do(reset)

	if cfg.Friday.Shutdown {
		gocron.Every(1).Friday().At(cfg.Friday.At).Do(shutdown)
	}

	for index, hour := range cfg.ADP.Auto {
		if hour == "" || menusHistory[index].Checked() {
			continue
		}

		menusHistory[index].SetTitle("Auto pointage à " + hour)
		gocron.Every(1).Day().At(hour).Do(pointer, true)
	}
}

func startSchedul() {
	<-gocron.Start()
}

func reset() {
	fmt.Println("Refresh")

	gocron.Remove(reset)
	gocron.Remove(pointer)
	gocron.Clear()

	schedul()
	refreshHistory()
}

var lastActivity time.Time

func activity() {
	var current, previous win.POINT

	for {
		win.GetCursorPos(&current)

		if current.X == previous.X && current.Y == previous.Y {
		} else {
			lastActivity = time.Now()
		}

		previous = current
		time.Sleep(10 * time.Second)
	}
}

func shutdown() {
	if err := exec.Command("cmd", "/C", "shutdown", "/s", "/f", "/t", "60").Run(); err != nil {
		log.Print("failed to initiate shutdown:", err)
	}
}
