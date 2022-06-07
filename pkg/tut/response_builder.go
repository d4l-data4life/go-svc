package tut

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseBuilder func(http.ResponseWriter)

func BuildResponse(fns ...ResponseBuilder) ResponseBuilder {
	return func(w http.ResponseWriter) {
		for _, f := range fns {
			f(w)
		}
	}
}

func RespWithStatus(statusCode int) ResponseBuilder {
	return func(w http.ResponseWriter) {
		w.WriteHeader(statusCode)
	}
}

func RespWithJSONBody(body interface{}) ResponseBuilder {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Errorf("can't marshal body: %v", err))
	}
	return func(w http.ResponseWriter) {
		_, _ = w.Write(jsonBody)
	}
}
