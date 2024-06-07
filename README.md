# goSpider Navigation Library
 This Go library provides functions to navigate websites and retrieve information using `chromedp`. It supports basic actions like fetching HTML content, clicking buttons, filling forms, handling alerts, and more complex interactions such as handling AJAX requests and dynamically loaded content.

## Installation

To use this library, you need to install:

```sh
go get github.com/chromedp/chromedp
go get github.com/DanielFillol/goSpider
```

## Usage
Importing the Library
First, import the library in your Go project:
```sh
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
	users := []goSpider.Requests{
		{SearchString: "1017927-35.2023.8.26.0008"},
		{SearchString: "0002396-75.2013.8.26.0201"},
		{SearchString: "1551285-50.2021.8.26.0477"},
		{SearchString: "0015386-82.2013.8.26.0562"},
		{SearchString: "0007324-95.2015.8.26.0590"},
		{SearchString: "1545639-85.2023.8.26.0090"},
		{SearchString: "1557599-09.2021.8.26.0090"},
		{SearchString: "1045142-72.2021.8.26.0002"},
		{SearchString: "0208591-43.2009.8.26.0004"},
		{SearchString: "1024511-70.2022.8.26.0003"},
	}

	numberOfWorkers := 1
	duration := 0 * time.Millisecond

	results, err := goSpider.ParallelRequests(users, numberOfWorkers, duration, Crawler)
	if err != nil {
		log.Println("Expected %d results, but got %d, List results: %v", len(users), 0, len(results))
	}

	log.Println("Finish Parallel Requests!")

	fmt.Println(len(results))
}

func Crawler(d string) (*html.Node, error) {
	url := "https://esaj.tjsp.jus.br/cpopg/open.do"
	nav := goSpider.NewNavigator("", true)

	err := nav.OpenURL(url)
	if err != nil {
		log.Printf("OpenURL error: %v", err)
		return nil, err
	}

	err = nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)")
	if err != nil {
		log.Printf("CheckRadioButton error: %v", err)
		return nil, err
	}

	err = nav.FillField("#nuProcessoAntigoFormatado", d)
	if err != nil {
		log.Printf("filling field error: %v", err)
		return nil, err
	}

	err = nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
		return nil, err
	}

	err = nav.WaitForElement("#tabelaUltimasMovimentacoes > tr:nth-child(1) > td.dataMovimentacao", 15*time.Second)
	if err != nil {
		log.Printf("WaitForElement error: %v", err)
		return nil, err
	}

	pageSource, err := nav.GetPageSource()
	if err != nil {
		log.Printf("GetPageSource error: %v", err)
		return nil, err
	}

	return pageSource, nil
}


```

## Functions
Functions Overview

