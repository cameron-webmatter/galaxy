package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/galaxy/galaxy/pkg/config"
	"github.com/galaxy/galaxy/pkg/server"
	"github.com/spf13/cobra"
)

var (
	devPort    int
	devHost    string
	devOpen    bool
	devVerbose bool
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start the development server",
	Long:  `Start the development server with hot reload`,
	RunE:  runDev,
}

func init() {
	rootCmd.AddCommand(devCmd)
	devCmd.Flags().IntVar(&devPort, "port", 4322, "port to run server on")
	devCmd.Flags().StringVar(&devHost, "host", "localhost", "host to bind to")
	devCmd.Flags().BoolVar(&devOpen, "open", false, "open browser on start")
	devCmd.Flags().BoolVar(&devVerbose, "verbose", true, "enable request logging")
}

func runDev(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if rootDir != "" {
		cwd = rootDir
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	srcDir := cfg.SrcDir
	if !filepath.IsAbs(srcDir) {
		srcDir = filepath.Join(cwd, srcDir)
	}

	pagesDir := filepath.Join(srcDir, "pages")
	publicDir := filepath.Join(cwd, "public")

	if _, err = os.Stat(pagesDir); os.IsNotExist(err) {
		return fmt.Errorf("pages directory not found: %s", pagesDir)
	}

	srv := server.NewDevServer(cwd, pagesDir, publicDir, devPort, devVerbose)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(srcDir); err != nil {
		return err
	}
	if err := addRecursive(watcher, srcDir); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
					if !verbose && !silent {
						fmt.Printf("🔄 Change detected: %s\n", filepath.Base(event.Name))
					}

					if event.Op&fsnotify.Create != 0 {
						info, err := os.Stat(event.Name)
						if err == nil && info.IsDir() && isUnderDir(event.Name, srcDir) {
							if err := watcher.Add(event.Name); err == nil {
								if err := addRecursive(watcher, event.Name); err != nil && !silent {
									fmt.Printf("⚠ Failed to watch new directory: %v\n", err)
								}
							}
						}
					}

					if filepath.Ext(event.Name) == ".gxc" && isUnderDir(event.Name, srcDir) {
						srv.Compiler.ClearCache()

						if isUnderDir(event.Name, pagesDir) && event.Op&(fsnotify.Create|fsnotify.Remove) != 0 {
							if err := srv.ReloadRoutes(); err != nil && !silent {
								fmt.Printf("⚠ Failed to reload routes: %v\n", err)
							}
						}
					}
				}
			case err := <-watcher.Errors:
				if !silent {
					fmt.Printf("⚠ Watcher error: %v\n", err)
				}
			}
		}
	}()

	go handleInput(srv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n👋 Shutting down...")
		if srv.Lifecycle != nil {
			if err := srv.Lifecycle.ExecuteShutdown(); err != nil {
				fmt.Printf("⚠ Shutdown error: %v\n", err)
			}
		}
		os.Exit(0)
	}()

	if devOpen {
		go openBrowser(fmt.Sprintf("http://%s:%d", devHost, devPort))
	}

	fmt.Println("\n⚡ Hotkeys:")
	fmt.Println("  o + enter  →  Open browser")
	fmt.Println("  r + enter  →  Restart server")
	fmt.Println("  c + enter  →  Clear console")
	fmt.Println("  q + enter  →  Quit")

	return srv.Start()
}

func addRecursive(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

func handleInput(srv *server.DevServer) {
	buf := make([]byte, 1)
	for {
		os.Stdin.Read(buf)
		switch buf[0] {
		case 'o':
			openBrowser(fmt.Sprintf("http://localhost:%d", srv.Port))
		case 'r':
			fmt.Println("🔄 Restart requested (not implemented)")
		case 'c':
			clearConsole()
		case 'q':
			fmt.Println("\n👋 Shutting down...")
			os.Exit(0)
		}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func clearConsole() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	}
	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func isUnderDir(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != "."
}
