package adp

import (
	"Pointage/config"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"errors"
	"github.com/PuerkitoBio/goquery"
)

func Preload() (string, error) {
	// No follow redirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest("GET", "https://pointage.adp.com/igested/2_02_01/pointage", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", config.UserAgent)

	dumpRequest(req)

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	dumpResponse(resp)

	location, err := resp.Location()

	if err != nil {
		return "", err
	}

	//result, err := ioutil.ReadAll(resp.Body)
	return location.String(), err
}

func Login(username, password, referer string) (string, error) {
	data := url.Values{
		"TARGET":   {"-SM-https://pointage.adp.com/igested/2_02_01/pointage"},
		"USER":     {username},
		"PASSWORD": {password},
	}

	// follow redirect
	client.CheckRedirect = nil

	req, err := http.NewRequest("POST", "https://pointage.adp.com/ipclogin/1/loginform.fcc", strings.NewReader(data.Encode()))

	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", "https://pointage.adp.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	dumpRequest(req)

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	dumpResponse(resp)

	result, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if !strings.Contains(string(result), "Enregistrer mon horaire") {
		return "", errors.New("unable to login")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(result))

	if err != nil {
		return "", err
	}

	return strings.Title(strings.ToLower(strings.TrimSpace(doc.Find(".texte_nom_prenom").Text())[11:])), nil
}

func History() ([]string, error) {
	data := url.Values{
		"ACTION":      {"POI_CONS"},
		"FONCTION":    {""},
		"GMT_DATE":    {time.Now().Format("2006/01/02 15:04:05")},
		"USER_OFFSET": {"MTIw"},
	}

	result, err := submitForm(data)

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(result))

	if err != nil {
		return nil, err
	}

	var hours []string

	doc.Find(".cel_liste.col_date").Each(func(i int, s *goquery.Selection) {
		hours = append(hours, strings.TrimSpace(s.Text()))
	})

	return hours, nil
}

func Pointage() error {
	data := url.Values{
		"ACTION":      {"ENR_PRES"},
		"FONCTION":    {""},
		"GMT_DATE":    {time.Now().Format("2006/01/02 15:04:05")},
		"USER_OFFSET": {"MTIw"},
	}

	result, err := submitForm(data)

	if err != nil {
		return err
	}

	if !strings.Contains(string(result), "Votre saisie a bien") {
		return errors.New("echec du pointage")
	}

	return nil
}

func submitForm(data url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", "https://pointage.adp.com/igested/2_02_01/pointage", strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Referer", "https://pointage.adp.com/igested/2_02_01/pointage")
	req.Header.Set("Origin", "https://pointage.adp.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	dumpRequest(req)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	dumpResponse(resp)

	return ioutil.ReadAll(resp.Body)
}
