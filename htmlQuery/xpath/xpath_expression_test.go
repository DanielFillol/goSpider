package xpath

import (
	"testing"
)

func Test_descendant_issue(t *testing.T) {
	// Issue #93 https://github.com/antchfx/xpath/issues/93
	/*
	   <div id="wrapper">
	     <span>span one</span>
	     <div>
	       <span>span two</span>
	     </div>
	   </div>
	*/
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.lines = 1
	div.addAttribute("id", "wrapper")
	span := div.createChildNode("span", ElementNode)
	span.lines = 2
	span.createChildNode("span one", TextNode)
	div = div.createChildNode("div", ElementNode)
	div.lines = 3
	span = div.createChildNode("span", ElementNode)
	span.lines = 4
	span.createChildNode("span two", TextNode)

	testXpathElements(t, doc, `//div[@id='wrapper']/descendant::span[1]`, 2)
	testXpathElements(t, doc, `//div[@id='wrapper']//descendant::span[1]`, 2, 4)
}

// https://github.com/antchfx/htmlquery/issues/52

func TestRelativePaths(t *testing.T) {
	testXpathElements(t, bookExample, `//bookstore`, 2)
	testXpathElements(t, bookExample, `//book`, 3, 9, 15, 25)
	testXpathElements(t, bookExample, `//bookstore/book`, 3, 9, 15, 25)
	testXpathTags(t, bookExample, `//book/..`, "bookstore")
	testXpathElements(t, bookExample, `//book[@category="cooking"]/..`, 2)
	testXpathElements(t, bookExample, `//book/year[text() = 2005]/../..`, 2) // bookstore
	testXpathElements(t, bookExample, `//book/year/../following-sibling::*`, 9, 15, 25)
	testXpathCount(t, bookExample, `//bookstore/book/*`, 20)
	testXpathTags(t, htmlExample, "//title/../..", "html")
	testXpathElements(t, htmlExample, "//ul/../p", 19)
}

func TestAbsolutePaths(t *testing.T) {
	testXpathElements(t, bookExample, `bookstore`, 2)
	testXpathElements(t, bookExample, `bookstore/book`, 3, 9, 15, 25)
	testXpathElements(t, bookExample, `(bookstore/book)`, 3, 9, 15, 25)
	testXpathElements(t, bookExample, `bookstore/book[2]`, 9)
	testXpathElements(t, bookExample, `bookstore/book[last()]`, 25)
	testXpathElements(t, bookExample, `bookstore/book[last()]/title`, 26)
	testXpathValues(t, bookExample, `/bookstore/book[last()]/title/text()`, "Learning XML")
	testXpathValues(t, bookExample, `/bookstore/book[@category = "children"]/year`, "2005")
	testXpathElements(t, bookExample, `bookstore/book/..`, 2)
	testXpathElements(t, bookExample, `/bookstore/*`, 3, 9, 15, 25)
	testXpathElements(t, bookExample, `/bookstore/*/title`, 4, 10, 16, 26)
}

func TestAttributes(t *testing.T) {
	testXpathTags(t, htmlExample.FirstChild, "@*", "lang")
	testXpathCount(t, employeeExample, `//@*`, 9)
	testXpathValues(t, employeeExample, `//@discipline`, "web", "DBA", "appdev")
	testXpathCount(t, employeeExample, `//employee/@id`, 3)
}

func TestExpressions(t *testing.T) {
	testXpathElements(t, bookExample, `//book[@category = "cooking"] | //book[@category = "children"]`, 3, 9)
	testXpathCount(t, htmlExample, `//ul/*`, 3)
	testXpathCount(t, htmlExample, `//ul/*/a`, 3)
	// Sequence
	//
	// table/tbody/tr/td/(para, .[not(para)], ..)
}

func TestSequence(t *testing.T) {
	// `//table/tbody/tr/td/(para, .[not(para)],..)`
	testXpathCount(t, htmlExample, `//body/(h1, h2, p)`, 2)
	testXpathCount(t, htmlExample, `//body/(h1, h2, p, ..)`, 3)
}
