// pylearn/internal/package_manager/get_packages.go

package package_manager

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// --- ADD THIS NEW FUNCTION for the package manager ---

// handleGetCommand processes the `pylearn get <url>` command.
func HandleGetCommand(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Package manager 'get' error: Missing repository URL.")
		fmt.Fprintln(os.Stderr, "Usage: pylearn get <https://github.com/user/repo>")
		os.Exit(1)
	}
	repoURL := args[0]

	// 1. Parse Repo Name from URL
	// e.g., "https://github.com/user/pylearn-requests" -> "pylearn-requests"
	cleanURL := strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(cleanURL, "/")
	if len(parts) < 2 {
		fmt.Fprintf(os.Stderr, "Error: Invalid repository URL format: %s\n", repoURL)
		os.Exit(1)
	}
	repoName := parts[len(parts)-1]

	// 2. Check Lockfile
	lockfile, err := readLockfile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not read lockfile: %v\n", err)
		os.Exit(1)
	}
	if lockfile.isPackageInstalled(repoName) {
		fmt.Printf("--> Package '%s' is already installed. Skipping.\n", repoName)
		return
	}

	// 3. Construct Download URL and Fetch
	zipURL := cleanURL + "/archive/refs/heads/main.zip"
	fmt.Printf("--> Fetching package '%s' from %s\n", repoName, zipURL)
	resp, err := http.Get(zipURL)
	if err != nil { /* ... handle error ... */
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { /* ... handle error ... */
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil { /* ... handle error ... */
	}

	// 4. Unzip into the CORRECT subdirectory
	modulesDir := "modules"
	// The destination is now `modules/reponame/`
	packageDestDir := filepath.Join(modulesDir, repoName)
	fmt.Printf("--> Installing package into '%s/'\n", packageDestDir)

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil { /* ... handle error ... */
	}
	if len(zipReader.File) == 0 { /* ... handle error ... */
	}

	rootZipFolder := zipReader.File[0].Name // e.g., "pylearn-requests-main/"

	for _, f := range zipReader.File {
		// Strip the top-level folder from the zip file path to get the relative path
		relativePath, err := filepath.Rel(rootZipFolder, f.Name)
		if err != nil {
			// This can happen for the root folder itself, which we can ignore.
			continue
		}

		destPath := filepath.Join(packageDestDir, relativePath)

		if !strings.HasPrefix(destPath, filepath.Clean(packageDestDir)+string(os.PathSeparator)) {
			fmt.Fprintf(os.Stderr, "Error: Invalid file path in zip: %s\n", destPath)
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory for %s: %v\n", destPath, err)
			continue
		}
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil { /* ... handle error, continue ... */
		}
		zippedFile, err := f.Open()
		if err != nil { /* ... handle error, continue ... */
		}
		if _, err := io.Copy(destFile, zippedFile); err != nil { /* ... handle error ... */
		}
		destFile.Close()
		zippedFile.Close()
	}

	// 5. Update and Write Lockfile
	lockfile.addPackage(Package{Name: repoName, URL: repoURL})
	if err := writeLockfile(lockfile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to write lockfile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("--> Package '%s' installed successfully.\n", repoName)
}
