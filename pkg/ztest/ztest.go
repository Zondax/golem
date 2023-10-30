package ztest

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RequestAssertionParams struct {
	T                   *testing.T
	Router              http.Handler
	Responses           map[string]interface{}
	Method              string
	URL                 string
	ExpectedStatusCode  int
	ExpectedResponseKey string
	Body                io.Reader
}

func MakeRequestAndAssert(params RequestAssertionParams) {
	req, err := http.NewRequest(params.Method, params.URL, params.Body)
	if err != nil {
		params.T.Fatal(err)
	}

	w := httptest.NewRecorder()
	params.Router.ServeHTTP(w, req)
	assert.Equal(params.T, params.ExpectedStatusCode, w.Code)

	var obtainedResponse interface{}
	err = json.Unmarshal(w.Body.Bytes(), &obtainedResponse)
	if err != nil {
		params.T.Fatal(err)
	}

	obtainedJSON, err := json.Marshal(obtainedResponse)
	if err != nil {
		params.T.Fatal(err)
	}

	expectedVal, exists := params.Responses[params.ExpectedResponseKey]
	if !exists {
		params.T.Fatalf("Key %s does not exist in testResponses", params.ExpectedResponseKey)
	}

	expectedJSON, err := json.Marshal(expectedVal)
	if err != nil {
		params.T.Fatal(err)
	}

	assert.Equal(params.T, string(expectedJSON), string(obtainedJSON))
}
