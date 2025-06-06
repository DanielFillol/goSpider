package goSpider

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/DanielFillol/goSpider/htmlQuery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Navigator is a struct that holds the context for the ChromeDP session and a logger.
type Navigator struct {
	Ctx         context.Context
	Cancel      context.CancelFunc
	Logger      *log.Logger
	DebugLogger bool
	Timeout     time.Duration
	Cookies     []*network.Cookie
	QueryOption chromedp.QueryOption
}

// NewNavigator creates a new Navigator instance.
//
// Parameters:
//   - profilePath: the path to chrome profile defined by the user; can be passed as an empty string
//   - headless: if false will show chrome UI
//
// Example:
//
//	nav := goSpider.NewNavigator("/Users/USER_NAME/Library/Application Support/Google/Chrome/Profile 2", true, initialCookies)
//
// NewNavigator creates a new Navigator instance with enhanced logging for troubleshooting authentication issues.
func NewNavigator(profilePath string, headless bool) *Navigator {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-features", "SameSiteByDefaultCookies,CookiesWithoutSameSiteMustBeSecure"), // Disable SameSite restrictions
		chromedp.Flag("disable-site-isolation-trials", true),                                             // Allow third-party content
		chromedp.Flag("allow-running-insecure-content", true),                                            // Allow mixed content (http & https)
		chromedp.Flag("ignore-certificate-errors", true),                                                 // Ignore certificate errors
		chromedp.Flag("enable-cookies", true),                                                            // Ensure cookies are enabled
	)

	if headless {
		opts = append(opts, chromedp.Headless)
		opts = append(opts, chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"))
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	if profilePath != "" {
		opts = append(opts, chromedp.UserDataDir(profilePath))
	}

	allocCtx, cancelAllocCtx := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	logger := log.New(os.Stdout, "goSpider: ", log.LstdFlags)
	navigator := &Navigator{
		Ctx: ctx,
		Cancel: func() {
			cancelCtx()
			cancelAllocCtx()
		},
		Logger:      logger,
		Cookies:     []*network.Cookie{},
		QueryOption: chromedp.ByQuery,
	}

	// Set standard timeout with enhanced logging
	navigator.SetTimeOut(300 * time.Millisecond)
	if navigator.DebugLogger {
		logger.Printf("Navigator initialized with timeout: %v\n", navigator.Timeout)
	}

	return navigator
}

// SetQueryType defines selector type (CSS ou XPath)
func (nav *Navigator) SetQueryType(queryType chromedp.QueryOption) {
	nav.QueryOption = queryType
}

func (nav *Navigator) UseXPath() {
	nav.SetQueryType(chromedp.BySearch)
}

func (nav *Navigator) UseCSS() {
	nav.SetQueryType(chromedp.ByQuery)
}

// SetTimeOut sets a timeout for all the waiting functions on the package. The standard timeout of the Navigator is 300 ms.
func (nav *Navigator) SetTimeOut(timeOut time.Duration) {
	nav.Timeout = timeOut
}

// GetElementAttribute retrieves the value of a specified attribute from an element identified by a CSS selector.
// Parameters:
// - selector: The CSS selector of the element.
// - attribute: The name of the attribute to retrieve the value of.
// Returns:
// - The value of the specified attribute.
// - An error if the attribute value could not be retrieved.
func (nav *Navigator) GetElementAttribute(selector, attribute string) (string, error) {
	var value string

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return "", err
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.AttributeValue(selector, attribute, &value, nil, nav.QueryOption),
	)
	if err != nil {
		return "", fmt.Errorf("error getting attribute %s: %v", attribute, err)
	}
	return value, nil
}

// SwitchToNewTab returns the Navigator with a new context
func (nav *Navigator) SwitchToNewTab() (*Navigator, error) {
	ctx, cancel := context.WithTimeout(nav.Ctx, nav.Timeout)
	defer cancel()

	// Targets antes do clique
	targetsBefore, err := chromedp.Targets(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting initial targets: %v", err)
	}

	// Esperar breve momento para permitir que nova aba seja criada
	time.Sleep(500 * time.Millisecond)

	// Targets após o clique
	targetsAfter, err := chromedp.Targets(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting targets after click: %v", err)
	}

	var newTabID target.ID
	for _, t := range targetsAfter {
		if t.Type == "page" && !containsTarget(targetsBefore, t.TargetID) {
			newTabID = t.TargetID
			break
		}
	}

	if newTabID == "" {
		return nil, fmt.Errorf("failed to detect new tab: no new target found")
	}

	newCtx, _ := chromedp.NewContext(nav.Ctx, chromedp.WithTargetID(newTabID))

	return &Navigator{
		Ctx:         newCtx,
		Cancel:      func() { chromedp.Cancel(newCtx) },
		Logger:      nav.Logger,
		DebugLogger: nav.DebugLogger,
		Timeout:     nav.Timeout,
		QueryOption: nav.QueryOption,
	}, nil
}
func containsTarget(targets []*target.Info, id target.ID) bool {
	for _, t := range targets {
		if t.TargetID == id {
			return true
		}
	}
	return false
}

// SwitchToFrame switches the context to the specified iframe.
func (nav *Navigator) SwitchToFrame(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Println("Switching to frame", selector)
	}

	// Wait for the iframe to be visible
	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return err
	}

	// Switch to the iframe context by evaluating JavaScript
	err = chromedp.Run(nav.Ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var res interface{}
			err := chromedp.Evaluate(fmt.Sprintf(`
				var iframe = document.querySelector('%s');
				iframe.contentWindow.document.body.innerHTML`, selector), &res).Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to switch to iframe: %v", err)
			}
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to switch to iframe: %v", err)
	}

	if nav.DebugLogger {
		nav.Logger.Println("SwitchToFrame", selector, "successfully")
	}
	return nil
}

