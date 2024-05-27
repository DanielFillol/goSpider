package goSpider

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"time"
)

// Navigator is a struct that holds the context for the ChromeDP session and a logger
type Navigator struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Logger *log.Logger
}

// NewNavigator creates a new Navigator instance
func NewNavigator() *Navigator {
	ctx, cancel := chromedp.NewContext(context.Background())
	logger := log.New(os.Stdout, "webnav: ", log.LstdFlags)
	return &Navigator{
		Ctx:    ctx,
		Cancel: cancel,
		Logger: logger,
	}
}

// FetchHTML fetches the HTML content of a given URL
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

// ClickButton clicks a button identified by the given selector
func (nav *Navigator) ClickButton(selector string) error {
	nav.Logger.Printf("Clicking button with selector: %s\n", selector)
	err := chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to click button: %v\n", err)
		return fmt.Errorf("failed to click button: %v", err)
	}
	nav.Logger.Println("Button clicked successfully")
	return nil
}

// FillSearchBar fills a search bar identified by the given selector and submits the form
func (nav *Navigator) FillSearchBar(selector, query string) error {
	nav.Logger.Printf("Filling search bar with selector: %s and query: %s\n", selector, query)
	err := chromedp.Run(nav.Ctx,
		chromedp.SetValue(selector, query, chromedp.NodeVisible),
		chromedp.EvaluateAsDevTools(fmt.Sprintf(`document.querySelector('%s').closest('form').submit()`, selector), nil),
	)
	if err != nil {
		nav.Logger.Printf("Failed to fill search bar: %v\n", err)
		return fmt.Errorf("failed to fill search bar: %v", err)
	}

	nav.Logger.Println("Search bar filled and form submitted successfully")
	return nil
}

// OpenNewTab opens a new tab with the given URL
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
	nav.Logger.Println("New tab opened successfully")
	return nil
}

// ExtractLinks extracts all the links from the current page
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
	nav.Logger.Println("Links extracted successfully")
	return links, nil
}

// ExtractText extracts all the text content from the current page
func (nav *Navigator) ExtractText() (string, error) {
	nav.Logger.Println("Extracting text content from the current page")
	var text string
	err := chromedp.Run(nav.Ctx,
		chromedp.OuterHTML("body", &text, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to extract text: %v\n", err)
		return "", fmt.Errorf("failed to extract text: %v", err)
	}
	nav.Logger.Println("Text content extracted successfully")
	return text, nil
}

// FillForm fills a form with the provided data
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
	nav.Logger.Println("Form filled and submitted successfully")
	return nil
}

// HandleAlert handles a JavaScript alert by accepting it
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

	nav.Logger.Println("JavaScript alert accepted successfully")
	return nil
}

// SelectDropdown selects an option from a dropdown menu identified by the selector and option value
func (nav *Navigator) SelectDropdown(selector, value string) error {
	nav.Logger.Printf("Selecting dropdown option with selector: %s and value: %s\n", selector, value)
	err := chromedp.Run(nav.Ctx,
		chromedp.SetValue(selector, value, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to select dropdown option: %v\n", err)
		return fmt.Errorf("failed to select dropdown option: %v", err)
	}
	nav.Logger.Println("Dropdown option selected successfully")
	return nil
}

// CheckCheckbox checks a checkbox identified by the selector
func (nav *Navigator) CheckCheckbox(selector string) error {
	nav.Logger.Printf("Checking checkbox with selector: %s\n", selector)
	err := chromedp.Run(nav.Ctx,
		chromedp.SetAttributeValue(selector, "checked", "true", chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to check checkbox: %v\n", err)
		return fmt.Errorf("failed to check checkbox: %v", err)
	}
	nav.Logger.Println("Checkbox checked successfully")
	return nil
}

// UncheckCheckbox unchecks a checkbox identified by the selector
func (nav *Navigator) UncheckCheckbox(selector string) error {
	nav.Logger.Printf("Unchecking checkbox with selector: %s\n", selector)
	err := chromedp.Run(nav.Ctx,
		chromedp.RemoveAttribute(selector, "checked", chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to uncheck checkbox: %v\n", err)
		return fmt.Errorf("failed to uncheck checkbox: %v", err)
	}
	nav.Logger.Println("Checkbox unchecked successfully")
	return nil
}

// SelectRadioButton selects a radio button identified by the selector
func (nav *Navigator) SelectRadioButton(selector string) error {
	nav.Logger.Printf("Selecting radio button with selector: %s\n", selector)
	err := chromedp.Run(nav.Ctx,
		chromedp.Click(selector, chromedp.NodeVisible),
	)
	if err != nil {
		nav.Logger.Printf("Failed to select radio button: %v\n", err)
		return fmt.Errorf("failed to select radio button: %v", err)
	}
	nav.Logger.Println("Radio button selected successfully")
	return nil
}

// UploadFile uploads a file to a file input identified by the selector
func (nav *Navigator) UploadFile(selector, filePath string) error {
	nav.Logger.Printf("Uploading file with selector: %s and file path: %s\n", selector, filePath)
	err := chromedp.Run(nav.Ctx,
		chromedp.SetUploadFiles(selector, []string{filePath}),
	)
	if err != nil {
		nav.Logger.Printf("Failed to upload file: %v\n", err)
		return fmt.Errorf("failed to upload file: %v", err)
	}
	nav.Logger.Println("File uploaded successfully")
	return nil
}

// WaitForElement waits for an element identified by the selector to be visible
func (nav *Navigator) WaitForElement(selector string, timeout time.Duration) error {
	nav.Logger.Printf("Waiting for element with selector: %s to be visible\n", selector)
	ctx, cancel := context.WithTimeout(nav.Ctx, timeout)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
	)
	if err != nil {
		nav.Logger.Printf("Failed to wait for element: %v\n", err)
		return fmt.Errorf("failed to wait for element: %v", err)
	}
	nav.Logger.Println("Element is now visible")
	return nil
}

// WaitForAJAX waits for AJAX requests to complete by monitoring the network activity
func (nav *Navigator) WaitForAJAX(timeout time.Duration) error {
	nav.Logger.Println("Waiting for AJAX requests to complete")
	ctx, cancel := context.WithTimeout(nav.Ctx, timeout)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Sleep(timeout),
	)
	if err != nil {
		nav.Logger.Printf("Failed to wait for AJAX requests: %v\n", err)
		return fmt.Errorf("failed to wait for AJAX requests: %v", err)
	}
	nav.Logger.Println("AJAX requests completed")
	return nil
}

// Close closes the Navigator instance
func (nav *Navigator) Close() {
	nav.Logger.Println("Closing the Navigator instance")
	nav.Cancel()
	nav.Logger.Println("Navigator instance closed successfully")
}

// GetCurrentURL extracts the current URL from the browser
// Returns the current URL as a string and an error if any
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
	nav.Logger.Println("Current URL extracted successfully")
	return currentURL, nil
}
