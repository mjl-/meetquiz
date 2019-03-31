meetquiz - interactive quizes with an audience

NOTE: WORK IN PROGRESS



## user experience
- look at projected view of public page, where URL and code are shown.
- go to URL
- enter quiz code and go.
- enter name and signup. perhaps let users select an icon?
- wait for first question from host
- repeat:
	- read question on public page.
	- wait until answering is opened.
	- select answer on phone
	- wait until answering is closed, this is displayed on the phone.
	- wait until results of question are shown. correct answer and number of users for each shown on public page. points scored on phone.
- see ending non-question slide, preparing to go to the results.
- see total ranking on public page. see own ranking on phone.
- the end.

## host/public experience
- assuming the quiz itself has already been configured...
- open public URL, hit the "open signups" button
- show the signup page, including users joined so far. show button to go to first question.
- have a button to start the quiz in a new tab. the page will be shown on projector. should show the URL and quiz code in big font, perhaps a QR as well.
- on start-page, display the names that have joined. and perhaps their icons. and a "next" button to go to first question.
- repeat:
	- go to next page (question)
	- on next-button, show the answers (A, B, C, etc), send cue to phones, show countdown for answering
	- receive responses, display number of responses so far, perhaps with their icons
	- when countdown is done, show the correct answer and the stats
	- click next button
- after last question, show that this is the end
- show how quickly people responded
- show the results

## admin/creation experience
- TODO, probably create new quiz with random ID by sending link to email, in which user can paste quiz JSON, upload files, reset answers, show answers, show link to go to public page to start the quiz.


# first approach, just proof of concept
- just run this as a single binary, no failovers and complications
- start with SSE to clients, maybe websockets later
- simple JS for frontend
- simple single JSON file storing questions and answers, not configurable.
- no auth
- no answer/progress storage.

To run this, first install sherpadoc:

	go get bitbucket.org/mjl/sherpadoc/

Now run the program:

	make

Now open the URLs:

	http://localhost:8008/public/ for the page showing the questions, answers, results typically projected on a big screen.

	http://localhost:8008/ for the page to be opened by users on their mobile phones.


## todo (a lot!)
- need to know for which user an sse connection is, sessionID and send user-specific messages (correctness of each answers with points scored, final rank)
- sse reconnect and restart should work well for both user and public.
- allow changing answer until question is closed
- prettier pages, optimize user frontend for mobile, allow quiz to be configured with images?
- store the results to disk
- multiple quizes in single instance
- quiz-creation flow
- higher-availability with multiple instances, store state somewhere.