// SwitchToDefaultContent switches the context back to the main content from an iframe context.
func (nav *Navigator) SwitchToDefaultContent() error {
	if nav.DebugLogger {
		nav.Logger.Println("Switching to default content")
	}

	// Switch back to the main content
	err := chromedp.Run(nav.Ctx,
		chromedp.Tasks{
			chromedp.ActionFunc(func(ctx context.Context) error {
				// Evaluate JavaScript to switch back to the top window
				err := chromedp.Evaluate(`window.top.location.reload()`, nil).Do(ctx)
				if err != nil {
					return fmt.Errorf("failed to switch to default content: %v", err)
				}
				return nil
			}),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to switch to default content: %v", err)
	}

	if nav.DebugLogger {
		nav.Logger.Println("Switch to default content successfully")
	}
	return nil
}

// CheckPageTitle navigates to the provided URL and checks if the page title equals "Ah, não!".
// It returns true if the error title is detected, otherwise false.
func (nav *Navigator) CheckPageTitle(url string) (bool, error) {
	var title string
	// Run the navigation and title extraction actions.
	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(url),
		chromedp.Title(&title),
	)
	if err != nil {
		return false, fmt.Errorf("failed to navigate or get title: %v", err)
	}

	// Optionally, log the title if DebugLogger is enabled.
	if nav.DebugLogger {
		nav.Logger.Printf("Page title: %s\n", title)
	}

	// Check if the title indicates the error.
	if strings.TrimSpace(title) == "Ah, não!" {
		return true, nil
	}
	return false, nil
}

// OpenURL opens the specified URL in the current browser context.
// It will retry up to 3 times if the page title indicates an error ("Ah, não!").
// Example:
//
//	err := nav.OpenURL("https://www.example.com")
func (nav *Navigator) OpenURL(url string) error {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if nav.DebugLogger {
			nav.Logger.Printf("Attempt %d: Opening URL: %s\n", attempt, url)
		}
		// Navigate to the URL and wait for the page's body to be ready.
		err := chromedp.Run(nav.Ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body"),
		)
		if err != nil {
			return fmt.Errorf("error - failed to open URL: %v", err)
		}

		// Wait for the page load (this method may include additional checks as needed).
		_, err = nav.WaitPageLoad()
		if err != nil {
			return err
		}

		// Check if the page title indicates the error "Ah, não!".
		isError, err := nav.CheckPageTitle(url)
		if err != nil {
			return fmt.Errorf("error checking page title: %v", err)
		}

		if !isError {
			if nav.DebugLogger {
				nav.Logger.Printf("URL opened successfully with URL: %s\n", url)
			}
			return nil
		}

		// Log the retry if error detected
		nav.Logger.Printf("Attempt %d: Detected error in page title. Retrying...\n", attempt)
		// Optionally, wait a bit before retrying
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("failed to open URL %s after %d attempts", url, maxRetries)
}

