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
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Navigator is a struct that holds the context for the ChromeDP session and a logger.
type Navigator struct {
	Ctx     context.Context
	Cancel  context.CancelFunc
	Logger  *log.Logger
	Timeout time.Duration
	Cookies []*network.Cookie
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
		Logger:  logger,
		Cookies: []*network.Cookie{},
	}

	// Set standard timeout with enhanced logging
	navigator.SetTimeOut(300 * time.Millisecond)
	logger.Printf("Navigator initialized with timeout: %v\n", navigator.Timeout)

	return navigator
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
		chromedp.AttributeValue(selector, attribute, &value, nil),
	)
	if err != nil {
		return "", fmt.Errorf("error getting attribute %s: %v", attribute, err)
	}
	return value, nil
}

// SwitchToFrame switches the context to the specified iframe.
func (nav *Navigator) SwitchToFrame(selector string) error {
	nav.Logger.Println("Switching to frame", selector)

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

	nav.Logger.Println("SwitchToFrame", selector, "successfully")
	return nil
}

// SwitchToDefaultContent switches the context back to the main content from an iframe context.
func (nav *Navigator) SwitchToDefaultContent() error {
	nav.Logger.Println("Switching to default content")

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

	nav.Logger.Println("Switch to default content successfully")
	return nil
}

// OpenURL opens the specified URL in the current browser context.
// Example:
//
//	err := nav.OpenURL("https://www.example.com")
func (nav *Navigator) OpenURL(url string) error {
	nav.Logger.Printf("Opening URL: %s\n", url)
	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"), // Ensures the page is fully loaded
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to open URL: %v\n", err)
		return fmt.Errorf("error - failed to open URL: %v", err)
	}

	_, err = nav.WaitPageLoad()
	if err != nil {
		return err
	}

	nav.Logger.Printf("URL opened successfully with URL: %s\n", url)
	return nil
}

// GetCurrentURL returns the current URL of the browser.
// Example:
//
//	currentURL, err := nav.GetCurrentURL()
func (nav *Navigator) GetCurrentURL() (string, error) {
	nav.Logger.Println("Extracting the current URL")
	var currentURL string
	err := chromedp.Run(nav.Ctx,
		chromedp.Location(&currentURL),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to extract current URL: %v\n", err)
		return "", fmt.Errorf("error - failed to extract current URL: %v", err)
	}
	nav.Logger.Println("Current URL extracted successfully")
	return currentURL, nil
}

// Login logs into a website using the provided credentials and selectors.
// Example:
//
//	err := nav.Login("https://www.example.com/login", "username", "password", "#username", "#password", "#login-button", "#login-message-fail")
func (nav *Navigator) Login(url, username, password, usernameSelector, passwordSelector, loginButtonSelector string, messageFailedSuccess string) error {
	nav.Logger.Printf("Logging into URL: %s\n", url)

	if url != "" {
		err := nav.OpenURL(url)
		if err != nil {
			nav.Logger.Printf("Error - Failed to open URL: %v\n", err)
			return fmt.Errorf("error - failed to open URL: %v", err)
		}
	}

	err := nav.WaitForElement(usernameSelector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = nav.WaitForElement(passwordSelector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = nav.WaitForElement(loginButtonSelector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
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
				nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
				return fmt.Errorf("error - failed waiting for element: %v", err)
			}
			message, err := nav.GetElement(messageFailedSuccess)
			if err == nil {
				nav.Logger.Printf("Error - Failed to log in: %v\n", err)
				nav.Logger.Printf("Message found: %s", message)
				return fmt.Errorf("error - message: %v", message)
			} else {
				nav.Logger.Printf("Message was not found")
				return fmt.Errorf("error - failed to log in: %v", err)
			}
		}
		nav.Logger.Printf("Error - Failed to log in: %v\n", err)
		return fmt.Errorf("error - failed to log in: %v", err)
	}

	//sometimes the page does accept the login information but still returns a error message
	if messageFailedSuccess != "" {
		err = nav.WaitForElement(messageFailedSuccess, nav.Timeout)
		if err == nil {
			message, err := nav.GetElement(messageFailedSuccess)
			if err == nil {
				nav.Logger.Printf("Message found: %s", message)
				return fmt.Errorf("error - message: %v", message)
			} else {
				nav.Logger.Printf("Message was not found")
				return fmt.Errorf("error - failed to log in: %v", err)
			}
		}
	}

	nav.Logger.Println("Logged in successfully")
	return nil
}

