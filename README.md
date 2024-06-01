# goSpider Navigation Library
 This Go library provides functions to navigate websites and retrieve information using `chromedp`. It supports basic actions like fetching HTML content, clicking buttons, filling forms, handling alerts, and more complex interactions such as handling AJAX requests and dynamically loaded content.

## Installation

To use this library, you need to install:

```sh
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
	"github.com/DanielFillol/goSpider"
	"log"
)

func main() {
	url := "https://esaj.tjsp.jus.br/sajcas/login"
	nav := goSpider.NewNavigator()
	defer nav.Close()

	err := nav.OpenURL(url)
	if err != nil {
		log.Printf("OpenURL error: %v", err)
	}

	usernameSelector := "#usernameForm"
	passwordSelector := "#passwordForm"
	username := "363.400.878-41"
	password := "Remoto123*"
	loginButtonSelector := "#pbEntrar"
	messageFailedSuccess := "#mensagemRetorno > li"

	err = nav.Login(url, username, password, usernameSelector, passwordSelector, loginButtonSelector, messageFailedSuccess)
	if err != nil {
		log.Printf("Login error: %v", err)
	}

	err = nav.ClickButton("#esajConteudoHome > table:nth-child(7) > tbody > tr > td.esajCelulaDescricaoServicos > a")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
	}

	err = nav.ClickButton("#esajConteudoHome > table:nth-child(3) > tbody > tr > td.esajCelulaDescricaoServicos > a")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
	}

	err = nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
	}

	err = nav.FillField("#nuProcessoAntigoFormatado", "1017927-35.2023.8.26.0008")
	if err != nil {
		log.Printf("filling field error: %v", err)
	}

	err = nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
	}

	err = nav.ClickElement("#linkmovimentacoes")
	if err != nil {
		log.Printf("ClickElement error: %v", err)
	}

	people, err := nav.ExtractTableData("#tablePartesPrincipais")
	if err != nil {
		log.Printf("ExtractTableData error: %v", err)
	}

	movements, err := nav.ExtractTableData("#tabelaTodasMovimentacoes")
	if err != nil {
		log.Printf("ExtractTableData error: %v", err)
	}

	cover, err := nav.ExtractDivText("#containerDadosPrincipaisProcesso", "#maisDetalhes")
	if err != nil {
		log.Printf("ExtractDivText error: %v", err)
	}

	//movements
	for rowIndex, row := range movements {
		fmt.Printf("Row %d:\n", rowIndex)
		for cellIndex, cell := range row {
			fmt.Printf("  Cell %d: %s\n", cellIndex, cell["text"])
			if spans, ok := cell["spans"]; ok {
				for spanIndex, spanText := range spans.(map[int]string) {
					fmt.Printf("    Span %d: %s\n", spanIndex, spanText)
				}
			}
		}
	}

	//people involved on the lawsuit
	for rowIndex, row := range people {
		fmt.Printf("Row %d:\n", rowIndex)
		for cellIndex, cell := range row {
			fmt.Printf("  Cell %d: %s\n", cellIndex, cell["text"])
			if spans, ok := cell["spans"]; ok {
				for spanIndex, spanText := range spans.(map[int]string) {
					fmt.Printf("    Span %d: %s\n", spanIndex, spanText)
				}
			}
		}
	}

	//lawsuit cover data
	for key, value := range cover {
		fmt.Printf("%s: %s\n", key, value)
	}

	fmt.Println(len(people))
	fmt.Println(len(movements))
	fmt.Println(len(cover))

	err = nav.CaptureScreenshot()
	if err != nil {
		log.Printf("Screenshot error: %v", err)
	}

}
```
## Functions
Functions Overview

- NewNavigator() *Navigator
Creates a new instance of the Navigator struct, initializing a new ChromeDP context and logger.
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