// GetCurrentURL returns the current URL of the browser.
// Example:
//
//	currentURL, err := nav.GetCurrentURL()
func (nav *Navigator) GetCurrentURL() (string, error) {
	if nav.DebugLogger {
		nav.Logger.Println("Extracting the current URL")
	}
	var currentURL string
	err := chromedp.Run(nav.Ctx,
		chromedp.Location(&currentURL),
	)
	if err != nil {
		return "", fmt.Errorf("error - failed to extract current URL: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Println("Current URL extracted successfully")
	}
	return currentURL, nil
}

// Login logs into a website using the provided credentials and selectors.
// Example:
//
//	err := nav.Login("https://www.example.com/login", "username", "password", "#username", "#password", "#login-button", "#login-message-fail")
func (nav *Navigator) Login(url, username, password, usernameSelector, passwordSelector, loginButtonSelector string, messageFailedSuccess string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Logging into URL: %s\n", url)
	}

	if url != "" {
		err := nav.OpenURL(url)
		if err != nil {
			return fmt.Errorf("error - failed to open URL: %v", err)
		}
	}

	err := nav.WaitForElement(usernameSelector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = nav.WaitForElement(passwordSelector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = nav.WaitForElement(loginButtonSelector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.SendKeys(usernameSelector, username, chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, password, chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.WaitReady("body"), // Wait for the next page to load
	)
	if err != nil {
		if messageFailedSuccess != "" {
			err = nav.WaitForElement(messageFailedSuccess, nav.Timeout)
			if err != nil {
				return fmt.Errorf("error - failed waiting for element: %v", err)
			}
			message, err := nav.GetElement(messageFailedSuccess)
			if err == nil {
				if nav.DebugLogger {
					nav.Logger.Printf("Message found: %s", message)
				}
				return fmt.Errorf("error - message: %v", message)
			} else {
				return fmt.Errorf("error - failed to log in: %v", err)
			}
		}
		if nav.DebugLogger {
			nav.Logger.Printf("Error - Failed to log in: %v\n", err)
		}
		return fmt.Errorf("error - failed to log in: %v", err)
	}

	//sometimes the page does accept the login information but still returns an error message
	if messageFailedSuccess != "" {
		err = nav.WaitForElement(messageFailedSuccess, nav.Timeout)
		if err == nil {
			message, err := nav.GetElement(messageFailedSuccess)
			if err == nil {
				if nav.DebugLogger {
					nav.Logger.Printf("Message found: %s", message)
				}
				return fmt.Errorf("error - message: %v", message)
			} else {
				return fmt.Errorf("error - failed to log in: %v", err)
			}
		}
	}

	if nav.DebugLogger {
		nav.Logger.Println("Logged in successfully")
	}
	return nil
}

// LoginAccountsGoogle performs the Google login on the given URL
func (nav *Navigator) LoginAccountsGoogle(email, password string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Opening URL: %s\n", "accounts.google.com")
	}
	err := chromedp.Run(nav.Ctx, chromedp.Navigate("https://accounts.google.com"))
	if err != nil {
		return fmt.Errorf("failed to open URL: %v", err)
	}

	_, err = nav.WaitPageLoad()
	if err != nil {
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Fill the Google login form
	if nav.DebugLogger {
		nav.Logger.Println("Filling in the Google login form")
	}
	err = nav.FillField(`#identifierId`, email)
	if err != nil {
		err = nav.WaitForElement("#yDmH0d > c-wiz > div > div:nth-child(2) > div > c-wiz > c-wiz > div > div.s7iwrf.gMPiLc.Kdcijb > div > div > header > h1", nav.Timeout)
		if err != nil {
			return fmt.Errorf("error - failed to check login: %v", err)
		} else {
			s, err := nav.GetElement("#yDmH0d > c-wiz > div > div:nth-child(2) > div > c-wiz > c-wiz > div > div.s7iwrf.gMPiLc.Kdcijb > div > div > header > h1")
			if err != nil {
				return fmt.Errorf("error - failed to check login: %v", err)
			}
			if nav.DebugLogger {
				nav.Logger.Printf("Already logged in! \n%s", s)
			}
			return nil
		}
	}

	err = nav.ClickButton(`#identifierNext`)
	if err != nil {
		if nav.DebugLogger {
			nav.Logger.Printf("Failed to click the 'Next' button: %v\n", err)
		}
		return fmt.Errorf("failed to click the 'Next' button: %v", err)
	}

	// Adding a small delay to allow the next page to load
	_, err = nav.WaitPageLoad()
	if err != nil {
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}
	time.Sleep(2 * time.Second)

	err = nav.FillField("#password > div.aCsJod.oJeWuf > div > div.Xb9hP > input", password)
	if err != nil {
		return fmt.Errorf("failed to fill the password field: %v", err)
	}

	err = nav.ClickButton(`#passwordNext`)
	if err != nil {
		return fmt.Errorf("failed to click the 'Next' button for password: %v", err)
	}

	// Adding a small delay to allow the next page to load
	_, err = nav.WaitPageLoad()
	if err != nil {
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}
	time.Sleep(2 * time.Second)

	authCode := AskForString("Google verification pass: ")

	//"#yDmH0d > c-wiz > div > div.UXFQgc > div > div > div > form > span > section:nth-child(2) > div > div > div.AFTWye.GncK > div > div.aCsJod.oJeWuf > div > div.Xb9hP"
	err = nav.FillField("#idvPin", authCode)
	if err != nil {
		return fmt.Errorf("failed to fill the idvPin with code: %s\n field: %v\n", authCode, err)
	}

	if nav.DebugLogger {
		nav.Logger.Println("Google login completed successfully")
	}
	return nil
}

// LoginWithGoogle performs the Google login on the given URL
func (nav *Navigator) LoginWithGoogle(url string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Opening URL: %s\n", url)
	}
	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(url),
	)
	if err != nil {
		return fmt.Errorf("failed to open URL: %v", err)
	}

	if nav.DebugLogger {
		nav.Logger.Println("Clicking the 'Continuar com o Google' button")
	}
	err = nav.ClickButton(".SocialButton")
	if err != nil {
		if nav.DebugLogger {
			nav.Logger.Printf("Alredy logged in: %v\n", err)
		}
		return nil
		//if nav.DebugLogger {nav.Logger.Printf("Failed to click the Google login button: %v\n", err)}
		//return fmt.Errorf("failed to click the Google login button: %v", err)
	}

	// Wait for the popup to appear and switch to it
	if nav.DebugLogger {
		nav.Logger.Println("Switching to the Google login popup")
	}
	var popupCtx context.Context
	var popupCancel context.CancelFunc
	for {
		select {
		case <-time.After(1 * time.Second):
			targets, _ := chromedp.Targets(nav.Ctx)
			if len(targets) > 1 {
				for _, t := range targets {
					if t.Type == "page" && t.TargetID != chromedp.FromContext(nav.Ctx).Target.TargetID {
						popupCtx, popupCancel = chromedp.NewContext(nav.Ctx, chromedp.WithTargetID(targets[1].TargetID))
						break
					}
				}
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("failed to detect the Google login popup")
		}
		if popupCtx != nil {
			break
		}
	}

	// Ensure the popup context is cancelled after use
	defer popupCancel()

	// Create a new logger for the popup context
	popupLogger := log.New(os.Stdout, "popup: ", log.LstdFlags)
	newNav := Navigator{
		Ctx:    popupCtx,
		Cancel: popupCancel,
		Logger: popupLogger,
	}

	// Log the current URL of the popup
	currentURL, err := newNav.GetCurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get the current URL of the popup: %v\n", err)
	}
	fmt.Printf("Popup URL: %s\n", currentURL)

	// Check if the popup navigated to the Google login page
	if !strings.Contains(currentURL, "accounts.google.com") {
		if nav.DebugLogger {
			nav.Logger.Printf("Popup did not navigate to Google login page, current URL: %s\n", currentURL)
		}
		return fmt.Errorf("popup did not navigate to Google login page")
	}

	// Increase the timeout for filling the form fields
	popupCtx, popupCancel = context.WithTimeout(popupCtx, nav.Timeout)
	defer popupCancel()

	// Fill the Google login form
	err = newNav.ClickElement("#container")
	if err != nil {
		return fmt.Errorf("failed to click element: %v", err)
	}

	_, err = newNav.WaitPageLoad()
	if err != nil {
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}

	err = newNav.ClickButton("#credentials-picker > div.fFW7wc-ibnC6b-sM5MNb.TAKBxb")
	if err != nil {
		return fmt.Errorf("failed to click button: %v", err)
	}

	if nav.DebugLogger {
		newNav.Logger.Println("Google login completed successfully")
	}
	return nil
}

// AskForString prompts the user to enter a string and returns the trimmed input.
func AskForString(prompt string) string {
	fmt.Print(prompt)                     // Display the prompt to the user
	reader := bufio.NewReader(os.Stdin)   // Create a new reader for standard input
	input, err := reader.ReadString('\n') // Read the input until a newline character is encountered
	if err != nil {                       // Check if there was an error during input
		fmt.Println("Error reading input:", err) // Print the error message
		return ""                                // Return an empty string in case of an error
	}
	return strings.TrimSpace(input) // Trim any leading/trailing whitespace including the newline character
}

// CaptureScreenshot captures a screenshot of the current browser window.
// Example:
//
//	err := nav.CaptureScreenshot("img")
func (nav *Navigator) CaptureScreenshot(nameFile string) error {
	var buf []byte
	if nav.DebugLogger {
		nav.Logger.Println("Capturing screenshot")
	}
	err := chromedp.Run(nav.Ctx,
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return fmt.Errorf("error - failed to capture screenshot: %v", err)
	}
	err = ioutil.WriteFile(nameFile+"_screenshot.png", buf, 0644)
	if err != nil {
		return fmt.Errorf("error - failed to save screenshot: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Screenshot saved successfully with name: %s\n", nameFile)
	}
	return nil
}

// ReloadPage reloads the current page with retry logic
// retryCount: number of times to retry reloading the page in case of failure
// Returns an error if any
func (nav *Navigator) ReloadPage(retryCount int) error {
	var err error
	for i := 0; i < retryCount; i++ {
		if nav.DebugLogger {
			nav.Logger.Printf("Attempt %d: Reloading the page\n", i+1)
		}
		err = chromedp.Run(nav.Ctx,
			chromedp.Reload(),
		)
		if err == nil {
			if nav.DebugLogger {
				nav.Logger.Println("Page reloaded successfully")
			}
			return nil
		}
		if nav.DebugLogger {
			nav.Logger.Printf("Info: Failed to reload page: %v. Retrying...\n", err)
		}
		time.Sleep(2 * time.Second)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Error - Failed to reload page after %d attempts: %v\n", retryCount, err)
	}
	return fmt.Errorf("error - failed to reload page after %d attempts: %v", retryCount, err)
}

// WaitPageLoad waits for the current page to fully load by checking the document.readyState property
// It will retry until the page is fully loaded or the timeout of one minute is reached
// Returns the page readyState as a string and an error if any
func (nav *Navigator) WaitPageLoad() (string, error) {
	start := time.Now()
	var pageHTML string
	for {
		if time.Since(start) > time.Minute {
			return "", fmt.Errorf("error - timeout waiting for page to fully load")
		}

		err := chromedp.Run(nav.Ctx,
			chromedp.Evaluate(`document.readyState`, &pageHTML),
		)
		if err != nil {
			return "", fmt.Errorf("error - failed to check page readiness: %v", err)
		}

		if pageHTML == "complete" {
			break
		}
		if nav.DebugLogger {
			nav.Logger.Println("INFO: Page is not fully loaded yet, retrying...")
		}
		time.Sleep(nav.Timeout)
	}

	if nav.DebugLogger {
		nav.Logger.Println("INFO: Page is fully loaded")
	}
	return pageHTML, nil
}

// GetPageSource captures all page HTML from the current page
// Returns the page HTML as a string and an error if any
// Example:
//
//	pageSource, err := nav.GetPageSource()
func (nav *Navigator) GetPageSource() (*html.Node, error) {
	if nav.DebugLogger {
		nav.Logger.Println("Getting the HTML content of the page")
	}
	var pageHTML string

	// Ensure the context is not cancelled and the page is fully loaded
	pageHTML, err := nav.WaitPageLoad()
	if err != nil {
		return nil, err
	}

	// Get the outer HTML of the page
	err = chromedp.Run(nav.Ctx,
		chromedp.OuterHTML("html", &pageHTML),
	)
	if err != nil {
		return nil, fmt.Errorf("error - failed to get page HTML: %v", err)
	}

	htmlPgSrc, err := htmlquery.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return nil, fmt.Errorf("error - failed to convert page HTML: %v", err)
	}

	if nav.DebugLogger {
		nav.Logger.Println("Page HTML retrieved successfully")
	}
	return htmlPgSrc, nil
}

