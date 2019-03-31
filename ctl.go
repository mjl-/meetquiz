package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

type register struct {
	name string
	rc   chan string // respond with sessionID, or empty for error (eg name taken or not yet started)
}

type newAnswer struct {
	sessionID string
	question  int
	label     string
	rc        chan error
}

var ctl struct {
	start            chan chan error
	register         chan register
	unregister       chan string // unregister by id
	answer           chan newAnswer
	openNextQuestion chan chan error
	closeQuestion    chan chan questionResultsResponse
	showAnswers      chan chan error
	finalResults     chan chan resultsResponse

	unregisterUserEvents chan chan userEvent
	registerUserEvents   chan chan userEvent

	unregisterPublicEvents chan chan publicEvent
	registerPublicEvents   chan chan publicEvent
}

func init() {
	rand.Seed(time.Now().UnixNano())
	ctl.start = make(chan chan error, 1)
	ctl.register = make(chan register, 1)
	ctl.unregister = make(chan string, 1)
	ctl.answer = make(chan newAnswer, 1)
	ctl.openNextQuestion = make(chan chan error, 1)
	ctl.closeQuestion = make(chan chan questionResultsResponse, 1)
	ctl.showAnswers = make(chan chan error, 1)
	ctl.finalResults = make(chan chan resultsResponse, 1)

	ctl.unregisterUserEvents = make(chan chan userEvent, 1)
	ctl.registerUserEvents = make(chan chan userEvent, 1)

	ctl.unregisterPublicEvents = make(chan chan publicEvent, 1)
	ctl.registerPublicEvents = make(chan chan publicEvent, 1)
}

var (
	errBadSession        = errors.New("bad session")
	errQuestionClosed    = errors.New("question closed")
	errBadAnswer         = errors.New("bad answer")
	errWrongQuestion     = errors.New("wrong question")
	errAlreadyStarted    = errors.New("already started")
	errNotYetStarted     = errors.New("not yet started")
	errQuestionStillOpen = errors.New("question still open")
	errNoOpenQuestion    = errors.New("no open question")
	errNoMoreQuestions   = errors.New("no more questions")
	errNotFinishedYet    = errors.New("not finished yet")
)

type answered struct {
	label    string
	duration time.Duration
}

type client struct {
	sessionID string
	name      string
	answers   map[int]answered
	points    int
}

func newID() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}

