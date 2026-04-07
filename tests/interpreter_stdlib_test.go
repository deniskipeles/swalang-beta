package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/deniskipeles/pylearn/internal/testhelpers"

	// Ensure stdlib modules are registered
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyhttp"
)

// No need for testEvalStdlib or duplicated helpers

func TestHttpBuiltinModuleGet(t *testing.T) {
	// Using httpbin.org - requires internet connection for tests.
	// Consider using net/http/httptest for offline tests if feasible.

	t.Run("Successful GET request", func(t *testing.T) {
		input := `
import http
response = http.get("https://httpbin.org/get?param=test") # Add query param
response # Evaluate the response object
`
		evaluated := testhelpers.Eval(t, input)
		respObj, ok := testhelpers.TestHttpResponseObject(t, evaluated) // Check type
		if !ok {
			return
		}

		// Check status code directly on the response object
		testhelpers.TestIntegerObject(t, respObj.StatusCode, 200)

		// Check text content
		respText := respObj.Text.Value
		if !strings.Contains(respText, `"url": "https://httpbin.org/get?param=test"`) {
			t.Errorf("Response text missing expected URL. Got:\n%s", respText)
		}
		if !strings.Contains(respText, `"param": "test"`) {
			t.Errorf("Response text missing expected query param. Got:\n%s", respText)
		}

		// Test attribute access via Pylearn evaluation
		statusEval := testhelpers.Eval(t, input+"\nresponse.status_code")
		testhelpers.TestIntegerObject(t, statusEval, 200)

		textEval := testhelpers.Eval(t, input+"\nresponse.text")
		testhelpers.TestStringObject(t, textEval, respText) // Should match the direct access
	})

	t.Run("GET request with 404", func(t *testing.T) {
		input := `
import http
response = http.get("https://httpbin.org/status/404")
response.status_code # Evaluate status code attribute
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestIntegerObject(t, evaluated, 404)

		// Also check the response object type and status
		fullResponseEval := testhelpers.Eval(t, `import http; http.get("https://httpbin.org/status/404")`)
		respObj, ok := testhelpers.TestHttpResponseObject(t, fullResponseEval)
		if ok {
			testhelpers.TestIntegerObject(t, respObj.StatusCode, 404)
		}
	})

	t.Run("GET error cases", func(t *testing.T) {
		errTests := []struct {
			input    string
			errParts []string
		}{
			{`import http; http.get("invalid-url-scheme:")`, []string{"ValueError", "Invalid URL"}},                                                         // More specific error type?
			{`import http; http.get("http://non-existent-domain-pylearn-test.invalid")`, []string{"RequestError", "Failed to perform GET", "no such host"}}, // Network error
			{`import http; http.get()`, []string{"TypeError", "http.get() takes exactly 1 argument (0 given)"}},
			{`import http; http.get(123)`, []string{"TypeError", "http.get() argument must be a string"}},
		}
		for _, et := range errTests {
			t.Run(et.input+" (error)", func(t *testing.T) {
				evalErr := testhelpers.Eval(t, et.input)
				testhelpers.TestErrorObject(t, evalErr, et.errParts...)
			})
		}
	})

	t.Run("Access non-existent attribute on Response", func(t *testing.T) {
		input := `
import http
response = http.get("https://httpbin.org/get")
response.non_existent_attr # Access invalid attribute
`
		evaluated := testhelpers.Eval(t, input)
		testhelpers.TestErrorObject(t, evaluated, "AttributeError", "'HTTPResponse' object has no attribute 'non_existent_attr'")
	})
}

func TestHttpBuiltinModulePost(t *testing.T) {
	// Using httpbin.org/post

	t.Run("Successful POST request", func(t *testing.T) {
		postData := `{"key": "value", "num": 123}`
		pylearnPostDataLiteral := fmt.Sprintf("%q", postData) // Ensure proper quoting for Pylearn

		input := fmt.Sprintf(`
import http
post_body = %s
response = http.post("https://httpbin.org/post", post_body)
response # Evaluate the response object
`, pylearnPostDataLiteral)

		evaluated := testhelpers.Eval(t, input)
		respObj, ok := testhelpers.TestHttpResponseObject(t, evaluated)
		if !ok {
			return
		}

		testhelpers.TestIntegerObject(t, respObj.StatusCode, 200)

		// Check echoed data in response text
		respText := respObj.Text.Value
		expectedEcho := fmt.Sprintf(`"data": "%s"`, postData) // httpbin echoes raw string data like this
		if !strings.Contains(respText, expectedEcho) {
			t.Errorf("Response text did not contain expected echoed data substring %q. Got:\n%s", expectedEcho, respText)
		}
		// Check if json field reflects parsed data (if httpbin does that)
		if !strings.Contains(respText, `"key": "value"`) {
			t.Errorf("Response 'json' field missing expected key. Got:\n%s", respText)
		}

		// Check attribute access
		statusEval := testhelpers.Eval(t, input+"\nresponse.status_code")
		testhelpers.TestIntegerObject(t, statusEval, 200)
	})

	t.Run("POST error cases", func(t *testing.T) {
		errTests := []struct {
			input    string
			errParts []string
		}{
			{`import http; http.post("url_only")`, []string{"TypeError", "takes exactly 2 arguments", "1 given"}},
			{`import http; http.post("url", "data", "extra")`, []string{"TypeError", "takes exactly 2 arguments", "3 given"}},
			{`import http; http.post()`, []string{"TypeError", "takes exactly 2 arguments", "0 given"}},
			{`import http; http.post(123, "data")`, []string{"TypeError", "argument 1 (url) must be a string"}},
			{`import http; http.post("url", 123)`, []string{"TypeError", "argument 2 (data) must be a string"}}, // Assuming data must be string for now
			{`import http; http.post("invalid:", "data")`, []string{"ValueError", "Invalid URL"}},
			{`import http; http.post("http://non-existent-domain-pylearn-test.invalid", "data")`, []string{"RequestError", "Failed to perform POST", "no such host"}},
		}
		for _, et := range errTests {
			t.Run(et.input+" (error)", func(t *testing.T) {
				evalErr := testhelpers.Eval(t, et.input)
				testhelpers.TestErrorObject(t, evalErr, et.errParts...)
			})
		}
	})
}