// WaitForElement waits for an element specified by the selector to be visible within the given timeout.
// Example:
//
//	err := nav.WaitForElement("#elementID", 5*time.Second)
func (nav *Navigator) WaitForElement(selector string, timeout time.Duration) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Waiting for element with selector: %s to be visible\n", selector)
	}
	ctx, cancel := context.WithTimeout(nav.Ctx, timeout)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to wait for element: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Element is now visible with selector: %s\n", selector)
	}
	return nil
}

// ClickButton clicks a button specified by the selector.
// Example:
//
//	err := nav.ClickButton("#buttonID")
func (nav *Navigator) ClickButton(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Clicking button with selector: %s\n", selector)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to click button: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Button clicked successfully with selector: %s\n", selector)
	}

	time.Sleep(nav.Timeout)

	// Ensure the context is not cancelled and the page is fully loaded
	_, err = nav.WaitPageLoad()
	if err != nil {
		return err
	}
	chromedp.WaitReady("body")
	return nil
}

// UnsafeClickButton clicks a button specified by the selector. Unsafe because this methode does not use the wait element feature.
// Example:
//
//	err := nav.ClickButton("#buttonID")
func (nav *Navigator) UnsafeClickButton(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Clicking button with selector: %s\n", selector)
	}

	err := chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.ByID, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to click button: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Button clicked successfully with selector: %s\n", selector)
	}
	return nil
}

