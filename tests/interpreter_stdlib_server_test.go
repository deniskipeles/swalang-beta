package tests

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
	"fmt"

	"github.com/deniskipeles/pylearn/internal/object" // Keep for IsError check potentially
	"github.com/deniskipeles/pylearn/internal/testhelpers"

	// Ensure stdlib modules are registered
	_ "github.com/deniskipeles/pylearn/internal/stdlib/httpserver"
)

// No duplicated helpers needed

func TestHttpServerModule(t *testing.T) {
	// Define Pylearn script to start the server
	// Using a fixed port is simpler for testing but can cause conflicts.
	// Port 0 would be ideal if the Go code could report the chosen port back.
	testPort := ":59877" // Use a different port than before
	serverAddr := "http://127.0.0.1" + testPort

	pylearnScript := fmt.Sprintf(`
import httpserver
import time

# Define a simple request handler in Pylearn
def my_handler(request):
    print(f"PY_HANDLER: Received {request.method} {request.path}") # Debug print in Pylearn
    path = request.path
    body = request.body
    headers = request.headers

    if path == "/hello":
        # Example: Echo back a header if present
        accept_enc = headers.get("accept-encoding", "default_encoding")
        resp_body = f"<h1>Hello Server!</h1> Header: {accept_enc}"
        # Implicitly returns 200 OK with the string as body
        return resp_body
    elif path == "/data" and request.method == "POST":
        return f"Received POST data: {body}"
    else:
        # How to signal 404? Current simple design returns 200 with body.
        # Need a way to return Response objects maybe. For now:
        return "Resource Not Found"

# Start the server. Make sure this runs in the background in the Go test.
# The httpserver.serve function likely needs to start the Go server in a goroutine internally.
address = %q
print(f"PY_SCRIPT: Attempting to start server on {address}...")
try:
    # Assuming httpserver.serve blocks if run directly,
    # or returns control if it starts a goroutine.
    # The Go test needs to handle the async nature.
    httpserver.serve(address, my_handler)
    print(f"PY_SCRIPT: Serve function returned (might run in background).")
except Exception as e:
    print(f"PY_SCRIPT: Error starting server: {e}")

# Keep script alive briefly if serve is non-blocking? Not ideal.
# time.sleep(0.1) # Avoid if possible
print("PY_SCRIPT: Script execution finished.")
`, testPort)

	// --- Start the Pylearn Server in a Goroutine ---
	serverErrChan := make(chan error, 1)
	go func() {
		// Use testhelpers.Eval to run the script
		// We don't expect a specific return value from the script itself,
		// but we capture potential errors during server setup/script execution.
		evalResult := testhelpers.Eval(t, pylearnScript)
		if object.IsError(evalResult) {
			// Report error from Pylearn execution back to main test goroutine
			serverErrChan <- fmt.Errorf("pylearn script error: %s", evalResult.Inspect())
		} else {
			// Signal successful script completion (doesn't mean server started listening yet)
			serverErrChan <- nil
		}
	}()

	// --- Wait for Server to Start or Script to Error ---
	// Poll the server endpoint or use the error channel. Polling is more robust.
	maxWait := 5 * time.Second
	startTime := time.Now()
	serverReady := false
	for time.Since(startTime) < maxWait {
		// Check if the script errored out first
		select {
		case err := <-serverErrChan:
			if err != nil {
				t.Fatalf("Failed to start/run Pylearn server script: %v", err)
			}
			// Script finished without error, but server might not be listening yet. Continue polling.
		default:
			// Non-blocking check
		}

		// Poll the server (e.g., a known endpoint like /hello)
		resp, err := http.Get(serverAddr + "/hello") // Simple GET to check connectivity
		if err == nil {
			resp.Body.Close() // Close the response body immediately
			t.Logf("Server responded, proceeding with tests.")
			serverReady = true
			break // Server is up
		}
		time.Sleep(100 * time.Millisecond) // Wait before retrying
	}

	if !serverReady {
		// Check channel one last time before failing
		select {
		case err := <-serverErrChan:
			if err != nil {
				t.Fatalf("Server did not become ready within %v. Pylearn script error: %v", maxWait, err)
			}
		default:
			// No error from channel either
		}
		t.Fatalf("Server did not become ready at %s within %v", serverAddr, maxWait)
	}

	// --- Run HTTP Client Tests ---
	client := &http.Client{Timeout: 3 * time.Second}

	t.Run("Server GET /hello", func(t *testing.T) {
		req, _ := http.NewRequest("GET", serverAddr+"/hello", nil)
		req.Header.Set("Accept-Encoding", "test-gzip")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP GET /hello failed: %v", err)
		}
		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /hello, got %d. Body: %s", resp.StatusCode, bodyStr)
		}
		expectedBodyPart := "<h1>Hello Server!</h1> Header: test-gzip"
		if !strings.Contains(bodyStr, expectedBodyPart) {
			t.Errorf("Response body for /hello mismatch.\nWant contains: %q\nGot: %q", expectedBodyPart, bodyStr)
		}
	})

	t.Run("Server POST /data", func(t *testing.T) {
		postData := "client test data"
		resp, err := client.Post(serverAddr+"/data", "text/plain", strings.NewReader(postData))
		if err != nil {
			t.Fatalf("HTTP POST /data failed: %v", err)
		}
		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for POST /data, got %d. Body: %s", resp.StatusCode, bodyStr)
		}
		expectedBody := "Received POST data: " + postData
		if bodyStr != expectedBody {
			t.Errorf("Response body for POST /data mismatch.\nWant: %q\nGot: %q", expectedBody, bodyStr)
		}
	})

	t.Run("Server GET /notfound", func(t *testing.T) {
		resp, err := client.Get(serverAddr + "/otherpath")
		if err != nil {
			t.Fatalf("HTTP GET /otherpath failed: %v", err)
		}
		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		// Assuming the simple handler returns 200 OK even for "not found"
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for /otherpath, got %d. Body: %s", resp.StatusCode, bodyStr)
		}
		expectedBody := "Resource Not Found"
		if bodyStr != expectedBody {
			t.Errorf("Response body for /otherpath mismatch.\nWant: %q\nGot: %q", expectedBody, bodyStr)
		}
	})

	// --- TODO: Server Shutdown ---
	// Need a mechanism to gracefully stop the server started by the Pylearn script.
	// This could involve:
	// 1. The Pylearn `httpserver` module providing a `shutdown()` function.
	// 2. The Go test sending a signal (e.g., via another HTTP request like /shutdown).
	// 3. Having `httpserver.serve` return a server handle that Go can use.
	// Without shutdown, the server process might linger after tests.
	t.Log("NOTE: Server shutdown mechanism not implemented in this test.")
}