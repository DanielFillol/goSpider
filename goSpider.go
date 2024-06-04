package goSpider

import (
	"context"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
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
	Ctx    context.Context
	Cancel context.CancelFunc
	Logger *log.Logger
}

// NewNavigator creates a new Navigator instance.
// Example:
//
//	nav := goSpider.NewNavigator()
func NewNavigator() *Navigator {
	ctx, cancel := chromedp.NewContext(context.Background())
	logger := log.New(os.Stdout, "goSpider: ", log.LstdFlags)
	return &Navigator{
		Ctx:    ctx,
		Cancel: cancel,
		Logger: logger,
	}
}

// OpenNewTab opens a new browser tab with the specified URL.
// Example:
//
//	err := nav.OpenNewTab("https://www.example.com")
func (nav *Navigator) OpenNewTab(url string) error {
	nav.Logger.Printf("Opening new tab with URL: %s\n", url)
	ctx, cancel := chromedp.NewContext(nav.Ctx)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
	)
	if err != nil {
		nav.Logger.Printf("Failed to open new tab: %v\n", err)
		return fmt.Errorf("failed to open new tab: %v", err)
	}
	// nav.Logger.Println("New tab opened successfully")
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
		nav.Logger.Printf("Failed to open URL: %v\n", err)
		return fmt.Errorf("failed to open URL: %v", err)
	}
	// nav.Logger.Println("URL opened successfully")
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
		nav.Logger.Printf("Failed to extract current URL: %v\n", err)
		return "", fmt.Errorf("failed to extract current URL: %v", err)
	}
	// nav.Logger.Println("Current URL extracted successfully")
	return currentURL, nil
}

// Login logs into a website using the provided credentials and selectors.
// Example:
//
//	err := nav.Login("https://www.example.com/login", "username", "password", "#username", "#password", "#login-button", "Login failed")
func (nav *Navigator) Login(url, username, password, usernameSelector, passwordSelector, loginButtonSelector string, messageFailedSuccess string) error {
	nav.Logger.Printf("Logging into URL: %s\n", url)
	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(usernameSelector, chromedp.ByQuery),
		chromedp.SendKeys(usernameSelector, username, chromedp.ByQuery),
		chromedp.WaitVisible(passwordSelector, chromedp.ByQuery),
		chromedp.SendKeys(passwordSelector, password, chromedp.ByQuery),
		chromedp.WaitVisible(loginButtonSelector, chromedp.ByQuery),
		chromedp.Click(loginButtonSelector, chromedp.ByQuery),
		chromedp.WaitReady("body"), // Wait for the next page to load
	)
	if err != nil {
		if messageFailedSuccess != "" {
			message, err := nav.GetElement(messageFailedSuccess)
			if err == nil {
				nav.Logger.Printf("Message found: %s", message)
			} else {
				nav.Logger.Printf("Message was not found")
			}
		}

		nav.Logger.Printf("Failed to log in: %v\n", err)
		return fmt.Errorf("failed to log in: %v", err)
	}
	// nav.Logger.Println("Logged in successfully")
	return nil
}

// CaptureScreenshot captures a screenshot of the current browser window.
// Example:
//
//	err := nav.CaptureScreenshot()
func (nav *Navigator) CaptureScreenshot() error {
	var buf []byte
	// nav.Logger.Println("Capturing screenshot")
	err := chromedp.Run(nav.Ctx,
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		nav.Logger.Printf("Failed to capture screenshot: %v\n", err)
		return fmt.Errorf("failed to capture screenshot: %v", err)
	}
	err = ioutil.WriteFile("screenshot.png", buf, 0644)
	if err != nil {
		nav.Logger.Printf("Failed to save screenshot: %v\n", err)
		return fmt.Errorf("failed to save screenshot: %v", err)
	}
	nav.Logger.Println("Screenshot saved successfully")
	return nil
}