// ClickElement clicks an element specified by the selector.
// Example:
//
//	err := nav.ClickElement("#elementID")
func (nav *Navigator) ClickElement(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Clicking element with selector: %s\n", selector)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.ByID, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - Failed chromedp.ByID chromedp error: %v", err)
	}

	if nav.DebugLogger {
		nav.Logger.Printf("Element clicked with selector: %s\n", selector)
	}
	return nil
}

// CheckRadioButton selects a radio button specified by the selector.
// Example:
//
//	err := nav.CheckRadioButton("#radioButtonID")
func (nav *Navigator) CheckRadioButton(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Selecting radio button with selector: %s\n", selector)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.NodeVisible, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to select radio button: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Radio button selected successfully with selector: %s\n", selector)
	}
	return nil
}

// UncheckRadioButton unchecks a checkbox specified by the selector.
// Example:
//
//	err := nav.UncheckRadioButton("#checkboxID")
func (nav *Navigator) UncheckRadioButton(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Unchecking checkbox with selector: %s\n", selector)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.RemoveAttribute(selector, "checked", chromedp.NodeVisible, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to uncheck radio button: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Checkbox unchecked successfully with selector: %s\n", selector)
	}
	return nil
}

// FillField fills a field specified by the selector with the provided value.
// Example:
//
//	err := nav.FillField("#fieldID", "value")
func (nav *Navigator) FillField(selector string, value string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Filling field with selector: %s\n", selector)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.SendKeys(selector, value, chromedp.ByQuery, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to fill field with selector: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Field filled with selector: %s\n", selector)
	}
	return nil
}

// UnsafeFillField fills a field specified by the selector with the provided value. Unsafe because this methode does not use the wait element feature.
// Example:
//
//	err := nav.FillField("#fieldID", "value")
func (nav *Navigator) UnsafeFillField(selector string, value string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Filling field with selector: %s\n", selector)
	}

	err := chromedp.Run(nav.Ctx,
		chromedp.SendKeys(selector, value, chromedp.ByQuery, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to fill field with selector: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Field filled with selector: %s\n", selector)
	}
	return nil
}

// ExtractLinks extracts all links from the current page.
// Example:
//
//	links, err := nav.ExtractLinks()
func (nav *Navigator) ExtractLinks() ([]string, error) {
	if nav.DebugLogger {
		nav.Logger.Println("Extracting links from the current page")
	}
	var links []string
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &links),
	)
	if err != nil {
		return nil, fmt.Errorf("error - failed to extract links: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Println("Links extracted successfully")
	}
	return links, nil
}

// FillForm fills out a form specified by the selector with the provided data and submits it.
// Example:
//
//	formData := map[string]string{
//	    "username": "myUsername",
//	    "password": "myPassword",
//	}
//	err := nav.FillForm("#loginForm", formData)
func (nav *Navigator) FillForm(selector string, data map[string]string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Filling form with selector: %s and data: %v\n", selector, data)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	tasks := []chromedp.Action{
		chromedp.WaitVisible(selector, nav.QueryOption),
	}
	for field, value := range data {
		tasks = append(tasks, chromedp.SetValue(fmt.Sprintf("%s [name=%s]", selector, field), value))
	}
	tasks = append(tasks, chromedp.Submit(selector, nav.QueryOption))

	err = chromedp.Run(nav.Ctx, tasks...)
	if err != nil {
		return fmt.Errorf("error - failed to fill form: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Form filled and submitted successfully with selector: %s\n", selector)
	}
	return nil
}

// HandleAlert handles JavaScript alerts by accepting them.
// Example:
//
//	err := nav.HandleAlert()
func (nav *Navigator) HandleAlert() error {
	if nav.DebugLogger {
		nav.Logger.Println("Handling JavaScript alert by accepting it")
	}

	listenCtx, cancel := context.WithCancel(nav.Ctx)
	defer cancel()

	chromedp.ListenTarget(listenCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			if nav.DebugLogger {
				nav.Logger.Printf("Alert detected: %s", ev.Message)
			}
			err := chromedp.Run(nav.Ctx,
				page.HandleJavaScriptDialog(true),
			)
			if err != nil {
				if nav.DebugLogger {
					nav.Logger.Printf("Error - Failed to handle alert: %v\n", err)
				}
			}
		}
	})

	// Run a no-op to wait for the dialog to be handled
	err := chromedp.Run(nav.Ctx, chromedp.Sleep(nav.Timeout))
	if err != nil {
		return fmt.Errorf("error - failed to handle alert: %v", err)
	}

	if nav.DebugLogger {
		nav.Logger.Println("JavaScript alert accepted successfully")
	}
	return nil
}

// SelectDropdown selects an option in a dropdown specified by the selector and value.
// Example:
//
//	err := nav.SelectDropdown("#dropdownID", "optionValue")
func (nav *Navigator) SelectDropdown(selector, value string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Selecting dropdown option with selector: %s and value: %s\n", selector, value)
	}

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.SetValue(selector, value, chromedp.NodeVisible, nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("error - failed to select dropdown option: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Println("Dropdown option selected successfully")
	}
	return nil
}

// ExecuteScript runs the specified JavaScript on the current page
// script: the JavaScript code to execute
// Returns an error if any
func (nav *Navigator) ExecuteScript(script string) error {
	if nav.DebugLogger {
		nav.Logger.Println("Executing script on the page")
	}
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(script, nil),
	)
	if err != nil {
		return fmt.Errorf("error - failed to execute script: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Println("Script executed successfully")
	}
	return nil
}

