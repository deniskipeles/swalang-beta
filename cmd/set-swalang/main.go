package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("===========================================")
	fmt.Println("🌟 Swalang Environment Setup Utility 🌟")
	fmt.Println("===========================================")
	fmt.Println("1. Set/Update SWALANG_PATH (and add to System PATH)")
	fmt.Println("2. Remove SWALANG_PATH")
	fmt.Println("3. Exit")
	fmt.Print("\nSelect an option (1-3): ")

	opt, _ := reader.ReadString('\n')
	opt = strings.TrimSpace(opt)

	switch opt {
	case "1":
		setupPath(reader)
	case "2":
		removePath()
	case "3":
		fmt.Println("Exiting...")
		os.Exit(0)
	default:
		fmt.Println("Invalid option.")
	}

	fmt.Println("\nPress Enter to exit...")
	reader.ReadString('\n')
}

func setupPath(reader *bufio.Reader) {
	currentDir, _ := os.Getwd()
	fmt.Printf("\nEnter the absolute path to your Swalang 'bin' folder.\n[Leave blank to use current directory: %s]: ", currentDir)
	
	inputPath, _ := reader.ReadString('\n')
	inputPath = strings.TrimSpace(inputPath)

	targetPath := currentDir
	if inputPath != "" {
		targetPath = inputPath
	}

	targetPath, _ = filepath.Abs(targetPath)

	// Verify the executable exists in the target path
	exeName := "swalang"
	if runtime.GOOS == "windows" {
		exeName = "swalang.exe"
	}

	if _, err := os.Stat(filepath.Join(targetPath, exeName)); os.IsNotExist(err) {
		fmt.Printf("⚠️  Warning: Could not find '%s' in %s.\n", exeName, targetPath)
		fmt.Print("Are you sure you want to set this path? (y/n): ")
		ans, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(ans)) != "y" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	if runtime.GOOS == "windows" {
		setWindowsEnv(targetPath)
	} else {
		setUnixEnv(targetPath)
	}
}

func removePath() {
	if runtime.GOOS == "windows" {
		removeWindowsEnv()
	} else {
		removeUnixEnv()
	}
}

// --- Windows Implementation ---

func setWindowsEnv(targetPath string) {
	fmt.Println("\nSetting SWALANG_PATH via setx...")
	cmd := exec.Command("setx", "SWALANG_PATH", targetPath)
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Failed to set SWALANG_PATH. Try running as Administrator.")
		return
	}

	// Add to PATH
	fmt.Println("Adding to User PATH...")
	pathCmd := exec.Command("powershell", "-NoProfile", "-Command", 
		`$p = [Environment]::GetEnvironmentVariable("PATH", "User"); if ($p -notmatch [regex]::Escape("`+targetPath+`")) { [Environment]::SetEnvironmentVariable("PATH", $p + ";`+targetPath+`", "User") }`)
	
	if err := pathCmd.Run(); err != nil {
		fmt.Println("⚠️  SWALANG_PATH was set, but failed to append to standard PATH.")
	} else {
		fmt.Println("✅ Success! You may need to restart your terminal or VSCode for changes to take effect.")
	}
}

func removeWindowsEnv() {
	fmt.Println("\nRemoving SWALANG_PATH from registry...")
	cmd := exec.Command("REG", "delete", "HKCU\\Environment", "/F", "/V", "SWALANG_PATH")
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Failed to remove SWALANG_PATH (it might not exist).")
	} else {
		fmt.Println("✅ Successfully removed SWALANG_PATH.")
	}
}

// --- Unix (Linux/macOS) Implementation ---

func setUnixEnv(targetPath string) {
	home, _ := os.UserHomeDir()
	rcFiles := []string{".bashrc", ".zshrc", ".profile"}
	
	success := false
	for _, file := range rcFiles {
		rcPath := filepath.Join(home, file)
		if _, err := os.Stat(rcPath); err == nil {
			appendUnixEnv(rcPath, targetPath)
			success = true
		}
	}

	if success {
		fmt.Println("✅ Success! SWALANG_PATH added to your shell profiles.")
		fmt.Println("Run 'source ~/.bashrc' (or your respective config) or restart your terminal.")
	} else {
		fmt.Println("❌ Could not find .bashrc, .zshrc, or .profile in your home directory.")
	}
}

func appendUnixEnv(rcPath, targetPath string) {
	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString("\n# Swalang Environment Variables\n")
	f.WriteString(fmt.Sprintf("export SWALANG_PATH=\"%s\"\n", targetPath))
	f.WriteString("export PATH=\"$SWALANG_PATH:$PATH\"\n")
}

func removeUnixEnv() {
	fmt.Println("\n⚠️  Automatic removal on Unix is complex. Please manually remove the Swalang lines from your ~/.bashrc or ~/.zshrc.")
}