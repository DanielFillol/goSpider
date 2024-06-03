package goSpider

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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

// Requests structure to hold user data
type Requests struct {
	ProcessNumber string
}

// ResponseBody structure to hold response data
type ResponseBody struct {
	Cover     map[string]string
	Movements []map[int]map[string]interface{}
	People    []map[int]map[string]interface{}
	Error     error
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

// ExtractTableData extracts data from a table specified by the selector.
// Example:
//
//	tableData, err := nav.ExtractTableData("#tableID")
func (nav *Navigator) ExtractTableData(selector string) ([]map[int]map[string]interface{}, error) {
	nav.Logger.Printf("Extracting table data with selector: %s\n", selector)
	var rows []*cdp.Node
	err := chromedp.Run(nav.Ctx,
		chromedp.Nodes(selector+" tr", &rows, chromedp.ByQueryAll),
	)
	if err != nil {
		nav.Logger.Printf("Failed to extract table rows: %v\n", err)
		return nil, fmt.Errorf("failed to extract table rows: %v", err)
	}

	var tableData []map[int]map[string]interface{}
	for _, row := range rows {
		// nav.Logger.Printf("Processing row %d", rowIndex)
		var cells []*cdp.Node
		err = chromedp.Run(nav.Ctx,
			chromedp.Nodes("td, th", &cells, chromedp.ByQueryAll, chromedp.FromNode(row)),
		)
		if err != nil {
			nav.Logger.Printf("Failed to extract table cells: %v\n", err)
			return nil, fmt.Errorf("failed to extract table cells: %v", err)
		}

		rowData := make(map[int]map[string]interface{})
		for cellIndex, cell := range cells {
			// nav.Logger.Printf("Processing cell %d in row %d", cellIndex, rowIndex)
			cellData := make(map[string]interface{})

			var cellText string
			err = chromedp.Run(nav.Ctx,
				chromedp.Text(cell.FullXPath(), &cellText, chromedp.NodeVisible),
			)
			if err != nil {
				nav.Logger.Printf("Failed to get cell text: %v\n", err)
				return nil, fmt.Errorf("failed to get cell text: %v", err)
			}
			cellData["text"] = cellText

			// Check for any nested spans within the cell
			var nestedSpans []*cdp.Node
			nestedSpansErr := chromedp.Run(nav.Ctx,
				chromedp.Nodes(cell.FullXPath()+"//span", &nestedSpans, chromedp.ByQueryAll),
			)
			if nestedSpansErr != nil {
				// nav.Logger.Printf("No nested spans found in cell %d of row %d: %v\n", cellIndex, rowIndex, nestedSpansErr)
				// No nested spans found, continue processing
				nestedSpans = []*cdp.Node{}
			}

			spanData := make(map[int]string)
			for spanIndex, span := range nestedSpans {
				// nav.Logger.Printf("Processing span %d in cell %d of row %d", spanIndex, cellIndex, rowIndex)
				var spanText string
				err = chromedp.Run(nav.Ctx,
					chromedp.Text(span.FullXPath(), &spanText, chromedp.NodeVisible),
				)
				if err != nil {
					nav.Logger.Printf("Failed to get span text: %v\n", err)
					return nil, fmt.Errorf("failed to get span text: %v", err)
				}
				spanData[spanIndex] = spanText
			}

			if len(spanData) > 0 {
				cellData["spans"] = spanData
			}

			rowData[cellIndex] = cellData
		}
		tableData = append(tableData, rowData)
	}
	// nav.Logger.Println("Table data extracted successfully")
	return tableData, nil
}

// ExtractDivText extracts text content from divs specified by the parent selectors.
// Example:
//
//	textData, err := nav.ExtractDivText("#parent1", "#parent2")
func (nav *Navigator) ExtractDivText(parentSelectors ...string) (map[string]string, error) {
	nav.Logger.Println("Extracting text from divs")
	data := make(map[string]string)
	for _, parentSelector := range parentSelectors {
		var nodes []*cdp.Node
		err := chromedp.Run(nav.Ctx,
			chromedp.Nodes(parentSelector+" span, "+parentSelector+" div", &nodes, chromedp.ByQueryAll),
		)
		if err != nil {
			nav.Logger.Printf("Failed to extract nodes from %s: %v\n", parentSelector, err)
			return nil, fmt.Errorf("failed to extract nodes from %s: %v", parentSelector, err)
		}
		for _, node := range nodes {
			if node.NodeType == cdp.NodeTypeText {
				continue
			}
			var text string
			err = chromedp.Run(nav.Ctx,
				chromedp.TextContent(node.FullXPath(), &text),
			)
			if err != nil {
				nav.Logger.Printf("Failed to extract text content from %s: %v\n", node.FullXPath(), err)
				return nil, fmt.Errorf("failed to extract text content from %s: %v", node.FullXPath(), err)
			}
			data[node.AttributeValue("id")] = strings.TrimSpace(text)
		}
	}
	// nav.Logger.Println("Text extracted successfully from divs")
	return data, nil
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

// FetchHTML fetches the HTML content of the specified URL.
// Example:
//
//	htmlContent, err := nav.FetchHTML("https://www.example.com")
func (nav *Navigator) FetchHTML(url string) (string, error) {
	nav.Logger.Printf("Fetching HTML content from URL: %s\n", url)
	var htmlContent string
	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		nav.Logger.Printf("Failed to fetch URL: %v\n", err)
		return "", fmt.Errorf("failed to fetch URL: %v", err)
	}
	nav.Logger.Println("HTML content fetched successfully")
	return htmlContent, nil
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

// ParallelRequests performs web scraping tasks concurrently with a specified number of workers and a delay between requests.
// The crawlerFunc parameter allows for flexibility in defining the web scraping logic.
//
// Parameters:
// - requests: A slice of Requests structures containing the data needed for each request.
// - numberOfWorkers: The number of concurrent workers to process the requests.
// - duration: The delay duration between each request to avoid overwhelming the target server.
// - crawlerFunc: A user-defined function that takes a process number as input and returns cover data, movements, people, and an error.
//
// Returns:
// - A slice of ResponseBody structures containing the results of the web scraping tasks.
// - An error if any occurred during the requests.
//
// Example Usage:
//
//	results, err := asyncRequest(requests, numberOfWorkers, duration, crawlerFunc)
func ParallelRequests(requests []Requests, numberOfWorkers int, duration time.Duration, crawlerFunc func(string) (map[string]string, []map[int]map[string]interface{}, []map[int]map[string]interface{}, error)) ([]ResponseBody, error) {
	done := make(chan struct{})
	defer close(done)

	inputCh := streamInputs(done, requests)
	var wg sync.WaitGroup
	resultCh := make(chan ResponseBody, len(requests)) // Buffered channel to hold all results

	k := 0
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for input := range inputCh {
				k++
				time.Sleep(duration)
				cover, movements, people, err := crawlerFunc(input.ProcessNumber)
				resultCh <- ResponseBody{
					Cover:     cover,
					Movements: movements,
					People:    people,
					Error:     err,
				}
				if err != nil {
					log.Println(err)
					continue
				}
				if k == len(requests)-1 {
					break
				}
			}
		}()
	}

	// Close the result channel once all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var results []ResponseBody
	var errorOnApiRequests error

	// Collect results from the result channel
	for result := range resultCh {
		if result.Error != nil {
			errorOnApiRequests = result.Error
		}
		results = append(results, result)
	}

	if k == len(requests)-1 {
		l := log.New(os.Stdout, "goSpider: ", log.LstdFlags)
		l.Printf("Finished processing %d requests\n", len(requests))
		return results, errorOnApiRequests
	}

	return results, errorOnApiRequests
}

// streamInputs streams the input requests into a channel.
//
// Parameters:
// - done: A channel to signal when to stop processing inputs.
// - inputs: A slice of Requests structures containing the data needed for each request.
//
// Returns:
// - A channel that streams the input requests.
//
// Example Usage:
//
//	inputCh := streamInputs(done, inputs)
func streamInputs(done <-chan struct{}, inputs []Requests) <-chan Requests {
	inputCh := make(chan Requests)
	go func() {
		defer close(inputCh)
		for _, input := range inputs {
			select {
			case inputCh <- input:
			case <-done:
				return
			}
		}
	}()
	return inputCh
}
