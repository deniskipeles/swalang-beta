package tests

import (
	"fmt"
	"io/ioutil" // Keep for temp dir/file operations
	"os"
	"path/filepath"
	"testing"

	"github.com/deniskipeles/pylearn/internal/testhelpers"
	"github.com/deniskipeles/pylearn/internal/object" // Keep for NULL comparison

	// Ensure stdlib modules are registered via anonymous imports in a central place or test main
	_ "github.com/deniskipeles/pylearn/internal/stdlib/pyos"
)

// No need for duplicated testEvalStdlib or assertion helpers

func TestOsBuiltinModule(t *testing.T) {

	t.Run("os.getcwd", func(t *testing.T) {
		input := `import os; os.getcwd()`
		evaluated := testhelpers.Eval(t, input)
		goCwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Test setup error: could not get Go CWD: %v", err)
		}
		testhelpers.TestStringObject(t, evaluated, goCwd)
	})

	t.Run("os.getenv", func(t *testing.T) {
		testKey := "PYLEARN_TEST_VAR_GETENV"
		testVal := "pylearn_test_value_getenv"

		// Set env var using Go and ensure cleanup with t.Cleanup
		originalVal, originalValExists := os.LookupEnv(testKey)
		if err := os.Setenv(testKey, testVal); err != nil {
			t.Fatalf("Failed to set env var for test: %v", err)
		}
		t.Cleanup(func() {
			var errRestore error
			if originalValExists {
				errRestore = os.Setenv(testKey, originalVal)
			} else {
				errRestore = os.Unsetenv(testKey)
			}
			if errRestore != nil {
				// Log cleanup errors, don't fail test for cleanup issues
				t.Logf("Warning: failed to restore env var %s during cleanup: %v", testKey, errRestore)
			}
		})

		// Test getting existing var
		input1 := fmt.Sprintf(`import os; os.getenv(%q)`, testKey)
		eval1 := testhelpers.Eval(t, input1)
		testhelpers.TestStringObject(t, eval1, testVal)

		// Test getting non-existent var (should return None)
		input2 := `import os; os.getenv("NON_EXISTENT_VAR_XYZ_GO_GETENV")`
		eval2 := testhelpers.Eval(t, input2)
		testhelpers.TestNullObject(t, eval2) // Check for NULL singleton

		// Test getting non-existent var with default
		input3 := `import os; os.getenv("NON_EXISTENT_VAR_XYZ_GETENV", "default_val")`
		eval3 := testhelpers.Eval(t, input3)
		testhelpers.TestStringObject(t, eval3, "default_val")

		// Test getting non-existent var with non-string default
		input4 := `import os; os.getenv("NON_EXISTENT_VAR_XYZ_GETENV", 123)`
		eval4 := testhelpers.Eval(t, input4)
		testhelpers.TestIntegerObject(t, eval4, 123) // Default is returned as-is

		// Error cases
		errTests := []struct{ input string; errParts []string }{
			{`import os; os.getenv()`, []string{"TypeError", "takes 1 or 2 arguments", "0 given"}},
			{`import os; os.getenv(123)`, []string{"TypeError", "argument 1 must be str"}},
			{`import os; os.getenv("K", "D", "TooMany")`, []string{"TypeError", "takes 1 or 2 arguments", "3 given"}},
		}
		for _, et := range errTests {
			t.Run(et.input+" (error)", func(t *testing.T){
				evalErr := testhelpers.Eval(t, et.input)
				testhelpers.TestErrorObject(t, evalErr, et.errParts...)
			})
		}
	})

	t.Run("os.listdir", func(t *testing.T) {
		// Create temp dir using Go and ensure cleanup with t.Cleanup
		tempDir, err := ioutil.TempDir("", "pylearntest_listdir_*")
		if err != nil {
			t.Fatalf("Failed to create temp dir for test: %v", err)
		}
		t.Cleanup(func() {
			errRemove := os.RemoveAll(tempDir)
			if errRemove != nil {
				t.Logf("Warning: failed to remove temp dir %s during cleanup: %v", tempDir, errRemove)
			}
		})

		// Create items inside the temp dir
		testFile1 := filepath.Join(tempDir, "file1.txt")
		testFile2 := filepath.Join(tempDir, "another.dat")
		testSubDir := filepath.Join(tempDir, "subdir")
		if err := ioutil.WriteFile(testFile1, []byte("test1"), 0644); err != nil { t.Fatal(err) }
		if err := ioutil.WriteFile(testFile2, []byte("test2"), 0644); err != nil { t.Fatal(err) }
		if err := os.Mkdir(testSubDir, 0755); err != nil { t.Fatal(err) }

		// Test listing the specific temp directory
		input1 := fmt.Sprintf(`import os; os.listdir(%q)`, tempDir)
		eval1 := testhelpers.Eval(t, input1)

		// Check result is a list
		listObj1, ok := eval1.(*object.List)
		if !ok {
			t.Fatalf("os.listdir did not return a List. got %T: %s", eval1, eval1.Inspect())
		}

		// Check contents (order doesn't matter)
		expectedItems := map[string]bool{"file1.txt": true, "another.dat": true, "subdir": true}
		if len(listObj1.Elements) != len(expectedItems) {
			t.Fatalf("os.listdir length mismatch. Got %d, expected %d. Items: %s", len(listObj1.Elements), len(expectedItems), listObj1.Inspect())
		}
		listedItems := make(map[string]bool)
		for _, item := range listObj1.Elements {
			itemStr, ok := item.(*object.String)
			if !ok {
				t.Errorf("os.listdir item is not a string: %s", item.Inspect())
				continue
			}
			listedItems[itemStr.Value] = true
		}
		for name := range expectedItems {
			if !listedItems[name] {
				t.Errorf("os.listdir result missing expected item: %s. Got list items: %v", name, listedItems)
			}
		}

		// Test listing default dir ('.')
		input2 := `import os; os.listdir()`
		eval2 := testhelpers.Eval(t, input2)
		if _, ok2 := eval2.(*object.List); !ok2 {
			t.Fatalf("os.listdir() did not return a List. got %T: %s", eval2, eval2.Inspect())
		}
		// Can't easily assert contents of '.', just check type

		// Error cases
		errTests := []struct{ input string; errParts []string }{
			{`import os; os.listdir("no_such_directory_here_xyz_listdir")`, []string{"OSError", "no such file or directory"}}, // Or FileNotFoundError
			{`import os; os.listdir(123)`, []string{"TypeError", "argument must be str"}},
			{`import os; os.listdir(".", "extra")`, []string{"TypeError", "takes at most 1 argument", "2 given"}},
		}
        for _, et := range errTests {
            t.Run(et.input+" (error)", func(t *testing.T){
                evalErr := testhelpers.Eval(t, et.input)
				testhelpers.TestErrorObject(t, evalErr, et.errParts...)
            })
        }
	})
}