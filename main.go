package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/mjl-/httpasset"
	"github.com/mjl-/sherpa"
	"github.com/mjl-/sherpadoc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	address = flag.String("address", "localhost:8008", "address to listen on")

	httpFS = httpasset.Init("assets")

	meetquizVersion = "dev"

	quizConfig QuizConfig
)

type Answer struct {
	Label   string
	Correct bool
	Points  int
}

// todo: richer text, possibly question, markdown or just html
type Question struct {
	Question string
	Answers  []Answer
}

type QuizConfig struct {
	Title     string
	Questions []Question
}

func check(err error, action string) {
	if err != nil {
		log.Fatalf("%s: %s\n", action, err)
	}
}

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		log.Println("usage: meetquiz [flags]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}

	f, err := os.Open(args[0])
	check(err, "opening quiz file")
	err = json.NewDecoder(f).Decode(&quizConfig)
	check(err, "parsing quiz config")
	f.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.FileServer(httpFS))

	parseSherpaDoc := func(name string) *sherpadoc.Section {
		doc := &sherpadoc.Section{}
		ff, err := httpFS.Open(name)
		check(err, "open sherpadoc json")
		err = json.NewDecoder(ff).Decode(doc)
		check(err, "parsing sherpadoc json")
		err = ff.Close()
		check(err, "close")
		return doc
	}

	userDoc := parseSherpaDoc("/meetquiz.json")
	userHandler, err := sherpa.NewHandler("/meetquiz/", meetquizVersion, Meetquiz{}, userDoc, nil)
	check(err, "making sherpa user handler")
	http.Handle("/meetquiz/", userHandler)

	publicDoc := parseSherpaDoc("/meetquizpublic.json")
	publicHandler, err := sherpa.NewHandler("/meetquizpublic/", meetquizVersion, MeetquizPublic{}, publicDoc, nil)
	check(err, "making sherpa public handler")
	http.Handle("/meetquizpublic/", publicHandler)

	http.HandleFunc("/userevents", userEvents)
	http.HandleFunc("/publicevents", publicEvents)

	go control()

	log.Printf("meetquiz %s, listening on %s\n", meetquizVersion, *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
