package zrouter

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ChiContextAdapterSuite struct {
	suite.Suite
}

func (suite *ChiContextAdapterSuite) TestChiContextAdapter() {
	r := chi.NewRouter()
	r.Get("/hello/{name}", func(w http.ResponseWriter, req *http.Request) {
		adapter := &chiContextAdapter{ctx: w, req: req}
		assert.NotNil(suite.T(), adapter.Request())

		var input struct {
			Message string `json:"message"`
		}
		err := adapter.BindJSON(&input)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "Hello", input.Message)

		adapter.JSON(http.StatusOK, map[string]string{"response": "OK"})
		adapter.Header("Custom-Header", "CustomValue")
	})

	body := bytes.NewBuffer([]byte(`{"message":"Hello"}`))
	req := httptest.NewRequest("GET", "/hello/world?test=query", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var output map[string]string
	err := json.NewDecoder(rec.Body).Decode(&output)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "OK", output["response"])
	assert.Equal(suite.T(), "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(suite.T(), "CustomValue", rec.Header().Get("Custom-Header"))
}

func TestChiContextAdapterSuite(t *testing.T) {
	suite.Run(t, new(ChiContextAdapterSuite))
}
