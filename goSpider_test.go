package goSpider

import (
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
