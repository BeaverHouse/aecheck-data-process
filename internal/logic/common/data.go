package common

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const debugPort = "9222"

// Gets data from a URL.
//
// If the URL is invalid or the data is empty, it returns an error.
func GetDataFromURL(url string) ([]byte, error) {
	if strings.Contains(url, "anothereden.wiki") {
		return getDataWithBrowser(url)
	}
	return getDataWithHTTP(url)
}

// findInstalledBrowser returns the path to an installed Chrome/Edge on the system.
// Falls back to whatever launcher.LookPath finds if none of the known locations match.
func findInstalledBrowser() (string, bool) {
	candidates := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		env.GetEnv("LOCALAPPDATA", "") + `\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return launcher.LookPath()
}

// Launches a user-controlled Chrome/Edge with a dedicated profile and remote debugging
// enabled, asks the user to solve any Cloudflare challenge, then attaches via CDP and
// returns the current page HTML.
func getDataWithBrowser(url string) ([]byte, error) {
	start := time.Now()

	path, found := findInstalledBrowser()
	if !found {
		return nil, fmt.Errorf("GetDataFromURL: no installed browser found (Chrome/Edge required)")
	}

	profileDir := filepath.Join(os.TempDir(), "aecheck-browser-profile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return nil, fmt.Errorf("GetDataFromURL: create profile dir: %w", err)
	}

	cmd := exec.Command(path,
		"--remote-debugging-port="+debugPort,
		"--user-data-dir="+profileDir,
		"--no-first-run",
		"--no-default-browser-check",
		url,
	)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("GetDataFromURL: launch browser: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	controlURL, err := waitForDebugEndpoint()
	if err != nil {
		return nil, fmt.Errorf("GetDataFromURL: %w", err)
	}

	fmt.Printf("\n[browser] Opened: %s\n", url)
	fmt.Println("[browser] Solve any Cloudflare challenge in the browser, then press Enter to continue...")
	if err := waitForBrowserConfirmation(); err != nil {
		return nil, fmt.Errorf("GetDataFromURL: waiting for browser confirmation: %w", err)
	}

	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	pages, err := browser.Pages()
	if err != nil {
		return nil, fmt.Errorf("GetDataFromURL: list pages: %w", err)
	}
	var target *rod.Page
	for _, p := range pages {
		info, err := p.Info()
		if err != nil {
			continue
		}
		if isTargetWikiPage(info.URL, url) {
			target = p
			break
		}
	}
	if target == nil {
		for _, p := range pages {
			info, err := p.Info()
			if err != nil {
				continue
			}
			if strings.Contains(info.URL, "anothereden.wiki") {
				target = p
				break
			}
		}
	}
	if target == nil && len(pages) > 0 {
		target = pages[0]
	}
	if target == nil {
		return nil, fmt.Errorf("GetDataFromURL: no page found in browser")
	}
	target.MustWaitLoad()

	html, err := target.HTML()
	if err != nil {
		return nil, fmt.Errorf("GetDataFromURL: browser: %w", err)
	}

	body := []byte(html)
	if len(body) == 0 {
		return nil, fmt.Errorf("GetDataFromURL: the data from URL is empty: %s", url)
	}
	Log.Info("Response successfully fetched (browser)", logger.Field{Key: "url", Value: url}, logger.Field{Key: "duration", Value: time.Since(start)})
	return body, nil
}

func isTargetWikiPage(pageURL string, targetURL string) bool {
	return strings.TrimRight(pageURL, "/") == strings.TrimRight(targetURL, "/")
}

func waitForBrowserConfirmation() error {
	input := io.Reader(os.Stdin)
	if console, err := os.OpenFile("CONIN$", os.O_RDONLY, 0); err == nil {
		defer console.Close()
		input = console
	}

	_, err := bufio.NewReader(input).ReadString('\n')
	return err
}

func waitForDebugEndpoint() (string, error) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:" + debugPort + "/json/version")
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			idx := strings.Index(string(body), `"webSocketDebuggerUrl": "`)
			if idx >= 0 {
				rest := string(body)[idx+len(`"webSocketDebuggerUrl": "`):]
				end := strings.Index(rest, `"`)
				if end > 0 {
					return rest[:end], nil
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for browser debug endpoint")
}

func getDataWithHTTP(url string) ([]byte, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GetDataFromURL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GetDataFromURL: invalid status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetDataFromURL: %w", err)
	} else if len(body) == 0 {
		return nil, fmt.Errorf("GetDataFromURL: the data from URL is empty: %s", url)
	}
	Log.Info("Response successfully fetched", logger.Field{Key: "url", Value: url}, logger.Field{Key: "duration", Value: time.Since(start)})
	return body, nil
}