// LoginAccountsGoogle performs the Google login on the given URL
func (nav *Navigator) LoginAccountsGoogle(email, password string) error {
	nav.Logger.Printf("Opening URL: %s\n", "accounts.google.com")
	err := chromedp.Run(nav.Ctx, chromedp.Navigate("https://accounts.google.com"))
	if err != nil {
		nav.Logger.Printf("Failed to open URL: %v\n", err)
		return fmt.Errorf("failed to open URL: %v", err)
	}

	_, err = nav.WaitPageLoad()
	if err != nil {
		nav.Logger.Printf("Failed to WaitPageLoad: %v\n", err)
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Fill the Google login form
	nav.Logger.Println("Filling in the Google login form")
	err = nav.FillField(`#identifierId`, email)
	if err != nil {
		err = nav.WaitForElement("#yDmH0d > c-wiz > div > div:nth-child(2) > div > c-wiz > c-wiz > div > div.s7iwrf.gMPiLc.Kdcijb > div > div > header > h1", nav.Timeout)
		if err != nil {
			nav.Logger.Printf("Error - Failed to log in: %v\n", err)
			return fmt.Errorf("error - failed to check login: %v", err)
		} else {
			s, err := nav.GetElement("#yDmH0d > c-wiz > div > div:nth-child(2) > div > c-wiz > c-wiz > div > div.s7iwrf.gMPiLc.Kdcijb > div > div > header > h1")
			if err != nil {
				nav.Logger.Printf("Error - Failed to log in: %v\n", err)
				return fmt.Errorf("error - failed to check login: %v", err)
			}
			nav.Logger.Printf("Already logged in! \n%s", s)
			return nil
		}
	}

	err = nav.ClickButton(`#identifierNext`)
	if err != nil {
		nav.Logger.Printf("Failed to click the 'Next' button: %v\n", err)
		return fmt.Errorf("failed to click the 'Next' button: %v", err)
	}

	// Adding a small delay to allow the next page to load
	_, err = nav.WaitPageLoad()
	if err != nil {
		nav.Logger.Printf("Failed to WaitPageLoad: %v\n", err)
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}
	time.Sleep(2 * time.Second)

	err = nav.FillField("#password > div.aCsJod.oJeWuf > div > div.Xb9hP > input", password)
	if err != nil {
		nav.Logger.Printf("Failed to fill the password field: %v\n", err)
		return fmt.Errorf("failed to fill the password field: %v", err)
	}

	err = nav.ClickButton(`#passwordNext`)
	if err != nil {
		nav.Logger.Printf("Failed to click the 'Next' button for password: %v\n", err)
		return fmt.Errorf("failed to click the 'Next' button for password: %v", err)
	}

	// Adding a small delay to allow the next page to load
	_, err = nav.WaitPageLoad()
	if err != nil {
		nav.Logger.Printf("Failed to WaitPageLoad: %v\n", err)
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}
	time.Sleep(2 * time.Second)

	authCode := AskForString("Google verification pass: ")

	//"#yDmH0d > c-wiz > div > div.UXFQgc > div > div > div > form > span > section:nth-child(2) > div > div > div.AFTWye.GncK > div > div.aCsJod.oJeWuf > div > div.Xb9hP"
	err = nav.FillField("#idvPin", authCode)
	if err != nil {
		nav.Logger.Printf("Failed to fill the idvPin with code: %s\n field: %v\n", authCode, err)
		return fmt.Errorf("failed to fill the idvPin with code: %s\n field: %v\n", authCode, err)
	}

	nav.Logger.Println("Google login completed successfully")
	return nil
}

// LoginWithGoogle performs the Google login on the given URL
func (nav *Navigator) LoginWithGoogle(url string) error {
	nav.Logger.Printf("Opening URL: %s\n", url)
	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(url),
	)
	if err != nil {
		nav.Logger.Printf("Failed to open URL: %v\n", err)
		return fmt.Errorf("failed to open URL: %v", err)
	}

	nav.Logger.Println("Clicking the 'Continuar com o Google' button")
	err = nav.ClickButton(".SocialButton")
	if err != nil {
		nav.Logger.Printf("Alredy logged in: %v\n", err)
		return nil
		//nav.Logger.Printf("Failed to click the Google login button: %v\n", err)
		//return fmt.Errorf("failed to click the Google login button: %v", err)
	}

	// Wait for the popup to appear and switch to it
	nav.Logger.Println("Switching to the Google login popup")
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
			nav.Logger.Println("Failed to detect the Google login popup")
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
		nav.Logger.Printf("Failed to get the current URL of the popup: %v\n", err)
		return err
	}
	fmt.Printf("Popup URL: %s\n", currentURL)

	// Check if the popup navigated to the Google login page
	if !strings.Contains(currentURL, "accounts.google.com") {
		nav.Logger.Printf("Popup did not navigate to Google login page, current URL: %s\n", currentURL)
		return fmt.Errorf("popup did not navigate to Google login page")
	}

	// Increase the timeout for filling the form fields
	popupCtx, popupCancel = context.WithTimeout(popupCtx, nav.Timeout)
	defer popupCancel()

	// Fill the Google login form
	err = newNav.ClickElement("#container")
	if err != nil {
		nav.Logger.Printf("Failed to click element: %v\n", err)
		return fmt.Errorf("failed to click element: %v", err)
	}

	_, err = newNav.WaitPageLoad()
	if err != nil {
		nav.Logger.Printf("Failed to WaitPageLoad: %v\n", err)
		return fmt.Errorf("failed to WaitPageLoad: %v", err)
	}

	err = newNav.ClickButton("#credentials-picker > div.fFW7wc-ibnC6b-sM5MNb.TAKBxb")
	if err != nil {
		nav.Logger.Printf("Failed to click button: %v\n", err)
		return fmt.Errorf("failed to click button: %v", err)
	}

	newNav.Logger.Println("Google login completed successfully")
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
	nav.Logger.Println("Capturing screenshot")
	err := chromedp.Run(nav.Ctx,
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to capture screenshot: %v\n", err)
		return fmt.Errorf("error - failed to capture screenshot: %v", err)
	}
	err = ioutil.WriteFile(nameFile+"_screenshot.png", buf, 0644)
	if err != nil {
		nav.Logger.Printf("Error - Failed to save screenshot: %v\n", err)
		return fmt.Errorf("error - failed to save screenshot: %v", err)
	}
	nav.Logger.Printf("Screenshot saved successfully with name: %s\n", nameFile)
	return nil
}