// GetPageSource capture all page HTML from current page
// Return the page HTML as a string and an error if any
// Example:
//
//	pageSource,err := nav.GetPageSource()
func (nav *Navigator) GetPageSource() (*html.Node, error) {
	nav.Logger.Println("Getting the HTML content of the page")
	var pageHTML string
	err := chromedp.Run(nav.Ctx,
		chromedp.OuterHTML("html", &pageHTML),
	)
	if err != nil {
		nav.Logger.Printf("Failed to get page HTML: %v\n", err)
		return nil, fmt.Errorf("failed to get page HTML: %v", err)
	}

	htmlPgSrc, err := htmlquery.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to convert page HTML: %v", err)
	}

	//nav.Logger.Println("Page HTML retrieved successfully")
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
	_ = chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
	)
	// if err != nil {
	//     nav.Logger.Printf("Failed to wait for element: %v\n", err)
	//     return fmt.Errorf("failed to wait for element: %v", err)
	// }
	// nav.Logger.Println("Element is now visible")
	return nil
}

// ClickButton clicks a button specified by the selector.
// Example:
//
//	err := nav.ClickButton("#buttonID")
func (nav *Navigator) ClickButton(selector string) error {
	nav.Logger.Printf("Clicking button with selector: %s\n", selector)

	err := nav.WaitForElement(selector, 3*time.Second)
	if err != nil {
		nav.Logger.Printf("Failed waiting for element: %v\n", err)
		return fmt.Errorf("failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to click button: %v\n", err)
		return fmt.Errorf("failed to click button: %v", err)
	}
	// nav.Logger.Println("Button clicked successfully")
	chromedp.WaitReady("body")
	return nil
}

// ClickElement clicks an element specified by the selector.
// Example:
//
//	err := nav.ClickElement("#elementID")
func (nav *Navigator) ClickElement(selector string) error {
	nav.Logger.Printf("Clicking element with selector: %s\n", selector)
	err := chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.ByID),
	)
	if err != nil {
		log.Printf("chromedp error: %v", err)
	}

	return nil
}

// CheckRadioButton selects a radio button specified by the selector.
// Example:
//
//	err := nav.CheckRadioButton("#radioButtonID")
func (nav *Navigator) CheckRadioButton(selector string) error {
	nav.Logger.Printf("Selecting radio button with selector: %s\n", selector)

	err := nav.WaitForElement(selector, 3*time.Second)
	if err != nil {
		nav.Logger.Printf("Failed waiting for element: %v\n", err)
		return fmt.Errorf("failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to select radio button: %v\n", err)
		return fmt.Errorf("failed to select radio button: %v", err)
	}
	// nav.Logger.Println("Radio button selected successfully")
	return nil
}

// UncheckRadioButton unchecks a checkbox specified by the selector.
// Example:
//
//	err := nav.UncheckRadioButton("#checkboxID")
func (nav *Navigator) UncheckRadioButton(selector string) error {
	nav.Logger.Printf("Unchecking checkbox with selector: %s\n", selector)

	err := nav.WaitForElement(selector, 3*time.Second)
	if err != nil {
		nav.Logger.Printf("Failed waiting for element: %v\n", err)
		return fmt.Errorf("failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.RemoveAttribute(selector, "checked", chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to uncheck radio button: %v\n", err)
		return fmt.Errorf("failed to uncheck radio button: %v", err)
	}
	// nav.Logger.Println("Checkbox unchecked successfully")
	return nil
}

// FillField fills a field specified by the selector with the provided value.
// Example:
//
//	err := nav.FillField("#fieldID", "value")
func (nav *Navigator) FillField(selector string, value string) error {
	nav.Logger.Printf("Filling field with selector: %s\n", selector)
	err := nav.WaitForElement(selector, 3*time.Second)
	if err != nil {
		nav.Logger.Printf("Failed waiting for element: %v\n", err)
		return fmt.Errorf("failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.SendKeys(selector, value, chromedp.ByQuery),
	)
	if err != nil {
		nav.Logger.Printf("Failed to fill field with selector: %v\n", err)
		return fmt.Errorf("failed to fill field with selector: %v", err)
	}
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
		nav.Logger.Printf("Failed to extract links: %v\n", err)
		return nil, fmt.Errorf("failed to extract links: %v", err)
	}
	// nav.Logger.Println("Links extracted successfully")
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
func (nav *Navigator) FillForm(formSelector string, data map[string]string) error {
	nav.Logger.Printf("Filling form with selector: %s and data: %v\n", formSelector, data)
	tasks := []chromedp.Action{
		chromedp.WaitVisible(formSelector),
	}
	for field, value := range data {
		tasks = append(tasks, chromedp.SetValue(fmt.Sprintf("%s [name=%s]", formSelector, field), value))
	}
	tasks = append(tasks, chromedp.Submit(formSelector))

	err := chromedp.Run(nav.Ctx, tasks...)
	if err != nil {
		nav.Logger.Printf("Failed to fill form: %v\n", err)
		return fmt.Errorf("failed to fill form: %v", err)
	}
	// nav.Logger.Println("Form filled and submitted successfully")
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
				nav.Logger.Printf("Failed to handle alert: %v\n", err)
			}
		}
	})

	// Run a no-op to wait for the dialog to be handled
	err := chromedp.Run(nav.Ctx, chromedp.Sleep(2*time.Second))
	if err != nil {
		nav.Logger.Printf("Failed to handle alert: %v\n", err)
		return fmt.Errorf("failed to handle alert: %v", err)
	}

	// nav.Logger.Println("JavaScript alert accepted successfully")
	return nil
}