- NewNavigator(profilePath string, headless bool) *Navigator
Creates a new instance of the Navigator struct, initializing a new ChromeDP context and logger.
profilePath: the path to chrome profile defined by the user;can be passed as an empty string
headless: if false will show chrome UI
```go
nav := goSpider.NewNavigator()
```
- Close()
Closes the Navigator instance and releases resources.
```go
nav.Close()
```
- OpenNewTab(url string) error
Opens a new browser tab with the specified URL.
```go
err := nav.OpenNewTab("https://www.example.com")
```
- OpenURL(url string) error
Opens the specified URL in the current browser context.
```go
err := nav.OpenURL("https://www.example.com")
```
- GetCurrentURL() (string, error)
Returns the current URL of the browser.
```go
currentURL, err := nav.GetCurrentURL()
```
- Login(url, username, password, usernameSelector, passwordSelector, loginButtonSelector string, messageFailedSuccess string) error
Logs into a website using the provided credentials and selectors.
```go
err := nav.Login("https://www.example.com/login", "username", "password", "#username", "#password", "#login-button", "Login failed")
```
- CaptureScreenshot() error
Captures a screenshot of the current browser window and saves it as screenshot.png.
```go
err := nav.CaptureScreenshot()
```
- GetElement(selector string) (string, error)
Retrieves the text content of an element specified by the selector.
```go
text, err := nav.GetElement("#elementID")
```
- WaitForElement(selector string, timeout time.Duration) error
Waits for an element specified by the selector to be visible within the given timeout.
```go
err := nav.WaitForElement("#elementID", 5*time.Second)
```
- ClickButton(selector string) error
Clicks a button specified by the selector.
```go
err := nav.ClickButton("#buttonID")
```
- ClickElement(selector string) error
Clicks an element specified by the selector.
```go
err := nav.ClickElement("#elementID")
```
- CheckRadioButton(selector string) error
Selects a radio button specified by the selector.
```go
err := nav.CheckRadioButton("#radioButtonID")
```
- UncheckRadioButton(selector string) error
Unchecks a checkbox specified by the selector.
```go
err := nav.UncheckRadioButton("#checkboxID")
```
- FillField(selector string, value string) error
Fills a field specified by the selector with the provided value.
```go
err := nav.FillField("#fieldID", "value")
```
- ExtractTableData(selector string) ([]map[int]map[string]interface{}, error)
Extracts data from a table specified by the selector.
```go
tableData, err := nav.ExtractTableData("#tableID")
```
- ExtractDivText(parentSelectors ...string) (map[string]string, error)
Extracts text content from divs specified by the parent selectors.
```go
textData, err := nav.ExtractDivText("#parent1", "#parent2")
```
- FetchHTML(url string) (string, error)
Fetches the HTML content of the specified URL.
```go
htmlContent, err := nav.FetchHTML("https://www.example.com")
```
- ExtractLinks() ([]string, error)
Extracts all links from the current page.
```go
links, err := nav.ExtractLinks()
```
- FillForm(formSelector string, data map[string]string) error
Fills out a form specified by the selector with the provided data and submits it.
```go
formData := map[string]string{
    "username": "myUsername",
    "password": "myPassword",
}
err := nav.FillForm("#loginForm", formData)
```
- HandleAlert() error
Handles JavaScript alerts by accepting them.
```go
err := nav.HandleAlert()
```
- SelectDropdown(selector, value string) error
Selects an option in a dropdown specified by the selector and value.
```go
err := nav.SelectDropdown("#dropdownID", "optionValue")
```
- FindNodes(node *html.Node, nodeExpression string) ([]*html.Node, error) 
extracts nodes content from nodes specified by the parent selectors
```go
nodeData, err := goSpider.FindNode(pageSource,"#parent1")
```
- ExtractText(node *html.Node, nodeExpression string, Dirt string) (string, error)
```go
textData, err := goSpider.ExtractText(pageSource,"#parent1", "\n")
```
- func ExtractTable(pageSource *html.Node, tableRowsExpression string) ([]*html.Node, error)
```go
tableData, err := goSpider.ExtractTableData(pageSource,"#tableID")
```
- ParallelRequests(requests []Requests, numberOfWorkers int, duration time.Duration, crawlerFunc func(string) (map[string]string, []map[int]map[string]interface{}, []map[int]map[string]interface{}, error)) ([]ResponseBody, error) Performs web scraping tasks concurrently with a specified number of workers and a delay between requests. The crawlerFunc parameter allows for flexibility in defining the web scraping logic. Parameters:
requests: A slice of Requests structures containing the data needed for each request.
numberOfWorkers: The number of concurrent workers to process the requests.
duration: The delay duration between each request to avoid overwhelming the target server.
crawlerFunc: A user-defined function that takes a process number as input and returns cover data, movements, people, and an error.
```go
	users := []goSpider.Requests{
		{SearchString: "1017927-35.2023.8.26.0008"},
		{SearchString: "0002396-75.2013.8.26.0201"},
		{SearchString: "1551285-50.2021.8.26.0477"},
		{SearchString: "0015386-82.2013.8.26.0562"},
		{SearchString: "0007324-95.2015.8.26.0590"},
		{SearchString: "1545639-85.2023.8.26.0090"},
		{SearchString: "1557599-09.2021.8.26.0090"},
		{SearchString: "1045142-72.2021.8.26.0002"},
		{SearchString: "0208591-43.2009.8.26.0004"},
		{SearchString: "1024511-70.2022.8.26.0003"},
	}

	numberOfWorkers := 1
	duration := 0 * time.Millisecond

	results, err := goSpider.ParallelRequests(users, numberOfWorkers, duration, Crawler)
```
- EvaluateParallelRequests(previousResults []PageSource, crawlerFunc func(string) (*html.Node, error), evaluate func([]PageSource) ([]Request, []PageSource)) ([]PageSource, error)
  EvaluateParallelRequests iterates over a set of previous results, evaluates them using the provided evaluation function,
  and handles re-crawling of problematic sources until all sources are valid or no further progress can be made.
  Parameters:
   - previousResults: A slice of PageSource objects containing the initial crawl results.
   - crawlerFunc: A function that takes a string (URL or identifier) and returns a parsed HTML node and an error.
   - evaluate: A function that takes a slice of PageSource objects and returns two slices:
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
- func LoginWithGoogle(email, password string) error
performs the Google login on https://accounts.google.com. The email and password are required for loggin and the 2FA code is passed on prompt.
```go
err := nav.LoginWithGoogle("yor_login", "your_password")
```

