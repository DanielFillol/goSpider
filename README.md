# goSpider Navigation Library

This Go library provides functions to navigate websites and retrieve information using `chromedp`. It supports basic actions like fetching HTML content, clicking buttons, filling forms, handling alerts, and more complex interactions such as dynamically loaded content.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Example Usage](#example-usage)
- [API Reference](#api-reference)
  - [Navigator](#navigator)
    - [NewNavigator](#newnavigator)
    - [Close](#close)
    - [SetQueryType](#setquerytype)
    - [UseXPath](#usexpath)
    - [UseCSS](#usecss)
    - [SetTimeOut](#settimeout)
  - [Navigation](#navigation)
    - [OpenURL](#openurl)
    - [GetCurrentURL](#getcurrenturl)
    - [SwitchToNewTab](#switchtonewtab)
    - [SwitchToFrame](#switchtoframe)
    - [SwitchToDefaultContent](#switchtodefaultcontent)
    - [ReloadPage](#reloadpage)
    - [WaitPageLoad](#waitpageload)
  - [Page Interaction](#page-interaction)
    - [ClickButton](#clickbutton)
    - [UnsafeClickButton](#unsafeclickbutton)
    - [ClickElement](#clickelement)
    - [CheckRadioButton](#checkradiobutton)
    - [UncheckRadioButton](#uncheckradiobutton)
    - [FillField](#fillfield)
    - [UnsafeFillField](#unsafefillfield)
    - [FillForm](#fillform)
    - [SelectDropdown](#selectdropdown)
    - [HandleAlert](#handlealert)
    - [ExecuteScript](#executescript)
    - [EvaluateScript](#evaluatescript)
    - [Datepicker](#datepicker)
  - [Data Extraction](#data-extraction)
    - [GetPageSource](#getpagesource)
    - [GetElement](#getelement)
    - [GetElementAttribute](#getelementattribute)
    - [ExtractLinks](#extractlinks)
    - [ExtractTable](#extracttable)
    - [ExtractText](#extracttext)
    - [FindNodes](#findnodes)
    - [GetElementAttributeFromNode](#getelementattributefromnode)
    - [SaveImageBase64](#saveimagebase64)
  - [Login](#login)
    - [Login](#login-1)
    - [LoginWithGoogle](#loginwithgoogle)
    - [LoginAccountsGoogle](#loginaccountsgoogle)
  - [Utilities](#utilities)
    - [CaptureScreenshot](#capturescreenshot)
    - [CheckPageTitle](#checkpagetitle)
    - [MakeCaptchaElementVisible](#makecaptchaelementvisible)
    - [MakeElementVisible](#makeelementvisible)
    - [AskForString](#askforstring)
    - [ParseHtmlToString](#parsehtmltostring)
    - [ParseStringToHtmlNode](#parsestringtohtmlnode)
  - [Concurrency](#concurrency)
    - [ParallelRequests](#parallelrequests)
    - [EvaluateParallelRequests](#evaluateparallelrequests)
    - [RemovePageSource](#removepagesource)
    - [RemoveRequest](#removerequest)
- [Structs](#structs)
  - [Request](#request)
  - [PageSource](#pagesource)

## Installation

To use this library, you need to install:

```sh
go get github.com/chromedp/chromedp
go get github.com/DanielFillol/goSpider
```

## Usage

Importing the Library
First, import the library in your Go project:

```go
import "DanielFillol/goSpider"
```

## Example Usage

Here's an example of how to use the library:
If you need a more complete example of use, please take a look on this [project](https://github.com/DanielFillol/crawlers)

```go
package main

import (
	"fmt"
	"github.com/DanielFillol/goSpider"
	"golang.org/x/net/html"
	"log"
	"time"
)

func main() {
	// Create a list of search strings (e.g., case numbers)
	requests := []goSpider.Request{
		{SearchString: "1017927-35.2023.8.26.0008"},
		{SearchString: "0002396-75.2013.8.26.0201"},
		// ... add more search strings here
	}

	// Set the number of concurrent workers and the delay between requests
	numberOfWorkers := 1
	delay := 0 * time.Millisecond

	// Run the parallel requests
	results, err := goSpider.ParallelRequests(requests, numberOfWorkers, delay, Crawler)
	if err != nil {
		log.Printf("Error during parallel requests: %v", err)
	}

	// Process the results
	log.Printf("Finished parallel requests! Got %d results.", len(results))
	for _, result := range results {
		if result.Error != nil {
			log.Printf("Error processing request '%s': %v", result.Request, result.Error)
			continue
		}
		// ... process the successful result (result.Page)
		log.Printf("Successfully processed request: %s", result.Request)
	}
}

// Crawler is the function that will be executed by each worker.
func Crawler(searchString string) (*html.Node, error) {
	// Create a new navigator instance
	nav := goSpider.NewNavigator("", true)
	defer nav.Close() // Ensure the browser is closed when the function finishes

	// Navigate to the target URL
	url := "https://esaj.tjsp.jus.br/cpopg/open.do"
	if err := nav.OpenURL(url); err != nil {
		return nil, fmt.Errorf("failed to open URL: %w", err)
	}

	// Interact with the page
	if err := nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)"); err != nil {
		return nil, fmt.Errorf("failed to check radio button: %w", err)
	}

	if err := nav.FillField("#nuProcessoAntigoFormatado", searchString); err != nil {
		return nil, fmt.Errorf("failed to fill search field: %w", err)
	}

	if err := nav.ClickButton("#botaoConsultarProcessos"); err != nil {
		return nil, fmt.Errorf("failed to click search button: %w", err)
	}

	// Wait for the results to load
	if err := nav.WaitForElement("#tabelaUltimasMovimentacoes > tr:nth-child(1) > td.dataMovimentacao", 15*time.Second); err != nil {
		return nil, fmt.Errorf("failed to wait for results: %w", err)
	}

	// Get the page source
	pageSource, err := nav.GetPageSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get page source: %w", err)
	}

	return pageSource, nil
}
```

## API Reference

### Navigator

#### NewNavigator

`NewNavigator(profilePath string, headless bool) *Navigator`

Creates a new instance of the Navigator struct, initializing a new ChromeDP context and logger.
- `profilePath`: the path to chrome profile defined by the user;can be passed as an empty string
- `headless`: if false will show chrome UI

```go
nav := goSpider.NewNavigator()
```

#### Close

`Close()`

Closes the Navigator instance and releases resources.

```go
nav.Close()
```

#### SetQueryType

`SetQueryType(queryType chromedp.QueryOption)`

Sets the query selector type for the navigator. It can be `chromedp.ByQuery` (for CSS selectors) or `chromedp.BySearch` (for XPath selectors).

```go
nav.SetQueryType(chromedp.BySearch)
```

#### UseXPath

`UseXPath()`

A helper function that sets the query selector type to `chromedp.BySearch` (XPath).

```go
nav.UseXPath()
```

#### UseCSS

`UseCSS()`

A helper function that sets the query selector type to `chromedp.ByQuery` (CSS).

```go
nav.UseCSS()
```

#### SetTimeOut

`SetTimeOut(timeOut time.Duration)`

Sets a timeout for all the waiting functions in the package. The standard timeout of the Navigator is 300 ms.

```go
nav.SetTimeOut(5 * time.Second)
```

### Navigation

#### OpenURL

`OpenURL(url string) error`

Opens the specified URL in the current browser context.

```go
err := nav.OpenURL("https://www.example.com")
```

#### GetCurrentURL

`GetCurrentURL() (string, error)`

Returns the current URL of the browser.

```go
currentURL, err := nav.GetCurrentURL()
```

#### SwitchToNewTab

`SwitchToNewTab() (*Navigator, error)`

Switches the browser context to a new tab.

```go
newNav, err := nav.SwitchToNewTab()
if err != nil {
    log.Fatal(err)
}
// Use newNav for operations in the new tab
```

#### SwitchToFrame

`SwitchToFrame(selector string) error`

Switches the context to the specified iframe.

```go
err := nav.SwitchToFrame("#my-iframe")
```

#### SwitchToDefaultContent

`SwitchToDefaultContent() error`

Switches the context back to the main content from an iframe context.

```go
err := nav.SwitchToDefaultContent()
```

#### ReloadPage

`ReloadPage(retryCount int) error`

Reloads the current page with retry logic.

```go
err := nav.ReloadPage(3)
```

#### WaitPageLoad

`WaitPageLoad() (string, error)`

Waits for the current page to fully load by checking the `document.readyState` property.

```go
readyState, err := nav.WaitPageLoad()
```

### Page Interaction

#### ClickButton

`ClickButton(selector string) error`

Clicks a button specified by the selector.

```go
err := nav.ClickButton("#buttonID")
```

#### UnsafeClickButton

`UnsafeClickButton(selector string) error`

Clicks a button specified by the selector, but without waiting for the element to be visible first. This is faster but can be less reliable.

```go
err := nav.UnsafeClickButton("#buttonID")
```

#### ClickElement

`ClickElement(selector string) error`

Clicks an element specified by the selector.

```go
err := nav.ClickElement("#elementID")
```

#### CheckRadioButton

`CheckRadioButton(selector string) error`

Selects a radio button specified by the selector.

```go
err := nav.CheckRadioButton("#radioButtonID")
```

#### UncheckRadioButton

`UncheckRadioButton(selector string) error`

Unchecks a checkbox specified by the selector.

```go
err := nav.UncheckRadioButton("#checkboxID")
```

#### FillField

`FillField(selector string, value string) error`

Fills a field specified by the selector with the provided value.

```go
err := nav.FillField("#fieldID", "value")
```

#### UnsafeFillField

`UnsafeFillField(selector string, value string) error`

Fills a field specified by the selector with the provided value, but without waiting for the element to be visible first. This is faster but can be less reliable.

```go
err := nav.UnsafeFillField("#fieldID", "some value")
```

#### FillForm

`FillForm(formSelector string, data map[string]string) error`

Fills out a form specified by the selector with the provided data and submits it.

```go
formData := map[string]string{
    "username": "myUsername",
    "password": "myPassword",
}
err := nav.FillForm("#loginForm", formData)
```

#### SelectDropdown

`SelectDropdown(selector, value string) error`

Selects an option in a dropdown specified by the selector and value.

```go
err := nav.SelectDropdown("#dropdownID", "optionValue")
```

#### HandleAlert

`HandleAlert() error`

Handles JavaScript alerts by accepting them.

```go
err := nav.HandleAlert()
```

#### ExecuteScript

`ExecuteScript(script string) error`

Runs the specified JavaScript on the current page.

```go
err := nav.ExecuteScript("alert('Hello from goSpider!')")
```

#### EvaluateScript

`EvaluateScript(script string) (interface{}, error)`

Executes a JavaScript script and returns the result.

```go
result, err := nav.EvaluateScript("1 + 1")
fmt.Println(result) // Output: 2
```

#### Datepicker

`Datepicker(date, calendarButtonSelector, calendarButtonGoBack, calendarButtonsTableXpath, calendarButtonTR string) error`

Deals with date-picker elements on websites by receiving a date, calculating the amount of time it needs to go back in the picker, and finally selecting a day.

- `date`: string in the format "dd/mm/aaaa"
- `calendarButtonSelector`: the css selector of the data-picker
- `calendarButtonGoBack`: the css selector of the go back button
- `calendarButtonsTableXpath`: the xpath of the days table example: "//*[@id="ui-datepicker-div"]/table/tbody/tr";
- `calendarButtonTR`: the css selector of the days table row, example: "//*[@id="ui-datepicker-div"]/table/tbody/tr"

```go
err := nav.Datepicker("01/01/2023", "#datepicker-button", "#prev-month", "//*[@id='ui-datepicker-div']/table/tbody/tr", "//*[@id='ui-datepicker-div']/table/tbody/tr")
```

### Data Extraction

#### GetPageSource

`GetPageSource() (*html.Node, error)`

Captures all page HTML from the current page and returns it as an `*html.Node`.

```go
pageSource, err := nav.GetPageSource()
```

#### GetElement

`GetElement(selector string) (string, error)`

Retrieves the text content of an element specified by the selector.

```go
text, err := nav.GetElement("#elementID")
```

#### GetElementAttribute

`GetElementAttribute(selector, attribute string) (string, error)`

Retrieves the value of a specified attribute from an element identified by a selector.

```go
href, err := nav.GetElementAttribute("#my-link", "href")
```

#### ExtractLinks

`ExtractLinks() ([]string, error)`

Extracts all links from the current page.

```go
links, err := nav.ExtractLinks()
```

#### ExtractTable

`ExtractTable(pageSource *html.Node, tableRowsExpression string) ([]*html.Node, error)`

Extracts rows from a table specified by the selector.

```go
tableRows, err := goSpider.ExtractTable(pageSource,"#tableID tr")
```

#### ExtractText

`ExtractText(node *html.Node, nodeExpression string, Dirt string) (string, error)`

Extracts text content from nodes specified by the parent selectors.

```go
textData, err := goSpider.ExtractText(pageSource,"#parent1", "\n")
```

#### FindNodes

`FindNodes(node *html.Node, nodeExpression string) ([]*html.Node, error)`

Extracts nodes from a given node and expression.

```go
nodes, err := goSpider.FindNodes(pageSource,".my-nodes")
```

#### GetElementAttributeFromNode

`GetElementAttributeFromNode(node *html.Node, xpathExpr, attribute string) (string, error)`

Retrieves the value of a specified attribute from an element located using an XPath expression within a given HTML node.

```go
// Assuming 'node' is an *html.Node
value, err := goSpider.GetElementAttributeFromNode(node, "//img", "src")
```

#### SaveImageBase64

`SaveImageBase64(selector, outputPath, prefixClean string) (string, error)`

Extracts the base64 image data from the given selector and saves it to a file.

```go
base64Data, err := nav.SaveImageBase64("#captcha-image", "captcha.png", "data:image/png;base64,")
```

### Login

#### Login

`Login(url, username, password, usernameSelector, passwordSelector, loginButtonSelector string, messageFailedSuccess string) error`

Logs into a website using the provided credentials and selectors.

```go
err := nav.Login("https://www.example.com/login", "username", "password", "#username", "#password", "#login-button", "Login failed")
```

#### LoginWithGoogle

`LoginWithGoogle(url string) error`

Performs the Google login on the given URL.

```go
err := nav.LoginWithGoogle("https://www.example.com")
```

#### LoginAccountsGoogle

`LoginAccountsGoogle(email, password string) error`

Performs the Google login on `accounts.google.com`. The 2FA code is passed on prompt.

```go
err := nav.LoginAccountsGoogle("your-email@gmail.com", "your-password")
```

### Utilities

#### CaptureScreenshot

`CaptureScreenshot(nameFile string) error`

Captures a screenshot of the current browser window and saves it with the given name.

```go
err := nav.CaptureScreenshot("my_screenshot")
```

#### CheckPageTitle

`CheckPageTitle(url string) (bool, error)`

Navigates to the provided URL and checks if the page title equals "Ah, não!". It returns true if the error title is detected, otherwise false.

```go
isError, err := nav.CheckPageTitle("https://www.example.com")
```

#### MakeCaptchaElementVisible

`MakeCaptchaElementVisible(selector string) error`

Changes the style `display` of an element to `""` to make it visible. This is useful for interacting with hidden CAPTCHA elements.

```go
err := nav.MakeCaptchaElementVisible("#hidden-captcha")
```

#### MakeElementVisible

`MakeElementVisible(selector string) error`

Changes the `type` attribute of an element to `""`. This can be used to make certain elements visible or interactable.

```go
err := nav.MakeElementVisible("#hidden-element")
```

#### AskForString

`AskForString(prompt string) string`

Prompts the user to enter a string in the console and returns the trimmed input. This is useful for things like 2FA codes.

```go
code := goSpider.AskForString("Enter your 2FA code: ")
```

#### ParseHtmlToString

`ParseHtmlToString(pageSource *html.Node) (string, error)`

A utility function for parsing an `*html.Node` into a string, which can be useful for debugging.

```go
// Assuming 'pageSource' is an *html.Node
htmlString, err := goSpider.ParseHtmlToString(pageSource)
```

#### ParseStringToHtmlNode

`ParseStringToHtmlNode(pageSource string) (*html.Node, error)`

Takes a string and returns an `*html.Node`.

```go
// Assuming 'htmlString' is a string of HTML
node, err := goSpider.ParseStringToHtmlNode(htmlString)
```

### Concurrency

#### ParallelRequests

`ParallelRequests(requests []Request, numberOfWorkers int, delay time.Duration, crawlerFunc func(string) (*html.Node, error)) ([]PageSource, error)`

Performs web scraping tasks concurrently with a specified number of workers and a delay between requests. The `crawlerFunc` parameter allows for flexibility in defining the web scraping logic.

- `requests`: A slice of `Request` structures containing the data needed for each request.
- `numberOfWorkers`: The number of concurrent workers to process the requests.
- `delay`: The delay duration between each request to avoid overwhelming the target server.
- `crawlerFunc`: A user-defined function that takes a search string as input and returns the html as `*html.Node`, and an error.

```go
	requests := []goSpider.Request{
		{SearchString: "1017927-35.2023.8.26.0008"},
		{SearchString: "0002396-75.2013.8.26.0201"},
	}

	numberOfWorkers := 1
	delay := 0 * time.Millisecond

	results, err := goSpider.ParallelRequests(requests, numberOfWorkers, delay, Crawler)
```

#### EvaluateParallelRequests

`EvaluateParallelRequests(previousResults []PageSource, crawlerFunc func(string) (*html.Node, error), evaluate func([]PageSource) ([]Request, []PageSource)) ([]PageSource, error)`

EvaluateParallelRequests iterates over a set of previous results, evaluates them using the provided evaluation function, and handles re-crawling of problematic sources until all sources are valid or no further progress can be made.
- `previousResults`: A slice of PageSource objects containing the initial crawl results.
- `crawlerFunc`: A function that takes a string (URL or identifier) and returns a parsed HTML node and an error.
- `evaluate`: A function that takes a slice of PageSource objects and returns two slices:
  1. A slice of Request objects for sources that need to be re-crawled.
  2. A slice of valid PageSource objects.

Returns:
- A slice of valid PageSource objects after all problematic sources have been re-crawled and evaluated.
- An error if there is a failure in the crawling process.

Example usage:

```go
 results, err := EvaluateParallelRequests(resultsFirst, Crawler, Eval)

	func Eval(previousResults []PageSource) ([]Request, []PageSource) {
		var newRequests []Request
		var validResults []PageSource

		for _, result := range previousResults {
			_, err := extractDataCover(result.Page, "")
			if err != nil {
				newRequests = append(newRequests, Request{SearchString: result.Request})
			} else {
				validResults = append(validResults, result)
			}
		}

		return newRequests, validResults
	}
```

#### RemovePageSource

`RemovePageSource(slice []PageSource, s int) []PageSource`

Removes the element at index `s` from a slice of `PageSource` objects.

```go
// Assuming 'results' is a []PageSource
results = goSpider.RemovePageSource(results, 0) // Removes the first element
```

#### RemoveRequest

`RemoveRequest(slice []Request, s int) []Request`

Removes the element at index `s` from a slice of `Request` objects.

```go
// Assuming 'requests' is a []Request
requests = goSpider.RemoveRequest(requests, 0) // Removes the first element
```

## Structs

### Request

```go
type Request struct {
	SearchString string
}
```

### PageSource

```go
type PageSource struct {
	Page    *html.Node
	Request string
	Error   error
}
```