// ReloadPage reloads the current page with retry logic
// retryCount: number of times to retry reloading the page in case of failure
// Returns an error if any
func (nav *Navigator) ReloadPage(retryCount int) error {
	var err error
	for i := 0; i < retryCount; i++ {
		nav.Logger.Printf("Attempt %d: Reloading the page\n", i+1)
		err = chromedp.Run(nav.Ctx,
			chromedp.Reload(),
		)
		if err == nil {
			nav.Logger.Println("Page reloaded successfully")
			return nil
		}
		nav.Logger.Printf("Info: Failed to reload page: %v. Retrying...\n", err)
		time.Sleep(2 * time.Second)
	}
	nav.Logger.Printf("Error - Failed to reload page after %d attempts: %v\n", retryCount, err)
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
			nav.Logger.Println("Error - Timeout waiting for page to fully load")
			return "", fmt.Errorf("error - timeout waiting for page to fully load")
		}

		err := chromedp.Run(nav.Ctx,
			chromedp.Evaluate(`document.readyState`, &pageHTML),
		)
		if err != nil {
			nav.Logger.Printf("Error - Failed to check page readiness: %v\n", err)
			return "", fmt.Errorf("error - failed to check page readiness: %v", err)
		}

		if pageHTML == "complete" {
			break
		}
		nav.Logger.Println("INFO: Page is not fully loaded yet, retrying...")
		time.Sleep(nav.Timeout)
	}

	nav.Logger.Println("INFO: Page is fully loaded")
	return pageHTML, nil
}

// GetPageSource captures all page HTML from the current page
// Returns the page HTML as a string and an error if any
// Example:
//
//	pageSource, err := nav.GetPageSource()
func (nav *Navigator) GetPageSource() (*html.Node, error) {
	nav.Logger.Println("Getting the HTML content of the page")
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
		nav.Logger.Printf("Error - Failed to get page HTML: %v\n", err)
		return nil, fmt.Errorf("error - failed to get page HTML: %v", err)
	}

	htmlPgSrc, err := htmlquery.Parse(strings.NewReader(pageHTML))
	if err != nil {
		nav.Logger.Printf("Error - failed to convert page HTML: %v", err)
		return nil, fmt.Errorf("error - failed to convert page HTML: %v", err)
	}

	nav.Logger.Println("Page HTML retrieved successfully")
	return htmlPgSrc, nil
}

