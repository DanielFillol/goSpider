# goSpider Navigation Library
 This Go library provides functions to navigate websites and retrieve information using `chromedp`. It supports basic actions like fetching HTML content, clicking buttons, filling forms, handling alerts, and more complex interactions such as handling AJAX requests and dynamically loaded content.

## Installation

To use this library, you need to install the `chromedp` package:

```sh
go get -u github.com/chromedp/chromedp
go get -u github.com/DanielFillol/goSpider
```

## Usage
Importing the Library
First, import the library in your Go project:
```sh
import "DanielFillol/goSpider"
```
## Example Usage
Here's an example of how to use the library:
```sh
package main

import (
    "fmt"
    "webnav"
    "time"
)

func main() {
    nav := webnav.NewNavigator()
    defer nav.Close()

    url := "https://example.com"
    htmlContent, err := nav.FetchHTML(url)
    if err != nil {
        fmt.Printf("Error fetching HTML: %v\n", err)
        return
    }

    fmt.Println("HTML content fetched.")

    // Example of clicking a button
    err = nav.ClickButton("#exampleButton")
    if err != nil {
        fmt.Printf("Error clicking button: %v\n", err)
        return
    }

    time.Sleep(2 * time.Second) // Wait for the action to complete

    // Example of using a search bar
    err = nav.FillSearchBar("#searchBar", "example query")
    if err != nil {
        fmt.Printf("Error filling search bar: %v\n", err)
        return
    }

    time.Sleep(2 * time.Second) // Wait for the action to complete

    // Example of opening a new tab
    err = nav.OpenNewTab("https://example.com/newtab")
    if err != nil {
        fmt.Printf("Error opening new tab: %v\n", err)
        return
    }

    // Example of extracting links
    links, err := nav.ExtractLinks()
    if err != nil {
        fmt.Printf("Error extracting links: %v\n", err)
        return
    }

    fmt.Println("Links found:")
    for _, link := range links {
        fmt.Println(link)
    }

    // Example of extracting text
    text, err := nav.ExtractText()
    if err != nil {
        fmt.Printf("Error extracting text: %v\n", err)
        return
    }

    fmt.Println("Text content:")
    fmt.Println(text)

    // Example of filling a form
    formData := map[string]string{
        "username": "exampleuser",
        "password": "examplepass",
    }
    err = nav.FillForm("#loginForm", formData)
    if err != nil {
        fmt.Printf("Error filling form: %v\n", err)
        return
    }

    // Example of handling a JavaScript alert
    err = nav.HandleAlert()
    if err != nil {
        fmt.Printf("Error handling alert: %v\n", err)
        return
    }

    time.Sleep(2 * time.Second) // Wait for the action to complete

    // Example of selecting a dropdown option
    err = nav.SelectDropdown("#dropdown", "option1")
    if err != nil {
        fmt.Printf("Error selecting dropdown option: %v\n", err)
        return
    }

    // Example of checking a checkbox
    err = nav.CheckCheckbox("#checkbox")
    if err != nil {
        fmt.Printf("Error checking checkbox: %v\n", err)
        return
    }

    // Example of unchecking a checkbox
    err = nav.UncheckCheckbox("#checkbox")
    if err != nil {
        fmt.Printf("Error unchecking checkbox: %v\n", err)
        return
    }

    // Example of selecting a radio button
    err = nav.SelectRadioButton("#radioButton")
    if err != nil {
        fmt.Printf("Error selecting radio button: %v\n", err)
        return
    }

    // Example of uploading a file
    err = nav.UploadFile("#fileInput", "/path/to/your/file.txt")
    if err != nil {
        fmt.Printf("Error uploading file: %v\n", err)
        return
    }

    time.Sleep(2 * time.Second) // Wait for the action to complete

    // Example of waiting for an element to be visible
    err = nav.WaitForElement("#dynamicElement", 10*time.Second)
    if err != nil {
        fmt.Printf("Error waiting for element: %v\n", err)
        return
    }

    // Example of waiting for AJAX requests to complete
    err = nav.WaitForAJAX(5 * time.Second)
    if err != nil {
        fmt.Printf("Error waiting for AJAX requests: %v\n", err)
        return
    }
}
```
## Functions
- NewNavigator: Creates a new Navigator instance.
- FetchHTML: Fetches the HTML content of a given URL.
- ClickButton: Clicks a button identified by the given selector.
- FillSearchBar: Fills a search bar identified by the given selector and submits the form.
- OpenNewTab: Opens a new tab with the given URL.
- ExtractLinks: Extracts all the links from the current page.
- ExtractText: Extracts all the text content from the current page.
- FillForm: Fills a form with the provided data.
- HandleAlert: Handles a JavaScript alert by accepting it.
- SelectDropdown: Selects an option from a dropdown menu identified by the selector and option value.
- CheckCheckbox: Checks a checkbox identified by the selector.
- UncheckCheckbox: Unchecks a checkbox identified by the selector.
- SelectRadioButton: Selects a radio button identified by the selector.
- UploadFile: Uploads a file to a file input identified by the selector.
- WaitForElement: Waits for an element identified by the selector to be visible.
- WaitForAJAX: Waits for AJAX requests to complete by monitoring the network activity.
- Close: Closes the Navigator instance.
