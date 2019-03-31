package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// xxx remove duplication and horrible error checking

func writeUserEvent(w http.ResponseWriter, e userEvent) (err error) {
	defer recover()
	lcheck := func() {
		if err != nil {
			panic(err)
		}
	}

	buf := &bytes.Buffer{}
	_, err = buf.Write([]byte("data: "))
	lcheck()
	err = json.NewEncoder(buf).Encode(e)
	lcheck()
	_, err = buf.Write([]byte("\n\n"))
	lcheck()
	_, err = w.Write(buf.Bytes())
	return
}

func writePublicEvent(w http.ResponseWriter, e publicEvent) (err error) {
	defer recover()
	lcheck := func() {
		if err != nil {
			panic(err)
		}
	}

	buf := &bytes.Buffer{}
	_, err = buf.Write([]byte("data: "))
	lcheck()
	err = json.NewEncoder(buf).Encode(e)
	lcheck()
	_, err = buf.Write([]byte("\n\n"))
	lcheck()
	_, err = w.Write(buf.Bytes())
	return
}
