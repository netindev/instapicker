package browser

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const chromeDriverPort = 9515

type Browser struct {
	wd      selenium.WebDriver
	service *selenium.Service
}

func Start(headless bool) (*Browser, error) {
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService("chromedriver", chromeDriverPort, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start chromedriver: %w", err)
	}

	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
		},
	}
	if headless {
		chromeCaps.Args = append(chromeCaps.Args, "--headless")
	}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", chromeDriverPort))
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("failed to connect to chromedriver: %w", err)
	}

	return &Browser{wd: wd, service: service}, nil
}

func (b *Browser) Get(url string) error {
	if b.wd == nil {
		return fmt.Errorf("webdriver not initialized")
	}
	return b.wd.Get(url)
}

func (b *Browser) Close() {
	if b.wd != nil {
		b.wd.Quit()
	}
	if b.service != nil {
		b.service.Stop()
	}
}

func (b *Browser) CurrentURL() (string, error) {
	if b.wd == nil {
		return "", fmt.Errorf("webdriver not initialized")
	}
	return b.wd.CurrentURL()
}

func (b *Browser) WaitForURL(prefix string, timeout time.Duration) error {
	if b.wd == nil {
		return fmt.Errorf("webdriver not initialized")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		currentURL, err := b.wd.CurrentURL()
		if err == nil && len(currentURL) >= len(prefix) && currentURL[:len(prefix)] == prefix {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("URL did not start with %q after %v", prefix, timeout)
}

func (b *Browser) FindElement(by, value string) (selenium.WebElement, error) {
	if b.wd == nil {
		return nil, fmt.Errorf("webdriver not initialized")
	}
	return b.wd.FindElement(by, value)
}

func (b *Browser) FindElements(by, value string) ([]selenium.WebElement, error) {
	if b.wd == nil {
		return nil, fmt.Errorf("webdriver not initialized")
	}
	return b.wd.FindElements(by, value)
}

func (b *Browser) WaitForElement(by, value string, timeout time.Duration) (selenium.WebElement, error) {
	if b.wd == nil {
		return nil, fmt.Errorf("webdriver not initialized")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		elem, err := b.wd.FindElement(by, value)
		if err == nil {
			return elem, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("element not found [%s=%s] after %v", by, value, timeout)
}

func (b *Browser) ExecuteScript(script string, args []interface{}) (interface{}, error) {
	if b.wd == nil {
		return nil, fmt.Errorf("webdriver not initialized")
	}
	return b.wd.ExecuteScript(script, args)
}

func (b *Browser) ChildElements(parent selenium.WebElement, by, value string) ([]selenium.WebElement, error) {
	return parent.FindElements(by, value)
}

func (b *Browser) ChildElement(parent selenium.WebElement, by, value string) (selenium.WebElement, error) {
	return parent.FindElement(by, value)
}
