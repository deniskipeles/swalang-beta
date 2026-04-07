// pylearn/internal/package_manager/lockfile.go
package package_manager

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

const lockfileName = "pylearn.lock"

var lockfileMutex sync.Mutex

// Package represents a single entry in our lockfile.
type Package struct {
	Name string `json:"name"` // e.g., "pylearn-requests"
	URL  string `json:"url"`  // e.g., "https://github.com/user/pylearn-requests"
	// We could add version/commit hash here in the future
}

// Lockfile represents the structure of the pylearn.lock file.
type Lockfile struct {
	Packages []Package `json:"packages"`
}

// readLockfile reads the existing lockfile or returns a new empty one.
func readLockfile() (*Lockfile, error) {
	lockfileMutex.Lock()
	defer lockfileMutex.Unlock()

	data, err := ioutil.ReadFile(lockfileName)
	if err != nil {
		// If the file doesn't exist, that's okay. Return a new empty one.
		if os.IsNotExist(err) {
			return &Lockfile{Packages: []Package{}}, nil
		}
		// For any other error (e.g., permissions), return it.
		return nil, err
	}

	var lockfile Lockfile
	if err := json.Unmarshal(data, &lockfile); err != nil {
		return nil, err
	}
	return &lockfile, nil
}

// writeLockfile saves the lockfile data back to disk.
func writeLockfile(lockfile *Lockfile) error {
	lockfileMutex.Lock()
	defer lockfileMutex.Unlock()

	data, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(lockfileName, data, 0644)
}

// isPackageInstalled checks if a package is already listed in the lockfile.
func (lf *Lockfile) isPackageInstalled(repoName string) bool {
	for _, pkg := range lf.Packages {
		if pkg.Name == repoName {
			return true
		}
	}
	return false
}

// addPackage adds a new package to the lockfile (in memory).
func (lf *Lockfile) addPackage(pkg Package) {
	// Avoid duplicates
	if lf.isPackageInstalled(pkg.Name) {
		return
	}
	lf.Packages = append(lf.Packages, pkg)
}
