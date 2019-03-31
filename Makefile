default:
	CGO_ENABLED=0 go build
	sherpadoc Meetquiz >assets/meetquiz.json
	sherpadoc MeetquizPublic >assets/meetquizpublic.json
	#-rm meetquiz.zip
	#(cd assets && zip -q0 ../meetquiz.zip meetquiz.json)
	#cat meetquiz.zip >>meetquiz
	./meetquiz testdata/quiz.json

test:
	go test

fmt:
	go fmt ./...

clean:
	go clean