// WaitForElement waits for an element specified by the selector to be visible within the given timeout.
// Example:
//
//	err := nav.WaitForElement("#elementID", 5*time.Second)
func (nav *Navigator) WaitForElement(selector string, timeout time.Duration) error {
	nav.Logger.Printf("Waiting for element with selector: %s to be visible\n", selector)
	ctx, cancel := context.WithTimeout(nav.Ctx, timeout)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to wait for element: %v\n", err)
		return fmt.Errorf("error - failed to wait for element: %v", err)
	}
	nav.Logger.Printf("Element is now visible with selector: %s\n", selector)
	return nil
}

// ClickButton clicks a button specified by the selector.
// Example:
//
//	err := nav.ClickButton("#buttonID")
func (nav *Navigator) ClickButton(selector string) error {
	nav.Logger.Printf("Clicking button with selector: %s\n", selector)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to click button: %v\n", err)
		return fmt.Errorf("error - failed to click button: %v", err)
	}
	nav.Logger.Printf("Button clicked successfully with selector: %s\n", selector)

	time.Sleep(nav.Timeout)

	// Ensure the context is not cancelled and the page is fully loaded
	_, err = nav.WaitPageLoad()
	if err != nil {
		return err
	}
	chromedp.WaitReady("body")
	return nil
}

// ClickElement clicks an element specified by the selector.
// Example:
//
//	err := nav.ClickElement("#elementID")
func (nav *Navigator) ClickElement(selector string) error {
	nav.Logger.Printf("Clicking element with selector: %s\n", selector)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.ByID),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed chromedp.ByID clicking element: %v\n", err)
		return fmt.Errorf("error - Failed chromedp.ByID chromedp error: %v", err)
	}

	nav.Logger.Printf("Element clicked with selector: %s\n", selector)
	return nil
}

// CheckRadioButton selects a radio button specified by the selector.
// Example:
//
//	err := nav.CheckRadioButton("#radioButtonID")
func (nav *Navigator) CheckRadioButton(selector string) error {
	nav.Logger.Printf("Selecting radio button with selector: %s\n", selector)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to select radio button: %v\n", err)
		return fmt.Errorf("error - failed to select radio button: %v", err)
	}
	nav.Logger.Printf("Radio button selected successfully with selector: %s\n", selector)
	return nil
}

// UncheckRadioButton unchecks a checkbox specified by the selector.
// Example:
//
//	err := nav.UncheckRadioButton("#checkboxID")
func (nav *Navigator) UncheckRadioButton(selector string) error {
	nav.Logger.Printf("Unchecking checkbox with selector: %s\n", selector)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.RemoveAttribute(selector, "checked", chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to uncheck radio button: %v\n", err)
		return fmt.Errorf("error - failed to uncheck radio button: %v", err)
	}
	nav.Logger.Printf("Checkbox unchecked successfully with selector: %s\n", selector)
	return nil
}

// FillField fills a field specified by the selector with the provided value.
// Example:
//
//	err := nav.FillField("#fieldID", "value")
func (nav *Navigator) FillField(selector string, value string) error {
	nav.Logger.Printf("Filling field with selector: %s\n", selector)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.SendKeys(selector, value, chromedp.ByQuery),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to fill field with selector: %v\n", err)
		return fmt.Errorf("error - failed to fill field with selector: %v", err)
	}
	nav.Logger.Printf("Field filled with selector: %s\n", selector)
	return nil
}

// ExtractLinks extracts all links from the current page.
// Example:
//
//	links, err := nav.ExtractLinks()
func (nav *Navigator) ExtractLinks() ([]string, error) {
	nav.Logger.Println("Extracting links from the current page")
	var links []string
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &links),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to extract links: %v\n", err)
		return nil, fmt.Errorf("error - failed to extract links: %v", err)
	}
	nav.Logger.Println("Links extracted successfully")
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
	nav.Logger.Printf("Filling form with selector: %s and data: %v\n", selector, data)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	tasks := []chromedp.Action{
		chromedp.WaitVisible(selector),
	}
	for field, value := range data {
		tasks = append(tasks, chromedp.SetValue(fmt.Sprintf("%s [name=%s]", selector, field), value))
	}
	tasks = append(tasks, chromedp.Submit(selector))

	err = chromedp.Run(nav.Ctx, tasks...)
	if err != nil {
		nav.Logger.Printf("Error - Failed to fill form: %v\n", err)
		return fmt.Errorf("error - failed to fill form: %v", err)
	}
	nav.Logger.Printf("Form filled and submitted successfully with selector: %s\n", selector)
	return nil
}

