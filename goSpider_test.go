package goSpider

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestGetPageSource tests fetching the HTML content from a URL
func TestGetPageSource(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	nav.OpenURL("https://www.google.com")
	htmlContent, err := nav.GetPageSource()
	if err != nil {
		t.Errorf("FetchHTML error: %v", err)
	}
	if htmlContent == nil {
		t.Error("FetchHTML returned empty content")
	}
}

// TestClickButton tests clicking a button and waiting for dynamically loaded content
func TestClickButton(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	url := "https://esaj.tjsp.jus.br/cpopg/open.do"

	err := nav.OpenURL(url)
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	err = nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)")
	if err != nil {
		t.Errorf("CheckRadioButton error: %v", err)
	}

	err = nav.FillField("#nuProcessoAntigoFormatado", "1017927-35.2023.8.26.0008")
	if err != nil {
		t.Errorf("filling field error: %v", err)
	}

	err = nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		t.Errorf("ClickButton error: %v", err)
	}
}

// TestNestedElement tests waiting for a nested element to appear after a click
func TestNestedElement(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	url := "https://esaj.tjsp.jus.br/cpopg/open.do"

	err := nav.OpenURL(url)
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	err = nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)")
	if err != nil {
		t.Errorf("CheckRadioButton error: %v", err)
	}

	err = nav.FillField("#nuProcessoAntigoFormatado", "1017927-35.2023.8.26.0008")
	if err != nil {
		t.Errorf("filling field error: %v", err)
	}

	err = nav.ClickButton("#botaoConsultarProcessos")
	if err != nil {
		t.Errorf("ClickButton error: %v", err)
	}

	cUrl, err := nav.GetCurrentURL()
	if err != nil {
		t.Errorf("GetCurrentURL error: %v", err)
	}

	if !strings.Contains(cUrl, "https://esaj.tjsp.jus.br/cpopg/show.do?") {
		t.Errorf("WaitForElement (nested element) error: %s", cUrl)
	}
}

// TestFillFormAndHandleAlert tests filling a form and handling the resulting alert
func TestFillFormAndHandleAlert(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	url := "https://www.camarapassatempo.mg.gov.br/acessocentral/problemas/contato.htm"

	err := nav.OpenURL(url)
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	formData := map[string]string{
		"nome":     "Fulano de Tal",
		"endereco": "Avenida do Contorno",
		"telefone": "11912345678",
		"email":    "null@null.com",
	}

	err = nav.FillForm("body > form", formData)
	if err != nil {
		t.Errorf("FillForm error: %v", err)
	}
}

// TestSelectDropdown tests selecting an option from a dropdown menu
func TestSelectDropdown(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	url := "https://esaj.tjsp.jus.br/cpopg/open.do"

	err := nav.OpenURL(url)
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	err = nav.SelectDropdown("#cbPesquisa", "DOCPARTE")
	if err != nil {
		t.Errorf("SelectDropdown error: %v", err)
	}
}

// TestSelectRadioButton tests selecting a radio button
func TestSelectRadioButton(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	url := "https://esaj.tjsp.jus.br/cpopg/open.do"

	err := nav.OpenURL(url)
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	err = nav.CheckRadioButton("#interna_NUMPROC > div > fieldset > label:nth-child(5)")
	if err != nil {
		t.Errorf("SelectRadioButton error: %v", err)
	}
}

// TestWaitForElement tests waiting for an element to be visible after a delay
func TestWaitForElement(t *testing.T) {
	nav := NewNavigator("")
	defer nav.Close()

	url := "https://esaj.tjsp.jus.br/cpopg/open.do"

	err := nav.OpenURL(url)
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	err = nav.WaitForElement("#interna_NUMPROC > div > fieldset > label:nth-child(5)", 10*time.Second)
	if err != nil {
		t.Errorf("WaitForElement (delayed element) error: %v", err)
	}
}

// TestGetCurrentURL tests extracting the current URL from the browser
func TestGetCurrentURL(t *testing.T) {
	nav := NewNavigator("")

	// Navigate to the main page
	err := nav.OpenURL("https://www.google.com")
	if err != nil {
		t.Errorf("OpenURL error: %v", err)
	}

	// Extract and verify the current URL
	currentURL, err := nav.GetCurrentURL()
	if err != nil {
		t.Errorf("GetCurrentURL error: %v", err)
	}

	expectedURL := "https://www.google.com/"
	if currentURL != expectedURL {
		t.Errorf("Expected URL: %s, but got: %s", expectedURL, currentURL)
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

	numberOfWorkers := 5
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
	nav := NewNavigator("")
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
