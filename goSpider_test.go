package goSpider

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var server *http.Server
var nav *Navigator

// TestMain sets up the test environment and tears it down after the tests are done
func TestMain(m *testing.M) {
	// Start the test server
	server = StartTestServer()

	// Give the server a moment to start
	time.Sleep(1 * time.Second)

	// Create a new navigator instance
	nav = NewNavigator()

	// Run the tests
	code := m.Run()

	// Close the navigator
	nav.Close()

	// Close the test server
	server.Close()

	// Exit with the appropriate code
	os.Exit(code)
}

// TestFetchHTML tests fetching the HTML content from a URL
func TestFetchHTML(t *testing.T) {
	htmlContent, err := nav.FetchHTML("http://localhost:8080")
	if err != nil {
		t.Errorf("FetchHTML error: %v", err)
	}
	if htmlContent == "" {
		t.Error("FetchHTML returned empty content")
	}
}

// TestClickButton tests clicking a button and waiting for dynamically loaded content
func TestClickButton(t *testing.T) {
	err := nav.ClickButton("#exampleButton")
	if err != nil {
		t.Errorf("ClickButton error: %v", err)
	}

	err = nav.WaitForElement("#dynamicElement", 10*time.Second)
	if err != nil {
		t.Errorf("WaitForElement error: %v", err)
	}
}

// TestNestedElement tests waiting for a nested element to appear after a click
func TestNestedElement(t *testing.T) {
	err := nav.ClickButton("#exampleButton")
	if err != nil {
		t.Errorf("ClickButton error: %v", err)
	}

	err = nav.WaitForElement("#nestedElement", 10*time.Second)
	if err != nil {
		t.Errorf("WaitForElement (nested element) error: %v", err)
	}
}

// TestFillFormAndHandleAlert tests filling a form and handling the resulting alert
func TestFillFormAndHandleAlert(t *testing.T) {
	formData := map[string]string{
		"username": "test_user",
		"password": "test_pass",
	}
	err := nav.FillForm("#loginForm", formData)
	if err != nil {
		t.Errorf("FillForm error: %v", err)
	}

	err = nav.HandleAlert()
	if err != nil {
		t.Errorf("HandleAlert error: %v", err)
	}
}

// TestSelectDropdown tests selecting an option from a dropdown menu
func TestSelectDropdown(t *testing.T) {
	err := nav.SelectDropdown("#dropdown", "option2")
	if err != nil {
		t.Errorf("SelectDropdown error: %v", err)
	}
}

// TestSelectRadioButton tests selecting a radio button
func TestSelectRadioButton(t *testing.T) {
	err := nav.CheckRadioButton("#radioButton")
	if err != nil {
		t.Errorf("SelectRadioButton error: %v", err)
	}
}

// TestWaitForElement tests waiting for an element to be visible after a delay
func TestWaitForElement(t *testing.T) {
	err := nav.WaitForElement("#delayedElement", 10*time.Second)
	if err != nil {
		t.Errorf("WaitForElement (delayed element) error: %v", err)
	}
}

// TestGetCurrentURL tests extracting the current URL from the browser
func TestGetCurrentURL(t *testing.T) {
	// Navigate to the main page
	htmlContent, err := nav.FetchHTML("http://localhost:8080")
	if err != nil {
		t.Errorf("FetchHTML error: %v", err)
	}
	if htmlContent == "" {
		t.Error("FetchHTML returned empty content")
	}

	// Extract and verify the current URL
	currentURL, err := nav.GetCurrentURL()
	if err != nil {
		t.Errorf("GetCurrentURL error: %v", err)
	}

	expectedURL := "http://localhost:8080/"
	if currentURL != expectedURL {
		t.Errorf("Expected URL: %s, but got: %s", expectedURL, currentURL)
	}

	// Navigate to page 2
	err = nav.ClickButton("#linkToPage2")
	if err != nil {
		t.Errorf("ClickButton error: %v", err)
	}

	time.Sleep(2 * time.Second) // Wait for navigation to complete

	// Extract and verify the current URL for page 2
	currentURL, err = nav.GetCurrentURL()
	if err != nil {
		t.Errorf("GetCurrentURL error: %v", err)
	}

	expectedURL = "http://localhost:8080/page2"
	if currentURL != expectedURL {
		t.Errorf("Expected URL: %s, but got: %s", expectedURL, currentURL)
	}
}

