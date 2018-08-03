package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"Pointage/adp"
	"Pointage/config"
	"Pointage/msgbox"
	syst "Pointage/systray"

	"github.com/getlantern/systray"
	"github.com/jasonlvhit/gocron"
	"github.com/lxn/win"
	"github.com/sqweek/dialog"
	"gopkg.in/matryer/try.v1"
)

var cfg = config.Get()
var format = "15:04"
var lastActivity time.Time
var lastLogin time.Time
var mEstimated *systray.MenuItem
var menusHistory []*systray.MenuItem

func main() {
	go activity()
	go ADP()

	systray.Run(syst.OnReady, syst.OnExit)
}

func login() (string, error) {
	var name string
	var err error

	err = try.Do(func(attempt int) (bool, error) {
		location, err := adp.Preload()

		if err != nil {
			return attempt < 3, err
		}

		name, err = adp.Login(cfg.ADP.Username, cfg.ADP.Password, location)

		if err != nil {
			return attempt < 3, err
		}

		return false, nil
	})

	if err != nil {
		return "", err
	}

	lastLogin = time.Now()

	return name, nil
}

func ADP() {
	name, err := login()

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

	systray.AddSeparator()

	mEstimated = systray.AddMenuItem("Temps de travail estimé: inconnu", "")
	mEstimated.Disable()

	if cfg.Friday.Shutdown {
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

func pointer(auto bool) {
	if time.Since(lastLogin).Minutes() >= 10 {
		_, err := login()

		if err != nil {
			log.Print(err)
			return
		}
	}

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

func refreshHistory() {
	if time.Since(lastLogin).Minutes() >= 10 {
		_, err := login()

		if err != nil {
			log.Print(err)
			return
		}
	}

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

	matin, _ := time.Parse(format, cfg.ADP.Auto[0])
	midi, _ := time.Parse(format, cfg.ADP.Auto[1])
	aprem, _ := time.Parse(format, cfg.ADP.Auto[2])
	soir, _ := time.Parse(format, cfg.ADP.Auto[3])

	if len(hours) >= 1 {
		matin, _ = time.Parse(format, hours[0][13:])
	}

	if len(hours) >= 2 {
		midi, _ = time.Parse(format, hours[1][13:])
	}

	if len(hours) >= 3 {
		aprem, _ = time.Parse(format, hours[2][13:])
	}

	if len(hours) >= 4 {
		soir, _ = time.Parse(format, hours[3][13:])
	}

	h1 := midi.Sub(matin)
	h2 := soir.Sub(aprem)

	estimation := h1 + h2 - (30 * time.Minute)

	mEstimated.SetTitle("Temps de travail estimé: " + estimation.String())
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
