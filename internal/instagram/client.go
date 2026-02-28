package instagram

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/netindev/instapicker/internal/browser"
	"github.com/tebeka/selenium"
)

const loginURL = "https://www.instagram.com/accounts/login/"

type Client struct {
	browser *browser.Browser
}

func NewClient(b *browser.Browser) *Client {
	return &Client{browser: b}
}

func (c *Client) Login(user, pass string) error {
	if err := c.browser.Get(loginURL); err != nil {
		return fmt.Errorf("failed to open login page: %w", err)
	}
	time.Sleep(3 * time.Second)

	emailField, err := c.browser.WaitForElement(selenium.ByXPATH, `//*[@name="email"]`, 10*time.Second)
	if err != nil {
		return fmt.Errorf("email field not found: %w", err)
	}
	if err := emailField.Clear(); err != nil {
		return fmt.Errorf("failed to clear email field: %w", err)
	}
	if err := emailField.SendKeys(user); err != nil {
		return fmt.Errorf("failed to fill email field: %w", err)
	}

	passField, err := c.browser.WaitForElement(selenium.ByXPATH, `//*[@type="password"]`, 10*time.Second)
	if err != nil {
		return fmt.Errorf("password field not found: %w", err)
	}
	if err := passField.Clear(); err != nil {
		return fmt.Errorf("failed to clear password field: %w", err)
	}
	if err := passField.SendKeys(pass); err != nil {
		return fmt.Errorf("failed to fill password field: %w", err)
	}

	if err := passField.SendKeys(selenium.EnterKey); err != nil {
		return fmt.Errorf("failed to submit login: %w", err)
	}

	time.Sleep(5 * time.Second)

	codeField, err := c.browser.WaitForElement(selenium.ByXPATH, `//*[@maxlength="8"]`, 5*time.Second)
	if err == nil {
		fmt.Print("verification code required. enter code: ")
		reader := bufio.NewReader(os.Stdin)
		code, _ := reader.ReadString('\n')
		code = strings.TrimSpace(code)

		if err := codeField.Clear(); err != nil {
			return fmt.Errorf("failed to clear code field: %w", err)
		}
		if err := codeField.SendKeys(code); err != nil {
			return fmt.Errorf("failed to fill code field: %w", err)
		}
		if err := codeField.SendKeys(selenium.EnterKey); err != nil {
			return fmt.Errorf("failed to submit code: %w", err)
		}

		time.Sleep(15 * time.Second)
	}

	fmt.Println("waiting for redirect to home page...")
	if err := c.browser.WaitForURL("https://www.instagram.com/accounts/onetap/", 15*time.Second); err != nil {
		return fmt.Errorf("failed to redirect after login: %w", err)
	}
	fmt.Println("login successful.")

	return nil
}

func (c *Client) GetComments(url string) ([]Comment, error) {
	fmt.Printf("navigating to post: %s\n", url)

	if err := c.browser.Get(url); err != nil {
		return nil, fmt.Errorf("failed to open post: %w", err)
	}

	time.Sleep(5 * time.Second)

	commentsSection, err := c.browser.WaitForElement(
		selenium.ByXPATH,
		`(//hr/following-sibling::*)[1]`,
		10*time.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("comments section not found: %w", err)
	}

	if err := c.scrollUntilStable(commentsSection); err != nil {
		return nil, err
	}

	commentDivs, err := c.browser.FindElements(
		selenium.ByXPATH,
		`(//hr/following-sibling::*)[1]/div/div/div`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find comment elements: %w", err)
	}

	comments := make([]Comment, 0, len(commentDivs))

	for _, div := range commentDivs {
		comment, ok := c.parseComment(div)
		if ok {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

func (c *Client) scrollUntilStable(elem selenium.WebElement) error {
	prevCount := 0
	stableRounds := 0

	for {
		_, err := c.browser.ExecuteScript(
			"arguments[0].scrollTop = arguments[0].scrollHeight",
			[]interface{}{elem},
		)
		if err != nil {
			return fmt.Errorf("failed to scroll comments: %w", err)
		}

		time.Sleep(5 * time.Second)

		commentDivs, _ := c.browser.FindElements(
			selenium.ByXPATH,
			`(//hr/following-sibling::*)[1]/div/div/div`,
		)
		currentCount := len(commentDivs)
		fmt.Printf("loaded %d comments...\n", currentCount)

		if currentCount == prevCount {
			stableRounds++
			if stableRounds >= 3 {
				break
			}
		} else {
			stableRounds = 0
		}

		prevCount = currentCount
	}

	return nil
}

func (c *Client) parseComment(div selenium.WebElement) (Comment, bool) {
	username, ok := c.extractUsername(div)
	if !ok {
		return Comment{}, false
	}

	text, ok := c.extractText(div)
	if !ok {
		return Comment{}, false
	}

	profilePicURL, profilePicPath := c.extractAndDownloadProfilePic(div, username)

	return Comment{
		User:               username,
		Text:               text,
		ProfilePictureURL:  profilePicURL,
		ProfilePicturePath: profilePicPath,
	}, true
}

func (c *Client) extractUsername(div selenium.WebElement) (string, bool) {
	elem, err := c.browser.ChildElement(div, selenium.ByXPATH, `.//span[contains(@class, '_ap3a')]`)
	if err != nil {
		return "", false
	}

	username, err := elem.Text()
	if err != nil || username == "" {
		return "", false
	}

	return username, true
}

func (c *Client) extractText(div selenium.WebElement) (string, bool) {
	elems, err := c.browser.ChildElements(div, selenium.ByXPATH, `.//div/div[2]//div/div/div/div[2]//span`)
	if err != nil || len(elems) == 0 {
		return "", false
	}

	text, err := elems[0].Text()
	if err != nil || text == "" {
		return "", false
	}

	return text, true
}

func (c *Client) extractAndDownloadProfilePic(div selenium.WebElement, username string) (string, string) {
	elem, err := c.browser.ChildElement(div, selenium.ByXPATH, ".//img")
	if err != nil || elem == nil {
		return "", ""
	}

	url, _ := elem.GetAttribute("src")
	if url == "" {
		return "", ""
	}

	path, err := downloadProfilePic(username, url)
	if err != nil {
		return url, ""
	}

	return url, path
}

func downloadProfilePic(username, url string) (string, error) {
	safeName := sanitizeUsername(username)
	filename := safeName + ".jpg"

	dir := filepath.Join("..", "..", "result", "pictures")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fullpath := filepath.Join(dir, filename)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	f, err := os.Create(fullpath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	return filepath.Join("pictures", filename), nil
}

func sanitizeUsername(username string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	safe := re.ReplaceAllString(strings.ToLower(username), "")
	if safe == "" {
		return fmt.Sprintf("user_%d", time.Now().UnixNano())
	}
	return safe
}
