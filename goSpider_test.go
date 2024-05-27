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

// TestFillSearchBar tests filling a search bar and submitting the form
func TestFillSearchBar(t *testing.T) {
	err := nav.FillSearchBar("#searchBar", "test query")
	if err != nil {
		t.Errorf("FillSearchBar error: %v", err)
	}
}

// TestFillFormAndHandleAlert tests filling a form and handling the resulting alert
func TestFillFormAndHandleAlert(t *testing.T) {
	formData := map[string]string{
		"username": "testuser",
		"password": "testpass",
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

// TestCheckbox tests checking and unchecking a checkbox
func TestCheckbox(t *testing.T) {
	err := nav.CheckCheckbox("#checkbox")
	if err != nil {
		t.Errorf("CheckCheckbox error: %v", err)
	}

	err = nav.UncheckCheckbox("#checkbox")
	if err != nil {
		t.Errorf("UncheckCheckbox error: %v", err)
	}
}

// TestSelectRadioButton tests selecting a radio button
func TestSelectRadioButton(t *testing.T) {
	err := nav.SelectRadioButton("#radioButton")
	if err != nil {
		t.Errorf("SelectRadioButton error: %v", err)
	}
}

// TestUploadFile tests uploading a file
func TestUploadFile(t *testing.T) {
	err := nav.UploadFile("#fileInput", "testfile.txt")
	if err != nil {
		t.Errorf("UploadFile error: %v", err)
	}
}

// TestWaitForElement tests waiting for an element to be visible after a delay
func TestWaitForElement(t *testing.T) {
	err := nav.WaitForElement("#delayedElement", 10*time.Second)
	if err != nil {
		t.Errorf("WaitForElement (delayed element) error: %v", err)
	}
}

// TestWaitForAJAX tests waiting for AJAX requests to complete
func TestWaitForAJAX(t *testing.T) {
	err := nav.WaitForAJAX(10 * time.Second)
	if err != nil {
		t.Errorf("WaitForAJAX error: %v", err)
	}
}
