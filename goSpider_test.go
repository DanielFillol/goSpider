package goSpider

import (
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Start a local server to serve the mock HTML page
func startTestServer() *httptest.Server {
	handler := http.FileServer(http.Dir("server"))
	server := httptest.NewServer(handler)
	return server
}

// Setup Navigator for tests
func setupNavigator(t *testing.T) *Navigator {
	nav := NewNavigator("", true)
	nav.SetTimeOut(600 * time.Millisecond)
	t.Cleanup(nav.Close)
	return nav
}

// Test functions
func TestGetElementAttribute(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	a, err := nav.GetElementAttribute("#divInfraCaptcha > div", "data-sitekey")
	if err != nil {
		t.Fatalf("Error on GetElementAttribute: %v", err)
	}

	if a == "" {
		t.Error("Expected a non-empty attribute value")
	}

	fmt.Println(a)
}

func TestSwitchToFrame(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := NewNavigator("", true)
	defer nav.Close()

	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(server.URL+"/test.html"),
		chromedp.WaitVisible("iframe#test-iframe"),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}

	err = nav.SwitchToFrame("iframe#test-iframe")
	if err != nil {
		t.Fatalf("Failed to switch to iframe: %v", err)
	}

	var iframeContent string
	err = chromedp.Run(nav.Ctx,
		chromedp.Text("p", &iframeContent),
	)
	if err != nil {
		t.Fatalf("Failed to get iframe content: %v", err)
	}

	if iframeContent != "Iframe Content" {
		t.Fatalf("Unexpected iframe content: %s", iframeContent)
	} else {
		fmt.Println(iframeContent)
	}
}

func TestSwitchToFrameAndDefaultContent(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := NewNavigator("", true)
	defer nav.Close()

	err := chromedp.Run(nav.Ctx,
		chromedp.Navigate(server.URL+"/test.html"),
		chromedp.WaitVisible("iframe#test-iframe"),
	)
	if err != nil {
		t.Fatalf("Failed to navigate to test page: %v", err)
	}

	err = nav.SwitchToFrame("iframe#test-iframe")
	if err != nil {
		t.Fatalf("Failed to switch to iframe: %v", err)
	}

	var iframeContent string
	err = chromedp.Run(nav.Ctx,
		chromedp.Text("p", &iframeContent),
	)
	if err != nil {
		t.Fatalf("Failed to get iframe content: %v", err)
	}

	if iframeContent != "Iframe Content" {
		t.Fatalf("Unexpected iframe content: %s", iframeContent)
	}

	err = nav.SwitchToDefaultContent()
	if err != nil {
		t.Fatalf("Failed to switch to default content: %v", err)
	}

	var mainContent string
	err = chromedp.Run(nav.Ctx,
		chromedp.Text("h1", &mainContent),
	)
	if err != nil {
		t.Fatalf("Failed to get main content: %v", err)
	}

	if mainContent != "Main Content" {
		t.Fatalf("Unexpected main content: %s", mainContent)
	} else {
		fmt.Println(mainContent)
	}
}

func TestGetCurrentURL(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	currentURL, err := nav.GetCurrentURL()
	if err != nil {
		t.Fatalf("GetCurrentURL error: %v", err)
	}

	expectedURL := server.URL + "/test.html"
	if !strings.Contains(currentURL, expectedURL) {
		t.Errorf("Expected URL to contain: %s, but got: %s", expectedURL, currentURL)
	}
}

func TestLogin(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.Login(server.URL+"/test.html", "username", "password", "#txtUsuario", "#pwdSenha", "#sbmEntrar", "")
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
}

func TestCaptureScreenshot(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	err = nav.CaptureScreenshot("test_screenshot")
	if err != nil {
		t.Fatalf("CaptureScreenshot error: %v", err)
	}
}

func TestReloadPage(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	err = nav.ReloadPage(3)
	if err != nil {
		t.Fatalf("ReloadPage error: %v", err)
	}
}

func TestGetPageSource(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	htmlContent, err := nav.GetPageSource()
	if err != nil {
		t.Fatalf("FetchHTML error: %v", err)
	}
	if htmlContent == nil {
		t.Error("FetchHTML returned empty content")
	}
}

