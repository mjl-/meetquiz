package main

// Meetquiz is the API answering quizes, used on mobile phones.
type Meetquiz struct {
}

// Join starts a new session for a participating player.
func (Meetquiz) Join(name string) (sessionID string) {
	r := register{name, make(chan string, 1)}
	ctl.register <- r
	sessionID = <-r.rc
	if sessionID == "" {
		userError("cannot join under that name, it might already be taken or the quiz might not yet be started")
	}
	return
}

// Answer registers an answer for the given question.
// If this is not the current question, an error is returned.
// If the question is closed (timer expired), an error is returned.
// If answerLabel is not a valid answer to the question, an error is returned.
// Success will be returned when the answer has been received. This does not mean the answer is correct.
func (Meetquiz) Answer(sessionID string, question int, answerLabel string) {
	na := newAnswer{sessionID, question, answerLabel, make(chan error, 1)}
	ctl.answer <- na
	err := <-na.rc
	checkUserError(err)
}
