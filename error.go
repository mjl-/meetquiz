package main

import (
	"bitbucket.org/mjl/sherpa"
)

func userError(s string) {
	panic(&sherpa.Error{Code: "user:error", Message: s})
}

func serverError(s string) {
	panic(&sherpa.Error{Code: "server:error", Message: s})
}

func checkUserError(err error) {
	if err != nil {
		userError(err.Error())
	}
}

func checkServerError(err error) {
	if err != nil {
		serverError(err.Error())
	}
}