// EvaluateScript executes a JavaScript script and returns the result
func (nav *Navigator) EvaluateScript(script string) (interface{}, error) {
	var result interface{}
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(script, &result),
	)
	if err != nil {
		return nil, fmt.Errorf("error - failed to evaluate script: %v", err)
	}
	return result, nil
}

// GetElement retrieves the text content of an element specified by the selector.
// Example:
//
//	text, err := nav.GetElement("#elementID")
func (nav *Navigator) GetElement(selector string) (string, error) {
	if nav.DebugLogger {
		nav.Logger.Printf("Getting element with selector: %s\n", selector)
	}
	var content string

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		return "", fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Text(selector, &content, chromedp.ByQuery, chromedp.NodeVisible, nav.QueryOption),
	)
	if err != nil && err.Error() != "could not find node" {
		return "", fmt.Errorf("error - failed to get element: %v", err)
	}
	if content == "" {
		if nav.DebugLogger {
			nav.Logger.Printf("Element is empty with selector: %s\n", selector)
		}
		return "", nil // Element not found or empty
	}

	if nav.DebugLogger {
		nav.Logger.Printf("Got element with selector: %s\n", selector)
	}
	return content, nil
}

// SaveImageBase64 extracts the base64 image data from the given selector and saves it to a file.
//
// Parameters:
//   - selector: the CSS selector of the CAPTCHA image element
//   - outputPath: the file path to save the image
//   - prefixClean: the prefix to clear from the source, if any
//
// Example:
//
//	err := nav.SaveImageBase64("#imagemCaptcha", "captcha.png", "data:image/png;base64,")
func (nav *Navigator) SaveImageBase64(selector, outputPath, prefixClean string) (string, error) {
	var imageData string

	// Run the tasks
	err := chromedp.Run(nav.Ctx,
		chromedp.AttributeValue(selector, "src", &imageData, nil, nav.QueryOption),
	)
	if err != nil {
		return "", fmt.Errorf("error - failed to get image data: %w", err)
	}

	var base64Data string
	if prefixClean != "" {
		// Check if the image data is in base64 format
		if !strings.HasPrefix(imageData, prefixClean) {
			if nav.DebugLogger {
				nav.Logger.Printf("Error - Unexpected image format: %v\n", err)
			}
			return "", fmt.Errorf("error - unexpected image format")
		}

		// Remove the data URL prefix
		base64Data = strings.TrimPrefix(imageData, prefixClean)
	}

	// Remove any newlines or spaces (just in case)
	base64Data = strings.ReplaceAll(base64Data, "\n", "")
	base64Data = strings.ReplaceAll(base64Data, "\r", "")
	base64Data = strings.TrimSpace(base64Data)

	// Decode the base64 data
	imageBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// Check if decoded bytes are non-zero
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("decoded image bytes are zero, something went wrong with extraction or decoding")
	}

	// Save the image to a file
	err = ioutil.WriteFile(outputPath, imageBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	if nav.DebugLogger {
		nav.Logger.Printf("Captcha image saved successfully to %s", outputPath)
	}
	return base64Data, nil
}

// MakeCaptchaElementVisible changes the style display of an element to nil
func (nav *Navigator) MakeCaptchaElementVisible(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Making CAPTCHA response field with selector: %s visible\n", selector)
	}
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s').style.display = ""`, selector), nil),
	)
	if err != nil {
		return fmt.Errorf("error - failed to make element visible: %v", err)
	}
	if nav.DebugLogger {
		nav.Logger.Printf("Element with selector: %s is now visible\n", selector)
	}
	return nil
}

// MakeElementVisible changes the style display of an element to nil
func (nav *Navigator) MakeElementVisible(selector string) error {
	if nav.DebugLogger {
		nav.Logger.Printf("Making element with selector: %s visible\n", selector)
	}

	err := chromedp.Run(nav.Ctx,
		chromedp.SetAttributeValue(selector, "type", "", nav.QueryOption),
	)
	if err != nil {
		return fmt.Errorf("failed to make element visible: %w", err)
	}

	if nav.DebugLogger {
		nav.Logger.Printf("Element with selector: %s is now visible\n", selector)
	}
	return nil
}

