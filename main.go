package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	twiml "github.com/wherethebitsroam/twiml"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

type House struct {
	AccessNumber int64
	SecretNumber int64
	Roommates    []Roommate
}

type Roommate struct {
	Name  string
	Phone string
}

var sounds = []string{
	"http://phonebox.isaqt.com/sfx/lionking.mp3",
	"http://phonebox.isaqt.com/sfx/opensesame.mp3",
	"http://phonebox.isaqt.com/sfx/itsdeath.mp3",
}

func randomSound() string {
	return sounds[rand.Intn(len(sounds))]
}

func main() {
	houseConfig := House{}
	bytes, err := ioutil.ReadFile("roomies.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &houseConfig)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/secret", func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			panic(err)
		}
		if val, ok := req.Form["Digits"]; ok {
			intDigits, err := strconv.ParseInt(val[0], 0, 64)
			if err != nil {
				panic(err)
			}

			r := twiml.NewResponse()
			w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"))
			if intDigits == houseConfig.SecretNumber {
				r.Say("Welcome.", &twiml.SayAttr{Voice: "woman"})
				r.Play("", &twiml.PlayAttr{Digits: "9"})
				r.Play(randomSound(), &twiml.PlayAttr{})
				r.ToXML(w)
				return
			}

			r.Hangup()
			r.ToXML(w)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no digits"))
	})
	http.HandleFunc("/sort", func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			panic(err)
		}
		if val, ok := req.Form["Digits"]; ok {
			intDigits, err := strconv.ParseInt(val[0], 0, 64)
			if err != nil {
				panic(err)
			}

			r := twiml.NewResponse()
			w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"))

			if intDigits == houseConfig.AccessNumber {
				// Gather for the secret handler
				ga := twiml.GatherAttr{
					Action:  "/secret",
					Method:  "GET",
					Timeout: 15,
				}
				g := r.Gather(&ga)
				g.Say("End it with a pound.", &twiml.SayAttr{
					Voice: "man",
				})
				r.ToXML(w)
				return
			}

			if intDigits == 0 {
				intDigits = 1 // lol
			} else {
				intDigits -= 1
			}

			if int(intDigits) > len(houseConfig.Roommates) {
				r.Say("Try again.", &twiml.SayAttr{})
				r.Redirect("/greet", &twiml.RedirectAttr{})
				r.ToXML(w)
				return
			}

			// Dial the rooomate!
			d := r.Dial(&twiml.DialAttr{})
			d.Number(houseConfig.Roommates[intDigits].Phone, &twiml.NumberAttr{})
			r.Say("Goodbye.", &twiml.SayAttr{})
			r.ToXML(w)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no digits"))
	})
	http.HandleFunc("/greet", func(w http.ResponseWriter, req *http.Request) {
		var speech string
		for i, roomie := range houseConfig.Roommates {
			speech += fmt.Sprintf("For %s press %d", roomie.Name, i+1)
			if i+1 != len(houseConfig.Roommates) {
				speech += ", "
			}
		}
		w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"))
		r := twiml.NewResponse()
		ga := twiml.GatherAttr{
			Action:    "/sort",
			Method:    "GET",
			Timeout:   10,
			NumDigits: 1,
		}
		g := r.Gather(&ga)
		g.Say(speech, &twiml.SayAttr{
			Voice: "woman",
		})
		r.Say("Didn't hear you, goodbye.", &twiml.SayAttr{})
		r.ToXML(w)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
