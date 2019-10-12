package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type (
	Forvo struct {
		*http.Client
		Key string
	}

	Config struct {
		Language string `json:"language"` // See https://forvo.com/languages-codes/
		Key      string `json:"api_key"`
	}

	// See https://api.forvo.com/demo
	Results struct {
		Attributes struct {
			Total int `json:"total"`
		} `json:"attributes"`
		Items []Item `json:"items"`
	}

	Item struct {
		ID      int    `json:"id"`
		Word    string `json:"word"`
		PathMp3 string `json:"pathmp3"`
	}
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log(fmt.Sprintf("ERROR! %s\n%s", r, string(debug.Stack())))
		}
	}()

	config := loadConfig()

	forvo := Forvo{
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		Key: config.Key,
	}

	for _, word := range loadList() {
		log(fmt.Sprintf("downloading '%s'", word))
		forvo.Download(word, config.Language)
	}
}

func (this *Forvo) Download(word, language string) {
	mp3 := word + ".mp3"

	if exists(mp3) {
		log(fmt.Sprintf("'%s' already exists, skipping", mp3))
		return
	}

	uri := fmt.Sprintf(
		"https://apifree.forvo.com"+
			"/key/%s"+
			"/format/json"+
			"/action/word-pronunciations"+
			"/word/%s"+
			"/language/%s"+
			"/order/rate-desc"+
			"/limit/1",
		this.Key,
		word,
		language,
	)

	resp1, err := this.Get(uri)
	if err != nil {
		log(fmt.Sprintf("failed to search for '%s': %s", word, err))
		return
	}

	defer resp1.Body.Close()

	if resp1.StatusCode == 429 {
		panic("Daily API limit reached")
	}

	if resp1.StatusCode < 200 || resp1.StatusCode >= 300 {
		log(fmt.Sprintf(
			"failed to search for word, status code %d\n%s",
			resp1.StatusCode,
			body(resp1.Body),
		))
		return
	}

	results := new(Results)
	decoder := json.NewDecoder(resp1.Body)
	if err := decoder.Decode(results); err != nil {
		panic(wrap("failed to decode search results", err))
	}

	if len(results.Items) == 0 {
		log(fmt.Sprintf("no results for '%s'", word))
		return
	}

	resp2, err := this.Get(results.Items[0].PathMp3)
	if err != nil {
		log(fmt.Sprintf("failed to download '%s' audio", err))
		return
	}

	if resp2.StatusCode == 429 {
		panic("Daily API limit reached")
	}

	if resp2.StatusCode < 200 || resp2.StatusCode >= 300 {
		log(fmt.Sprintf(
			"failed to download word audio, status code %d\n%s",
			resp1.StatusCode,
			body(resp1.Body),
		))
		return
	}

	f, err := os.OpenFile(mp3, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		defer f.Close()
	}

	if _, err := io.Copy(f, resp2.Body); err != nil {
		log(fmt.Sprintf("failed to write '%s'", mp3))
		return
	}
}

func loadConfig() *Config {
	if buf, err := ioutil.ReadFile("forvo.json"); err != nil {
		panic(wrap("couldn't read configuration file", err))
	} else {
		config := new(Config)
		if err := json.Unmarshal(buf, config); err != nil {
			panic(wrap("couldn't deserialize configuration file", err))
		} else {
			return config
		}
	}
}

func loadList() []string {
	buf, err := ioutil.ReadFile("forvo.txt")
	if err != nil {
		panic("couldn't read word file")
	}

	lines := strings.Split(strings.Replace(string(buf), "\r\n", "\n", -1), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines
}

func log(message string) {
	f, err := os.OpenFile("forvo.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		defer f.Close()
	}

	data := fmt.Sprintf(
		"[%s] %s\n",
		time.Now().Format(time.RFC1123),
		message,
	)

	if _, err = f.WriteString(data); err != nil {
		panic(err)
	}
}

func wrap(msg string, err error) error {
	return errors.New(fmt.Sprintf("%s: %s", msg, err))
}

func body(r io.ReadCloser) string {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		panic(err)
	} else {
		str := strings.TrimSpace(buf.String())
		if str == "" {
			return "[NO RESPONSE BODY]"
		} else {
			return str
		}
	}
}

func exists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