func TestWaitForElement(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	err := nav.WaitForElement("#radioOption2", 10*time.Second)
	if err != nil {
		t.Fatalf("WaitForElement (delayed element) error: %v", err)
	}
}

func TestClickButton(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	err := nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		t.Fatalf("ClickButton error: %v", err)
	}

	cUrl, err := nav.GetCurrentURL()
	if err != nil {
		t.Fatalf("GetCurrentURL error: %v", err)
	}

	expectedURL := "https://esaj.tjsp.jus.br/cpopg/show.do?"
	if !strings.Contains(cUrl, expectedURL) {
		t.Errorf("Expected URL to contain: %s, but got: %s", expectedURL, cUrl)
	}
}

func TestUnsafeClickButton(t *testing.T) {
	server := startTestServer()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	err := nav.UnsafeClickButton("#botaoConsultarProcessos")
	if err != nil {
		t.Fatalf("ClickButton error: %v", err)
	}

	cUrl, err := nav.GetCurrentURL()
	if err != nil {
		t.Fatalf("GetCurrentURL error: %v", err)
	}

	expectedURL := "https://esaj.tjsp.jus.br/cpopg/show.do?"
	if !strings.Contains(cUrl, expectedURL) {
		t.Errorf("Expected URL to contain: %s, but got: %s", expectedURL, cUrl)
	}
}

func TestSelectRadioButton(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	err := nav.CheckRadioButton("#radioOption2")
	if err != nil {
		t.Fatalf("SelectRadioButton error: %v", err)
	}
}

func TestUncheckRadioButton(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	err = nav.UncheckRadioButton("#radioOption2")
	if err != nil {
		t.Fatalf("UncheckRadioButton error: %v", err)
	}
}

func TestFillField(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	err = nav.FillField("#nrProcessoInput", "1000113-34.2018.5.02.0386")
	if err != nil {
		t.Fatalf("FillField error: %v", err)
	}
}

func TestUnsafeFillField(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	err = nav.UnsafeFillField("#nrProcessoInput", "1000113-34.2018.5.02.0386")
	if err != nil {
		t.Fatalf("FillField error: %v", err)
	}
}

func TestExtractLinks(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	// Allow some time for the page to fully load
	time.Sleep(2 * time.Second)

	links, err := nav.ExtractLinks()
	if err != nil {
		t.Fatalf("ExtractLinks error: %v", err)
	}

	expectedLinks := []string{
		"https://www.example.com/",
		"https://www.google.com/",
		"https://www.bing.com/",
	}

	for _, expectedLink := range expectedLinks {
		found := false
		for _, link := range links {
			if link == expectedLink {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected link not found: %s", expectedLink)
		}
	}
}

func TestFillForm(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	formData := map[string]string{
		"nome":     "Fulano de Tal",
		"endereco": "Avenida do Contorno",
		"telefone": "11912345678",
		"email":    "null@null.com",
	}

	err = nav.FillForm("#contactForm", formData)
	if err != nil {
		t.Fatalf("FillForm error: %v", err)
	}
}

func TestHandleAlert(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	err = nav.HandleAlert()
	if err != nil {
		t.Fatalf("HandleAlert error: %v", err)
	}
}

func TestFillFormAndHandleAlert(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	formData := map[string]string{
		"nome":     "Fulano de Tal",
		"endereco": "Avenida do Contorno",
		"telefone": "11912345678",
		"email":    "null@null.com",
	}

	err := nav.FillForm("#contactForm", formData)
	if err != nil {
		t.Fatalf("FillForm error: %v", err)
	}
}

func TestSelectDropdown(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	err := nav.SelectDropdown("#cbPesquisa", "DOCPARTE")
	if err != nil {
		t.Fatalf("SelectDropdown error: %v", err)
	}
}

func TestExecuteScript(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	script := "document.title = 'New Title';"
	err = nav.ExecuteScript(script)
	if err != nil {
		t.Fatalf("ExecuteScript error: %v", err)
	}

	title, err := nav.EvaluateScript("document.title")
	if err != nil {
		t.Fatalf("EvaluateScript error: %v", err)
	}

	if title != "New Title" {
		t.Errorf("Expected title to be 'New Title', but got: %v", title)
	}
}

func TestGetElement(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Fatalf("OpenURL error: %v", err)
	}

	content, err := nav.GetElement("#screenshotPlaceholder")
	if err != nil {
		t.Fatalf("GetElement error: %v", err)
	}

	expectedContent := "Placeholder for Screenshot"
	if content != expectedContent {
		t.Errorf("Expected content to be: %s, but got: %s", expectedContent, content)
	}
}