func control() {
	started := false
	clients := map[string]*client{}
	question := -1
	questionOpen := false
	var questionStart time.Time
	var users []chan userEvent
	var publics []chan publicEvent

	sendUserEvent := func(e userEvent) {
		for _, c := range users {
			select {
			case c <- e:
			default:
				log.Println("blocking send to userEvent, event skipped")
			}
		}
	}

	sendPublicEvent := func(e publicEvent) {
		for _, c := range publics {
			select {
			case c <- e:
			default:
				log.Println("blocking send to publicEvent, event skipped")
			}
		}
	}

	start := func(rc chan error) {
		if started {
			rc <- errAlreadyStarted
			return
		}
		started = true
		rc <- nil
	}

	register := func(r register) {
		if !started {
			r.rc <- "" // xxx errNotYetStarted
			return
		}
		for _, c := range clients {
			if c.name == r.name {
				// Name already in use.
				r.rc <- ""
				return
			}
		}
		c := &client{
			sessionID: newID(),
			name:      r.name,
			answers:   map[int]answered{},
			points:    0,
		}
		clients[c.sessionID] = c
		r.rc <- c.sessionID

		e := publicEvent{
			Event: "join",
			Params: map[string]interface{}{
				"Name": r.name,
			},
		}
		sendPublicEvent(e)
	}

	unregister := func(sessionID string) {
		delete(clients, sessionID)
	}

	findAnswer := func(label string) (Answer, bool) {
		for _, a := range quizConfig.Questions[question].Answers {
			if a.Label == label {
				return a, true
			}
		}
		return Answer{}, false
	}

	answer := func(a newAnswer) {
		if !started {
			a.rc <- errNotYetStarted
			return
		}
		c, ok := clients[a.sessionID]
		if !ok {
			a.rc <- errBadSession
			return
		}
		if a.question < 0 || a.question != question {
			a.rc <- errWrongQuestion
			return
		}
		if !questionOpen {
			a.rc <- errQuestionClosed
			return
		}
		_, ok = findAnswer(a.label)
		if !ok {
			a.rc <- errBadAnswer
			return
		}
		c.answers[question] = answered{
			label:    a.label,
			duration: time.Now().Sub(questionStart),
		}
		a.rc <- nil

		e := publicEvent{
			Event: "answer",
			Params: map[string]interface{}{
				"Question":    question,
				"Name":        c.name,
				"AnswerLabel": a.label,
			},
		}
		sendPublicEvent(e)
	}

	openNextQuestion := func(rc chan error) {
		if !started {
			rc <- errNotYetStarted
			return
		}
		if questionOpen {
			rc <- errQuestionStillOpen
			return
		}
		next := question + 1
		if next >= len(quizConfig.Questions) {
			rc <- errNoMoreQuestions
			return
		}
		question = next
		questionOpen = true
		questionStart = time.Now()
		rc <- nil

		// Send event to users.
		answerLabels := []string{}
		for _, a := range quizConfig.Questions[question].Answers {
			answerLabels = append(answerLabels, a.Label)
		}
		e := userEvent{
			Event: "questionOpen",
			Params: map[string]interface{}{
				"Question":     question,
				"AnswerLabels": answerLabels,
			},
		}
		sendUserEvent(e)
	}

	closeQuestion := func(rc chan questionResultsResponse) {
		if !started {
			rc <- questionResultsResponse{err: errNotYetStarted}
			return
		}
		if !questionOpen {
			rc <- questionResultsResponse{err: errNoOpenQuestion}
			return
		}
		questionOpen = false

		counts := map[string]int{} // label to count
		for _, c := range clients {
			a, ok := c.answers[question]
			if ok {
				counts[a.label]++
			}
		}
		answerCounts := []AnswerCount{}
		for _, a := range quizConfig.Questions[question].Answers {
			answerCounts = append(answerCounts, AnswerCount{Label: a.Label, Count: counts[a.Label]})
		}
		qr := QuestionResults{
			AnswerCounts: answerCounts,
		}

		rc <- questionResultsResponse{r: qr}

		e := userEvent{
			Event: "questionClosed",
			Params: map[string]interface{}{
				"Question": question,
			},
		}
		sendUserEvent(e)
	}

	showAnswers := func(rc chan error) {
		if !started {
			rc <- errNotYetStarted
			return
		}
		if questionOpen {
			rc <- errQuestionStillOpen
			return
		}

		// xxx make it a per-user message, including whether answer was correct, and the points scored.
		e := userEvent{
			Event:  "answersShown",
			Params: map[string]interface{}{},
		}
		sendUserEvent(e)
	}

	finalResults := func(rc chan resultsResponse) {
		if !started {
			rc <- resultsResponse{err: errNotYetStarted}
			return
		}
		if questionOpen {
			rc <- resultsResponse{err: errQuestionStillOpen}
			return
		}
		if question < len(quizConfig.Questions)-1 {
			rc <- resultsResponse{err: errNotFinishedYet}
			return
		}

		r := Results{
			Users: []UserResult{},
		}
		for _, c := range clients {
			score := 0
			for i, q := range quizConfig.Questions {
				ans, ok := c.answers[i]
				if !ok {
					continue
				}
				for _, a := range q.Answers {
					if a.Label == ans.label {
						score += a.Points
						break
					}
				}
			}
			ur := UserResult{
				Name:  c.name,
				Score: score,
			}
			r.Users = append(r.Users, ur)
		}
		sort.Slice(r.Users, func(i, j int) bool {
			return r.Users[i].Score > r.Users[j].Score
		})
		rc <- resultsResponse{r: r}

		// xxx send per-user rank. need to know which user a userEvent chan is for...
		e := userEvent{
			Event: "ranking",
		}
		sendUserEvent(e)
	}

	unregisterUserEvents := func(ec chan userEvent) {
		for i, c := range users {
			if c == ec {
				users[i] = users[len(users)-1]
				users = users[:len(users)-1]
				return
			}
		}
		panic("userEvent channel not found")
	}

	registerUserEvents := func(ec chan userEvent) {
		users = append(users, ec)
	}

	unregisterPublicEvents := func(ec chan publicEvent) {
		for i, c := range publics {
			if c == ec {
				publics[i] = publics[len(publics)-1]
				publics = publics[:len(publics)-1]
				return
			}
		}
		panic("publicEvent channel not found")
	}

	registerPublicEvents := func(ec chan publicEvent) {
		publics = append(publics, ec)
	}

	for {
		select {
		case rc := <-ctl.start:
			start(rc)

		case r := <-ctl.register:
			register(r)

		case sessionID := <-ctl.unregister:
			unregister(sessionID)

		case a := <-ctl.answer:
			answer(a)

		case rc := <-ctl.openNextQuestion:
			openNextQuestion(rc)

		case rc := <-ctl.closeQuestion:
			closeQuestion(rc)

		case rc := <-ctl.showAnswers:
			showAnswers(rc)

		case rc := <-ctl.finalResults:
			finalResults(rc)

		case ec := <-ctl.unregisterUserEvents:
			unregisterUserEvents(ec)

		case ec := <-ctl.registerUserEvents:
			registerUserEvents(ec)

		case ec := <-ctl.unregisterPublicEvents:
			unregisterPublicEvents(ec)

		case ec := <-ctl.registerPublicEvents:
			registerPublicEvents(ec)
		}
	}
}
