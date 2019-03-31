package main

// MeetquizPublic is the API running the quiz, going to next question, etc.
type MeetquizPublic struct {
}

// Start opens the quiz for questions.
// If the quiz was already started, an error is returned.
func (MeetquizPublic) Start() {
	rc := make(chan error, 1)
	ctl.start <- rc
	err := <-rc
	checkUserError(err)
}

// Quiz returns all question for the quiz.
func (MeetquizPublic) Quiz() QuizConfig {
	return quizConfig
}

// OpenNextQuestion opens the next question for answering.
// If there is no more question, an error is returned.
func (MeetquizPublic) OpenNextQuestion() {
	rc := make(chan error, 1)
	ctl.openNextQuestion <- rc
	err := <-rc
	checkUserError(err)
}

type AnswerCount struct {
	Label string
	Count int
}

type QuestionResults struct {
	AnswerCounts []AnswerCount
}

type questionResultsResponse struct {
	r   QuestionResults
	err error
}

// CloseQuestion closes the current question for answering and returns the answer statistics.
// If there is no open question, an error is returned.
func (MeetquizPublic) CloseQuestion() QuestionResults {
	rc := make(chan questionResultsResponse, 1)
	ctl.closeQuestion <- rc
	resp := <-rc
	checkUserError(resp.err)
	return resp.r
}

// ShowAnswers marks the current question as shown.
// If there is no open question, an error is returned.
func (MeetquizPublic) ShowAnswers() {
	rc := make(chan error, 1)
	ctl.showAnswers <- rc
	err := <-rc
	checkUserError(err)
}

// UserResult holds the score of a user.
type UserResult struct {
	Name  string
	Score int
}

// Results represents the final results (ranking) of users, including points scored.
type Results struct {
	Users []UserResult
}

type resultsResponse struct {
	r   Results
	err error
}

// FinalResults returns the final results for the quiz, and sends a rank event to all users.
func (MeetquizPublic) FinalResults() Results {
	rc := make(chan resultsResponse, 1)
	ctl.finalResults <- rc
	resp := <-rc
	checkUserError(resp.err)
	return resp.r
}