func TestNestedElement(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	nav.OpenURL(server.URL + "/test.html")

	err := nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		t.Fatalf("ClickButton error: %v", err)
	}

	cUrl, err := nav.GetCurrentURL()
	if err != nil {
		t.Fatalf("GetCurrentURL error: %v", err)
	}

	if !strings.Contains(cUrl, "https://esaj.tjsp.jus.br/cpopg/show.do?") {
		t.Errorf("WaitForElement (nested element) error: %s", cUrl)
	}
}

func TestSaveImageBase64(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	runTest := func(headless bool) {
		nav := setupNavigator(t)
		defer nav.Close()

		err := nav.OpenURL(server.URL + "/test.html")
		if err != nil {
			t.Errorf("OpenURL error: %v", err)
			return
		}

		err = nav.WaitForElement("#nrProcessoInput", 20*time.Second)
		if err != nil {
			t.Errorf("WaitForElement error: %v", err)
			return
		}

		err = nav.FillField("#nrProcessoInput", "1000113-34.2018.5.02.0386")
		if err != nil {
			t.Errorf("FillField error: %v", err)
			return
		}

		err = nav.ClickButton("#btnPesquisar")
		if err != nil {
			t.Errorf("ClickButton error: %v", err)
			return
		}

		err = nav.WaitForElement("#imagemCaptcha", 20*time.Second)
		if err != nil {
			t.Errorf("WaitForElement error: %v", err)
			return
		}

		outputPath := filepath.Join(os.TempDir(), "image.png")
		_, err = nav.SaveImageBase64("#imagemCaptcha", outputPath, "data:image/png;base64,")
		if err != nil {
			t.Errorf("SaveImageBase64 error: %v", err)
			return
		}
	}

	t.Run("Headless Mode", func(t *testing.T) {
		runTest(true)
	})
}

func TestMakeElementVisible(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	defer nav.Close()

	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
		return
	}

	id, err := nav.GetElementAttribute("#divInfraCaptcha > div > iframe", "data-hcaptcha-widget-id")
	if err != nil {
		t.Errorf("GetElementAttribute error: %v", err)
		return
	}

	err = nav.MakeElementVisible("#h-captcha-response-" + id)
	if err != nil {
		t.Errorf("MakeElementVisible error: %v", err)
		return
	}

	err = nav.WaitForElement("#h-captcha-response-"+id, nav.Timeout)
	if err != nil {
		t.Errorf("MakeElementVisible error: %v", err)
		return
	}

	err = nav.FillField("#h-captcha-response-"+id, "54203432300")
	if err != nil {
		t.Errorf("FillField error: %v", err)
		return
	}
}

func TestPrintHtml(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	nav := setupNavigator(t)
	defer nav.Close()

	err := nav.OpenURL(server.URL + "/test.html")
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	ps, err := nav.GetPageSource()
	if err != nil {
		t.Errorf("GetPageSource error: %v", err)
	}

	s, err := PrintHtml(ps)
	if err != nil {
		t.Errorf("PrintHtml error: %v", err)
	}

	fmt.Println(s)

}