// Datepicker deals with date-picker elements on websites by receiving a date, calculates the amount of time it needs to go back in the picker and finally selects a day.
//
//	date: string in the format "dd/mm/aaaa"
//	calendarButtonSelector: the css selector of the data-picker
//	calendarButtonGoBack: the css selector of the go back button
//	calendarButtonsTableXpath: the xpath of the days table example: "//*[@id="ui-datepicker-div"]/table/tbody/tr";
//	calendarButtonTR: the css selector of the days table row, example: "//*[@id="ui-datepicker-div"]/table/tbody/tr"
func (nav *Navigator) Datepicker(date, calendarButtonSelector, calendarButtonGoBack, calendarButtonsTableXpath, calendarButtonTR string) error {
	regex := `^(0[1-9]|[12][0-9]|3[01])/(0[1-9]|1[0-2])/[0-9]{4}$`
	r := regexp.MustCompile(regex)
	if r.MatchString(date) == false {
		return errors.New("date does not match with dd/mm/aaaa")
	}

	parsedDate, err := time.Parse("02/01/2006", date)
	if err != nil {
		return errors.New("error parsing date: " + err.Error())
	}

	today := time.Now().Format("02/01/2006")
	parseToday, err := time.Parse("02/01/2006", today)
	if err != nil {
		return errors.New("error parsing today's date: " + err.Error())
	}

	// Ensure startDate is before endDate
	if parsedDate.After(parseToday) {
		return errors.New("date must be older then today")
	}
	years, months, _ := calculateDateDifference(parsedDate, parseToday)

	err = nav.ClickButton(calendarButtonSelector)
	if err != nil {
		return fmt.Errorf("error selecting callendar button, err:%s", err)
	}

	i := 0
	for {
		err = chromedp.Run(nav.Ctx, chromedp.Click(calendarButtonGoBack))
		if err != nil {
			break
		}
		i++
		if i == ((years * 12) + months) {
			break
		}
	}

	err = nav.WaitForElement(calendarButtonsTableXpath, time.Minute)
	if err != nil {
		return fmt.Errorf("error waiting for element on: %s, error:%s", calendarButtonsTableXpath, err)
	}

	pageSource, err := nav.GetPageSource()
	if err != nil {
		return fmt.Errorf("error - failed to get page source: %w", err)
	}

	tt, err := htmlquery.Find(pageSource, calendarButtonsTableXpath)
	if err != nil {
		return fmt.Errorf("error - failed to find calendar buttons table on path: %s, err: %w", calendarButtonsTableXpath, err)
	}

	for k, node := range tt {
		for l := 1; l < 8; l++ {
			day, err := ExtractText(node, "td["+strconv.Itoa(l)+"]", "")
			if err != nil {
				return err
			}
			if day == strconv.Itoa(parsedDate.Day()) {
				err = nav.ClickButton(calendarButtonsTableXpath + "[" + strconv.Itoa(k+1) + "]/td[" + strconv.Itoa(l) + "]")
				if err != nil {
					return errors.New("error clicking button on calendar button: " + calendarButtonTR + "(" + strconv.Itoa(k) + ") > td:nth-child(" + strconv.Itoa(l) + "). Error code: " + err.Error())
				} else {
					return nil
				}
			}

		}

	}
	return errors.New("could not pick date")
}
func calculateDateDifference(startDate, endDate time.Time) (int, int, int) {
	years := endDate.Year() - startDate.Year()
	months := int(endDate.Month()) - int(startDate.Month())
	days := endDate.Day() - startDate.Day()

	// Adjust the difference if necessary
	if days < 0 {
		// Borrow days from the previous month
		months--
		previousMonth := endDate.AddDate(0, -1, 0)
		days += previousMonth.Day()
	}

	if months < 0 {
		// Borrow months from the previous year
		years--
		months += 12
	}

	return years, months, days
}

// ParseHtmlToString used for parsing html.node into string for debugging purposes
func ParseHtmlToString(pageSource *html.Node) (string, error) {
	var sb strings.Builder
	err := html.Render(&sb, pageSource)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// ParseStringToHtmlNode takes a string and returns an *html.Node
func ParseStringToHtmlNode(pageSource string) (*html.Node, error) {
	reader := strings.NewReader(pageSource)
	node, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// Close closes the Navigator instance and releases resources.
// Example:
//
//	nav.Close()
func (nav *Navigator) Close() {
	if nav.DebugLogger {
		nav.Logger.Println("Closing the Navigator instance")
	}
	nav.Cancel()
	if nav.DebugLogger {
		nav.Logger.Println("Navigator instance closed successfully")
	}
}

// Request structure to hold user data
type Request struct {
	SearchString string
}

// PageSource structure to hold the HTML data
type PageSource struct {
	Page    *html.Node
	Request string
	Error   error
}

// RemovePageSource removes the element at index `s` from a slice of `PageSource` objects.
// It returns the modified slice without the element at index `s`.
func RemovePageSource(slice []PageSource, s int) []PageSource {
	return append(slice[:s], slice[s+1:]...)
}

// RemoveRequest removes the element at index `s` from a slice of `Request` objects.
// It returns the modified slice without the element at index `s`.
func RemoveRequest(slice []Request, s int) []Request {
	return append(slice[:s], slice[s+1:]...)
}

// ParallelRequests performs web scraping tasks concurrently with a specified number of workers and a delay between requests.
// The crawlerFunc parameter allows for flexibility in defining the web scraping logic.
//
// Parameters:
// - requests: A slice of Request structures containing the data needed for each request.
// - numberOfWorkers: The number of concurrent workers to process the requests.
// - delay: The delay duration between each request to avoid overwhelming the target server.
// - crawlerFunc: A user-defined function that takes a process number as input and returns the html as *html.Node, and an error.
//
// Returns:
// - A slice of ResponseBody structures containing the results of the web scraping tasks.
// - An error if any occurred during the requests.
//
// Example Usage:
//
// results, err := ParallelRequests(requests, numberOfWorkers, delay, crawlerFunc)
func ParallelRequests(requests []Request, numberOfWorkers int, delay time.Duration, crawlerFunc func(string) (*html.Node, error)) ([]PageSource, error) {
	done := make(chan struct{})
	defer close(done)

	inputCh := streamInputs(done, requests)
	resultCh := make(chan PageSource, len(requests)) // Buffered channel to hold all results

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for req := range inputCh {
				log.Printf("Worker %d processing request: %s", workerID, req.SearchString)
				time.Sleep(delay)
				pageSource, err := crawlerFunc(req.SearchString)
				resultCh <- PageSource{
					Page:    pageSource,
					Request: req.SearchString,
					Error:   err,
				}
			}
		}(i)
	}

	// Close the result channel once all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results from the result channel
	var results []PageSource
	var errorOnApiRequests error

	for result := range resultCh {
		if result.Error != nil {
			errorOnApiRequests = result.Error
		}
		results = append(results, result)
	}

	return results, errorOnApiRequests
}

