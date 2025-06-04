package cs

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type Browser struct{}

func NewBrowser() *Browser {
	return &Browser{}
}

func (b *Browser) OpenIde(path string) error {
	re := regexp.MustCompile(`/api`)
	ideUrl := re.ReplaceAllString(NewEnv().GetApiUrl(), "/ide")
	if !strings.HasPrefix(path, "/") {
		ideUrl += "/"
	}
	url := ideUrl + path

	fmt.Printf("Opening %s in web browser...\n", url)

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Run()
	case "linux":
		err = exec.Command("xdg-open", url).Run()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	default:
		return fmt.Errorf("platform not supported: %s", runtime.GOOS)
	}
	if err != nil {
		return fmt.Errorf("failed to open web browser: %w", err)
	}
	return nil
}