// HandleAlert handles JavaScript alerts by accepting them.
// Example:
//
//	err := nav.HandleAlert()
func (nav *Navigator) HandleAlert() error {
	nav.Logger.Println("Handling JavaScript alert by accepting it")

	listenCtx, cancel := context.WithCancel(nav.Ctx)
	defer cancel()

	chromedp.ListenTarget(listenCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			nav.Logger.Printf("Alert detected: %s", ev.Message)
			err := chromedp.Run(nav.Ctx,
				page.HandleJavaScriptDialog(true),
			)
			if err != nil {
				nav.Logger.Printf("Error - Failed to handle alert: %v\n", err)
			}
		}
	})

	// Run a no-op to wait for the dialog to be handled
	err := chromedp.Run(nav.Ctx, chromedp.Sleep(nav.Timeout))
	if err != nil {
		nav.Logger.Printf("Error - Failed to handle alert: %v\n", err)
		return fmt.Errorf("error - failed to handle alert: %v", err)
	}

	nav.Logger.Println("JavaScript alert accepted successfully")
	return nil
}

// SelectDropdown selects an option in a dropdown specified by the selector and value.
// Example:
//
//	err := nav.SelectDropdown("#dropdownID", "optionValue")
func (nav *Navigator) SelectDropdown(selector, value string) error {
	nav.Logger.Printf("Selecting dropdown option with selector: %s and value: %s\n", selector, value)

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.SetValue(selector, value, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to select dropdown option: %v\n", err)
		return fmt.Errorf("error - failed to select dropdown option: %v", err)
	}
	nav.Logger.Println("Dropdown option selected successfully")
	return nil
}

// ExecuteScript runs the specified JavaScript on the current page
// script: the JavaScript code to execute
// Returns an error if any
func (nav *Navigator) ExecuteScript(script string) error {
	nav.Logger.Println("Executing script on the page")
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(script, nil),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to execute script: %v\n", err)
		return fmt.Errorf("error - failed to execute script: %v", err)
	}
	nav.Logger.Println("Script executed successfully")
	return nil
}

// EvaluateScript executes a JavaScript script and returns the result
func (nav *Navigator) EvaluateScript(script string) (interface{}, error) {
	var result interface{}
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(script, &result),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to evaluate script: %v\n", err)
		return nil, fmt.Errorf("error - failed to evaluate script: %v", err)
	}
	return result, nil
}

// GetElement retrieves the text content of an element specified by the selector.
// Example:
//
//	text, err := nav.GetElement("#elementID")
func (nav *Navigator) GetElement(selector string) (string, error) {
	nav.Logger.Printf("Getting element with selector: %s\n", selector)
	var content string

	err := nav.WaitForElement(selector, nav.Timeout)
	if err != nil {
		nav.Logger.Printf("Error - Failed waiting for element: %v\n", err)
		return "", fmt.Errorf("error - failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Text(selector, &content, chromedp.ByQuery, chromedp.NodeVisible),
	)
	if err != nil && err.Error() != "could not find node" {
		nav.Logger.Printf("Error - Failed to get element: %v\n", err)
		return "", fmt.Errorf("error - failed to get element: %v", err)
	}
	if content == "" {
		nav.Logger.Printf("Element is empty with selector: %s\n", selector)
		return "", nil // Element not found or empty
	}

	nav.Logger.Printf("Got element with selector: %s\n", selector)
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
		chromedp.AttributeValue(selector, "src", &imageData, nil),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to get image data: %v\n", err)
		return "", fmt.Errorf("error - failed to get image data: %w", err)
	}

	var base64Data string
	if prefixClean != "" {
		// Check if the image data is in base64 format
		if !strings.HasPrefix(imageData, prefixClean) {
			nav.Logger.Printf("Error - Unexpected image format: %v\n", err)
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

	nav.Logger.Printf("Captcha image saved successfully to %s", outputPath)
	return base64Data, nil
}

// MakeElementVisible changes the style display of an element to nil
func (nav *Navigator) MakeElementVisible(selector string) error {
	nav.Logger.Printf("Making CAPTCHA response field with selector: %s visible\n", selector)
	err := chromedp.Run(nav.Ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s').style.display = ""`, selector), nil),
	)
	if err != nil {
		nav.Logger.Printf("Error - Failed to make element visible: %v\n", err)
		return fmt.Errorf("error - failed to make element visible: %v", err)
	}
	nav.Logger.Printf("Element with selector: %s is now visible\n", selector)
	return nil
}

// Close closes the Navigator instance and releases resources.
// Example:
//
//	nav.Close()
func (nav *Navigator) Close() {
	// nav.Logger.Println("Closing the Navigator instance")
	nav.Cancel()
	nav.Logger.Println("Navigator instance closed successfully")
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