// SelectDropdown selects an option in a dropdown specified by the selector and value.
// Example:
//
//	err := nav.SelectDropdown("#dropdownID", "optionValue")
func (nav *Navigator) SelectDropdown(selector, value string) error {
	nav.Logger.Printf("Selecting dropdown option with selector: %s and value: %s\n", selector, value)
	err := chromedp.Run(nav.Ctx,
		chromedp.SetValue(selector, value, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to select dropdown option: %v\n", err)
		return fmt.Errorf("failed to select dropdown option: %v", err)
	}
	// nav.Logger.Println("Dropdown option selected successfully")
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

// GetElement retrieves the text content of an element specified by the selector.
// Example:
//
//	text, err := nav.GetElement("#elementID")
func (nav *Navigator) GetElement(selector string) (string, error) {
	var content string

	err := nav.WaitForElement(selector, 3*time.Second)
	if err != nil {
		nav.Logger.Printf("Failed waiting for element: %v\n", err)
		return "", fmt.Errorf("failed waiting for element: %v", err)
	}

	err = chromedp.Run(nav.Ctx,
		chromedp.Text(selector, &content, chromedp.ByQuery, chromedp.NodeVisible),
	)
	if err != nil && err.Error() != "could not find node" {
		nav.Logger.Printf("Failed to get element: %v\n", err)
		return "", fmt.Errorf("failed to get element: %v", err)
	}
	if content == "" {
		return "", nil // Element not found or empty
	}
	return content, nil
}

// Requests structure to hold user data
type Requests struct {
	SearchString string
}

// PageSource structure to hold the HTML data
type PageSource struct {
	Page  *html.Node
	Error error
}

// ParallelRequests performs web scraping tasks concurrently with a specified number of workers and a delay between requests.
// The crawlerFunc parameter allows for flexibility in defining the web scraping logic.
//
// Parameters:
// - requests: A slice of Requests structures containing the data needed for each request.
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
func ParallelRequests(requests []Requests, numberOfWorkers int, delay time.Duration, crawlerFunc func(string) (*html.Node, error)) ([]PageSource, error) {
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
					Page:  pageSource,
					Error: err,
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
// - requests: A slice of Requests structures containing the data needed for each request.
//
// Returns:
// - A channel that streams the input requests.
//
// Example Usage:
//
// inputCh := streamInputs(done, requests)
func streamInputs(done <-chan struct{}, requests []Requests) <-chan Requests {
	inputCh := make(chan Requests)
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

// ExtractTable extracts data from a table specified by the selector.
// Example:
//
//	tableData, err := goSpider.ExtractTableData("#tableID")
func ExtractTable(pageSource *html.Node, tableRowsExpression string) ([]*html.Node, error) {
	log.Printf("Extracting table data with selector: %s\n", tableRowsExpression)
	rows := htmlquery.Find(pageSource, tableRowsExpression)
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
	tt := htmlquery.Find(node, nodeExpression)
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
//	textData, err := goSpider.FindNode(pageSource,"#parent1")
func FindNodes(node *html.Node, nodeExpression string) ([]*html.Node, error) {
	n := htmlquery.Find(node, nodeExpression)
	if len(n) > 0 {
		return n, nil
	}
	return nil, errors.New("could not find specified node")
}
