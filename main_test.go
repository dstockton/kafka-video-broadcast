package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"os"
	"testing"
)

// This function is used for setup before executing the test functions
func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}

func testHTTPResponse(t *testing.T, r *gin.Engine, req *http.Request, f func(w *httptest.ResponseRecorder) bool) {
	// Create a response recorder
	w := httptest.NewRecorder()

	// Create the service and process the above request.
	r.ServeHTTP(w, req)

	if !f(w) {
		t.Fail()
	}
}

func TestRouter(t *testing.T) {
	tests := []struct {
		path             string
		method           string
		expectedStatus   int
		expectedResponse map[string]interface{}
	}{
		{
			path:           "/not-found",
			method:         "GET",
			expectedStatus: http.StatusNotFound,
		},
		{
			path:           "/ping",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedResponse: gin.H{
				"message": "pong",
			},
		},
		{
			path:           "/version",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	router, _ := getRouter()

	for _, test := range tests {
		test := test
		t.Run(test.path, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.path, nil)

			testHTTPResponse(t, router, req, func(w *httptest.ResponseRecorder) bool {
				statusMatch := w.Code == test.expectedStatus
				if !statusMatch {
					t.Logf("Got status: %v, expectedStatus: %v", w.Code, test.expectedStatus)
					return false
				}

				if test.expectedStatus != http.StatusOK || test.expectedResponse == nil {
					return true
				}

				expectedResponseString, err := json.Marshal(test.expectedResponse)
				if err != nil {
					t.Errorf("Could not Marshal expectedResponse: %v", test.expectedResponse)
				}

				p, err := ioutil.ReadAll(w.Body)
				pageOK := err == nil && string(p) == string(expectedResponseString)
				if !pageOK {
					t.Errorf("Response did not match.\n\tExpected: %v\n\tActual: %v", string(expectedResponseString), string(p))
				}

				return pageOK
			})
		})
	}
}