func TestDatepicker(t *testing.T) {
	nav := NewNavigator("", false)

	err := nav.OpenURL("https://www.tjrs.jus.br/buscas/jurisprudencia/?conteudo_busca=ementa_completa&q_palavra_chave=&aba=jurisprudencia&q=&conteudo_busca=ementa_completa")
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
		return
	}

	err = nav.WaitForElement("#datas_julgamento_publicacao_div > div:nth-child(2) > div.col-md-4.col-xs-12.dataPublicacaoContainer > div > div > div.recuo_esquerdo_10px.col-md-5.col-xs-5 > div > input", time.Minute)
	if err != nil {
		t.Errorf("WaitForElement error: %v", err)
		return
	}

	err = nav.Datepicker("01/01/2000", "#datas_julgamento_publicacao_div > div:nth-child(2) > div.col-md-4.col-xs-12.dataPublicacaoContainer > div > div > div.recuo_esquerdo_10px.col-md-5.col-xs-5 > div > span > i", "#ui-datepicker-div > div > a.ui-datepicker-prev.ui-corner-all > span", "//*[@id=\"ui-datepicker-div\"]/table/tbody/tr", "#ui-datepicker-div > table > tbody > tr:nth-child")
	if err != nil {
		t.Errorf("Datepicker error: %v", err)
		return
	}

	err = nav.Datepicker("31/12/2000", "#datas_julgamento_publicacao_div > div:nth-child(2) > div.col-md-4.col-xs-12.dataPublicacaoContainer > div > div > div:nth-child(3) > div > span", "#ui-datepicker-div > div > a.ui-datepicker-prev.ui-corner-all > span", "//*[@id=\"ui-datepicker-div\"]/table/tbody/tr", "#ui-datepicker-div > table > tbody > tr:nth-child")
	if err != nil {
		t.Errorf("Datepicker error: %v", err)
		return
	}

	start := time.Now()
	for {
		if time.Since(start) > time.Minute {
			break
		}
	}
}

// Won't pass on test because 2FA requires input on the terminal by the user, for that reason alone the test will fail
//// TestLoginGoogle tests google single logon
//func TestLoginGoogle(t *testing.T) {
//	profilePath := "/Users/USER_NAME/Library/Application Support/Google/Chrome/Profile 2\""
//	nav := NewNavigator(profilePath)
//	defer nav.Close()
//
//	err := nav.LoginWithGoogle("", "")
//	if err != nil {
//		t.Errorf("LoginWithGoogle error: %v", err)
//	}
//
//}

//Full Crawlers