// streamInputs streams the input requests into a channel.
//
// Parameters:
// - done: A channel to signal when to stop processing inputs.
// - requests: A slice of Request structures containing the data needed for each request.
//
// Returns:
// - A channel that streams the input requests.
//
// Example Usage:
//
// inputCh := streamInputs(done, requests)
func streamInputs(done <-chan struct{}, requests []Request) <-chan Request {
	inputCh := make(chan Request)
	go func() {
		defer close(inputCh)
		for _, req := range requests {
			select {
			case inputCh <- req:
			case <-done:
				return
			}
		}
	}()
	return inputCh
}

// EvaluateParallelRequests iterates over a set of previous results, evaluates them using the provided evaluation function,
// and handles re-crawling of problematic sources until all sources are valid or no further progress can be made.
//
// Parameters:
// - previousResults: A slice of PageSource objects containing the initial crawl results.
// - crawlerFunc: A function that takes a string (URL or identifier) and returns a parsed HTML node and an error.
// - evaluate: A function that takes a slice of PageSource objects and returns two slices:
//  1. A slice of Request objects for sources that need to be re-crawled.
//  2. A slice of valid PageSource objects.
//
// Returns:
// - A slice of valid PageSource objects after all problematic sources have been re-crawled and evaluated.
// - An error if there is a failure in the crawling process.
//
// Example usage:
//
// results, err := EvaluateParallelRequests(resultsFirst, Crawler, Eval)
//
//	func Eval(previousResults []PageSource) ([]Request, []PageSource) {
//		var newRequests []Request
//		var validResults []PageSource
//
//		for _, result := range previousResults {
//			_, err := extractDataCover(result.Page, "")
//			if err != nil {
//				newRequests = append(newRequests, Request{SearchString: result.Request})
//			} else {
//				validResults = append(validResults, result)
//			}
//		}
//
//		return newRequests, validResults
//	}
func EvaluateParallelRequests(previousResults []PageSource, crawlerFunc func(string) (*html.Node, error), evaluate func([]PageSource) ([]Request, []PageSource)) ([]PageSource, error) {
	for {
		problematicPageSources, newResults := evaluate(previousResults)
		if len(problematicPageSources) == 0 {
			return newResults, nil
		}

		log.Printf("Crawling %d problematic sources", len(problematicPageSources))
		temporaryResults, err := ParallelRequests(problematicPageSources, 10, 0, crawlerFunc)
		if err != nil {
			return nil, fmt.Errorf("failed to crawl page sources, error: %s", err)
		}

		previousResults = newResults
		for _, tempResult := range temporaryResults {
			previousResults = append(previousResults, tempResult)
		}
	}
}

// ExtractTable extracts data from a table specified by the selector.
// Example:
//
//	tableData, err := goSpider.ExtractTableData(pageSource,"#tableID")
func ExtractTable(pageSource *html.Node, tableRowsExpression string) ([]*html.Node, error) {
	log.Printf("Extracting table data with selector: %s\n", tableRowsExpression)
	rows, err := htmlquery.Find(pageSource, tableRowsExpression)
	if err != nil {
		return nil, fmt.Errorf("failed to extract table data, error: %s", err)
	}
	if len(rows) > 0 {
		return rows, nil
	}
	// log.Printf("Table data extracted successfully")
	return nil, errors.New("could not find any table rows")
}

// ExtractText extracts text content from nodes specified by the parent selectors.
// Example:
//
//	textData, err := goSpider.ExtractText(pageSource,"#parent1", "\n")
func ExtractText(node *html.Node, nodeExpression string, Dirt string) (string, error) {
	//log.Print("Extracting text from node")
	var text string
	tt, err := htmlquery.Find(node, nodeExpression)
	if err != nil {
		return "", fmt.Errorf("failed to extract text, error: %s", err)
	}
	if len(tt) > 0 {
		text = strings.TrimSpace(strings.Replace(htmlquery.InnerText(htmlquery.FindOne(node, nodeExpression)), Dirt, "", -1))
		return text, nil
	}

	//log.Printf("Text %v extracted successfully from node", nodeExpression)
	return "", errors.New("could not find specified text")
}

// FindNodes extracts nodes content from nodes specified by the parent selectors.
// Example:
//
//	nodeData, err := goSpider.FindNode(pageSource,"#parent1")
func FindNodes(node *html.Node, nodeExpression string) ([]*html.Node, error) {
	n, err := htmlquery.Find(node, nodeExpression)
	if err != nil {
		return nil, fmt.Errorf("failed to find nodes, error: %s", err)
	}
	if len(n) > 0 {
		return n, nil
	}
	return nil, errors.New("could not find specified node")
}

// GetElementAttributeFromNode retrieves the value of a specified attribute from an element
// located using an XPath expression within a given HTML node.
// Parameters:
// - node: The root HTML node to search within.
// - xpathExpr: The XPath expression that identifies the target element.
// - attribute: The attribute name whose value you want to retrieve.
// Returns:
// - The attribute value as a string.
// - An error if the element or attribute cannot be found.
func GetElementAttributeFromNode(node *html.Node, xpathExpr, attribute string) (string, error) {
	// Locate the element using the provided XPath expression.
	target := htmlquery.FindOne(node, xpathExpr)
	if target == nil {
		return "", fmt.Errorf("failed to find element for XPath: %s", xpathExpr)
	}

	// Retrieve the attribute's value.
	// Option 1: using a loop to search through the node's attributes.
	for _, attr := range target.Attr {
		if attr.Key == attribute {
			return attr.Val, nil
		}
	}

	// Option 2: using htmlquery.SelectAttr (if you prefer a one-liner)
	// value := htmlquery.SelectAttr(target, attribute)
	// if value != "" {
	//     return value, nil
	// }

	return "", fmt.Errorf("attribute %s not found in element", attribute)
}