func TestParallelRequests(t *testing.T) {
	users := []Requests{
		{ProcessNumber: "1017927-35.2023.8.26.0008"},
		{ProcessNumber: "0002396-75.2013.8.26.0201"},
		{ProcessNumber: "1551285-50.2021.8.26.0477"},
		{ProcessNumber: "0015386-82.2013.8.26.0562"},
		{ProcessNumber: "0007324-95.2015.8.26.0590"},
		{ProcessNumber: "1545639-85.2023.8.26.0090"},
		{ProcessNumber: "1557599-09.2021.8.26.0090"},
		{ProcessNumber: "1045142-72.2021.8.26.0002"},
		{ProcessNumber: "0208591-43.2009.8.26.0004"},
		{ProcessNumber: "1017927-35.2023.8.26.0008"},
	}

	numberOfWorkers := 3
	duration := 2 * time.Second

	results, err := ParallelRequests(users, numberOfWorkers, duration, Crawler)
	if err != nil {
		log.Printf("GetCurrentURL error: %v", err)
	}

	log.Println("Finish Parallel Requests!")

	var found []string
	for _, u := range users {
		for _, result := range results {
			for _, value := range result.Cover {
				if value == u.ProcessNumber {
					found = append(found, value)
				}
			}
		}
	}

	if len(found) != len(users) {
		t.Errorf("Expected %d results, but got %d, List results: %v", len(users), len(found), found)
	}

}

func Crawler(d string) (map[string]string, []map[int]map[string]interface{}, []map[int]map[string]interface{}, error) {
	log.Printf("Crawling: %s", d)

	url := "https://esaj.tjsp.jus.br/sajcas/login"
	nav := NewNavigator()
	defer nav.Close()

	err := nav.OpenURL(url)
	if err != nil {
		log.Printf("OpenURL error: %v", err)
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}

	err = nav.ClickButton("#esajConteudoHome > table:nth-child(7) > tbody > tr > td.esajCelulaDescricaoServicos > a")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
		return nil, nil, nil, err
	}

	err = nav.ClickButton("#esajConteudoHome > table:nth-child(3) > tbody > tr > td.esajCelulaDescricaoServicos > a")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
		return nil, nil, nil, err
	}

	err = nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)")
	if err != nil {
		log.Printf("CheckRadioButton error: %v", err)
		return nil, nil, nil, err
	}

	err = nav.FillField("#nuProcessoAntigoFormatado", d)
	if err != nil {
		log.Printf("filling field error: %v", err)
		return nil, nil, nil, err
	}

	err = nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		log.Printf("ClickButton error: %v", err)
		return nil, nil, nil, err
	}

	err = nav.ClickElement("#linkmovimentacoes")
	if err != nil {
		log.Printf("ClickElement error: %v", err)
		return nil, nil, nil, err
	}

	people, err := nav.ExtractTableData("#tablePartesPrincipais")
	if err != nil {
		log.Printf("ExtractTableData error: %v", err)
		return nil, nil, nil, err
	}

	movements, err := nav.ExtractTableData("#tabelaTodasMovimentacoes")
	if err != nil {
		log.Printf("ExtractTableData error: %v", err)
		return nil, nil, nil, err
	}

	cover, err := nav.ExtractDivText("#containerDadosPrincipaisProcesso", "#maisDetalhes")
	if err != nil {
		log.Printf("ExtractDivText error: %v", err)
		return nil, nil, nil, err
	}

	return cover, movements, people, nil
}