func TestParallelRequests(t *testing.T) {
	users := []Request{
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

	numberOfWorkers := 10
	duration := 0 * time.Millisecond

	results, err := ParallelRequests(users, numberOfWorkers, duration, Crawler)
	if err != nil {
		log.Printf("ParallelRequests error: %v", err)
	}

	if len(results) != len(users) {
		t.Errorf("Expected %d results, but got %d, List results: %v, error: %v", len(users), 0, len(results), err)
	}

	log.Println("Finish Parallel Request!")

}

func TestRequestsDataStruct(t *testing.T) {
	users := []Request{
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
	duration := 500 * time.Millisecond

	resultsFirst, err := ParallelRequests(users, numberOfWorkers, duration, Crawler)
	if err != nil {
		t.Errorf("Expected %d results, but got %d, List results: %v", len(users), 0, len(resultsFirst))
	}

	results, err := EvaluateParallelRequests(resultsFirst, Crawler, Eval)
	if err != nil {
		t.Errorf("Expected %d results, but got %d, List results: %v", len(users), 0, len(results))
	}

	type Lawsuit struct {
		Cover     Cover
		Persons   []Person
		Movements []Movement
	}
	var lawsuits []Lawsuit
	for _, result := range results {
		// Cover
		c, err := extractDataCover(result.Page, "//*[@id=\"numeroProcesso\"]", "//*[@id=\"labelSituacaoProcesso\"]", "//*[@id=\"classeProcesso\"]", "//*[@id=\"assuntoProcesso\"]", "//*[@id=\"foroProcesso\"]", "//*[@id=\"varaProcesso\"]", "//*[@id=\"juizProcesso\"]", "//*[@id=\"dataHoraDistribuicaoProcesso\"]", "//*[@id=\"numeroControleProcesso\"]", "//*[@id=\"areaProcesso\"]/span", "//*[@id=\"valorAcaoProcesso\"]")
		if err != nil {
			t.Errorf("ExtractDataCover error: %v", err)
		}
		// Persons
		p, err := extractDataPerson(result.Page, "//*[@id=\"tableTodasPartes\"]/tbody/tr", "td[1]/span", "td[2]/text()", "\n")
		if err != nil {
			p, err = extractDataPerson(result.Page, "//*[@id=\"tablePartesPrincipais\"]/tbody/tr", "td[1]/text()", "td[2]/text()", "\n")
			if err != nil {
				t.Errorf("Expected some person but got none: %v", err.Error())
			}
		}
		// Movements
		m, err := extractDataMovement(result.Page, "//*[@id=\"tabelaTodasMovimentacoes\"]/tr", "\n")
		if err != nil {
			t.Errorf("Expected some movement but got none: %v", err.Error())
		}

		lawsuits = append(lawsuits, Lawsuit{
			Cover:     c,
			Persons:   p,
			Movements: m,
		})
	}

	if len(lawsuits) != len(users) {
		t.Errorf("Expected %d lawsuits, but got %d", len(users), len(lawsuits))
	}

	fmt.Println(lawsuits)

}

func Eval(previousResults []PageSource) ([]Request, []PageSource) {
	var newRequests []Request
	var validResults []PageSource

	for _, result := range previousResults {
		_, err := extractDataCover(result.Page, "//*[@id=\"numeroProcesso\"]", "//*[@id=\"labelSituacaoProcesso\"]", "//*[@id=\"classeProcesso\"]", "//*[@id=\"assuntoProcesso\"]", "//*[@id=\"foroProcesso\"]", "//*[@id=\"varaProcesso\"]", "//*[@id=\"juizProcesso\"]", "//*[@id=\"dataHoraDistribuicaoProcesso\"]", "//*[@id=\"numeroControleProcesso\"]", "//*[@id=\"areaProcesso\"]/span", "//*[@id=\"valorAcaoProcesso\"]")
		if err != nil {
			newRequests = append(newRequests, Request{SearchString: result.Request})
		} else {
			validResults = append(validResults, result)
		}
	}

	return newRequests, validResults
}

func Crawler(d string) (*html.Node, error) {
	nav := NewNavigator("", true)
	defer nav.Close()

	url := "https://esaj.tjsp.jus.br/cpopg/open.do"

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

	_, err = nav.WaitPageLoad()
	if err != nil {
		log.Printf("WaitPageLoad error: %v", err)
		return nil, err
	}

	pageSource, err := nav.GetPageSource()
	if err != nil {
		log.Printf("GetPageSource error: %v", err)
		return nil, err
	}

	return pageSource, nil
}

type Cover struct {
	Title       string
	Tag         string
	Class       string
	Subject     string
	Location    string
	Unit        string
	Judge       string
	InitialDate string
	Control     string
	Field       string
	Value       string
	Error       string
}

func extractDataCover(pageSource *html.Node, xpathTitle string, xpathTag string, xpathClass string, xpathSubject string, xpathLocation string, xpathUnit string, xpathJudge string, xpathInitDate string, xpathControl string, xpathField string, xpathValue string) (Cover, error) {
	var i int //count errors
	title, err := ExtractText(pageSource, xpathTitle, "                                                            ")
	if err != nil {
		log.Println("error extracting title")
	}

	tag, err := ExtractText(pageSource, xpathTag, "")
	if err != nil {
		i++
		log.Println("error extracting tag")
	}

	class, err := ExtractText(pageSource, xpathClass, "")
	if err != nil {
		i++
		log.Println("error extracting class")
	}

	subject, err := ExtractText(pageSource, xpathSubject, "")
	if err != nil {
		i++
		log.Println("error extracting subject")
	}

	location, err := ExtractText(pageSource, xpathLocation, "")
	if err != nil {
		i++
		log.Println("error extracting location")
	}

	unit, err := ExtractText(pageSource, xpathUnit, "")
	if err != nil {
		i++
		log.Println("error extracting unit")
	}

	judge, err := ExtractText(pageSource, xpathJudge, "")
	if err != nil {
		i++
		log.Println("error extracting existJudge")
	}

	initDate, err := ExtractText(pageSource, xpathInitDate, "")
	if err != nil {
		i++
		log.Println("error extracting initDate")
	}

	control, err := ExtractText(pageSource, xpathControl, "")
	if err != nil {
		i++
		log.Println("error extracting control")
	}

	field, err := ExtractText(pageSource, xpathField, "")
	if err != nil {
		log.Println("error extracting field")
	}

	value, err := ExtractText(pageSource, xpathValue, "R$         ")
	if err != nil {
		i++
		log.Println("error extracting field value")
	}

	var e string
	if err != nil {
		e = err.Error()
	}

	if i >= 5 {
		return Cover{}, fmt.Errorf("too many errors: %d", i)
	}

	return Cover{
		Title:       title,
		Tag:         tag,
		Class:       class,
		Subject:     subject,
		Location:    location,
		Unit:        unit,
		Judge:       judge,
		InitialDate: initDate,
		Control:     control,
		Field:       field,
		Value:       value,
		Error:       e,
	}, nil
}

type Person struct {
	Pole    string
	Name    string
	Lawyers []string
}

func extractDataPerson(pageSource *html.Node, xpathPeople string, xpathPole string, xpathLawyer string, dirt string) ([]Person, error) {
	Pole, err := FindNodes(pageSource, xpathPeople)
	if err != nil {
		return nil, err
	}

	var personas []Person
	for i, person := range Pole {
		pole, err := ExtractText(person, xpathPole, dirt)
		if err != nil {
			return nil, errors.New("error extract data person, pole not found: " + err.Error())
		}

		var name string
		_, err = FindNodes(person, xpathPeople+"["+strconv.Itoa(i)+"]/td[2]")
		if err != nil {
			name, err = ExtractText(person, "td[2]/text()", dirt)
			if err != nil {
				return nil, errors.New("error extract data person, name not found: " + err.Error())
			}
		} else {
			name, err = ExtractText(person, "td[2]/text()["+strconv.Itoa(1)+"]", dirt)
			if err != nil {
				return nil, errors.New("error extract data person, name not found: " + err.Error())
			}
		}

		var lawyers []string
		ll, err := FindNodes(person, xpathLawyer)
		if err != nil {
			lawyers = append(lawyers, "no lawyer found")
		}
		for j := range ll {
			n, err := ExtractText(person, "td[2]/text()["+strconv.Itoa(j+1)+"]", dirt)
			if err != nil {
				return nil, errors.New("error extract data person, lawyer not  found: " + err.Error())
			}
			lawyers = append(lawyers, n)
		}

		p := Person{
			Pole:    pole,
			Name:    name,
			Lawyers: lawyers,
		}

		personas = append(personas, p)
	}

	return personas, nil
}

type Movement struct {
	Date  string
	Title string
	Text  string
}

func extractDataMovement(pageSource *html.Node, node string, dirt string) ([]Movement, error) {
	xpathTable := node

	tableRows, err := ExtractTable(pageSource, xpathTable)
	if err != nil {
		return nil, err
	}

	if len(tableRows) > 0 {
		var allMovements []Movement
		for _, row := range tableRows {
			date, err := ExtractText(row, "td[1]", dirt)
			if err != nil {
				return nil, errors.New("error extracting table date: " + err.Error())
			}
			title, err := ExtractText(row, "td[3]", dirt)
			if err != nil {
				return nil, errors.New("error extracting table title: " + err.Error())
			}
			text, err := ExtractText(row, "td[3]/span", dirt)
			if err != nil {
				return nil, errors.New("error extracting table text: " + err.Error())
			}

			mv := Movement{
				Date:  strings.ReplaceAll(date, "\t", ""),
				Title: strings.ReplaceAll(strings.ReplaceAll(title, text, ""), dirt, ""),
				Text:  strings.TrimSpace(strings.ReplaceAll(text, "\t", "")),
			}

			allMovements = append(allMovements, mv)
		}
		return allMovements, nil
	}

	return nil, errors.New("error table: could not find any movements")
}
