package xpath

import (
	"math"
	"testing"
)

// Some test examples from http://zvon.org/comp/r/ref-XPath_2.html

func Test_func_boolean(t *testing.T) {
	testXpathEval(t, emptyExample, `true()`, true)
	testXpathEval(t, emptyExample, `false()`, false)
	testXpathEval(t, emptyExample, `boolean(0)`, false)
	testXpathEval(t, emptyExample, `boolean(null)`, false)
	testXpathEval(t, emptyExample, `boolean(1)`, true)
	testXpathEval(t, emptyExample, `boolean(2)`, true)
	testXpathEval(t, emptyExample, `boolean(true)`, false)
	testXpathEval(t, emptyExample, `boolean(1 > 2)`, false)
	testXpathEval(t, bookExample, `boolean(//*[@lang])`, true)
	testXpathEval(t, bookExample, `boolean(//*[@x])`, false)
}

func Test_func_name(t *testing.T) {
	testXpathEval(t, htmlExample, `name(//html/@lang)`, "lang")
	testXpathEval(t, htmlExample, `name(html/head/title)`, "title")
	testXpathCount(t, htmlExample, `//*[name() = "li"]`, 3)
}

func Test_func_not(t *testing.T) {
	//test_xpath_eval(t, empty_example, `not(0)`, true)
	//test_xpath_eval(t, empty_example, `not(1)`, false)
	testXpathElements(t, employeeExample, `//employee[not(@id = "1")]`, 8, 13)
	testXpathElements(t, bookExample, `//book[not(year = 2005)]`, 15, 25)
	testXpathCount(t, bookExample, `//book[not(title)]`, 0)
}

func Test_func_ceiling_floor(t *testing.T) {
	testXpathEval(t, emptyExample, "ceiling(5.2)", float64(6))
	testXpathEval(t, emptyExample, "floor(5.2)", float64(5))
}

func Test_func_concat(t *testing.T) {
	testXpathEval(t, emptyExample, `concat("1", "2", "3")`, "123")
	//test_xpath_eval(t, empty_example, `concat("Ciao!", ())`, "Ciao!")
	testXpathEval(t, bookExample, `concat(//book[1]/title, ", ", //book[1]/year)`, "Everyday Italian, 2005")
	result := concatFunc(testQuery("a"), testQuery("b"))(nil, nil).(string)
	assertEqual(t, result, "ab")
}

func Test_func_contains(t *testing.T) {
	testXpathEval(t, emptyExample, `contains("tattoo", "t")`, true)
	testXpathEval(t, emptyExample, `contains("tattoo", "T")`, false)
	testXpathEval(t, emptyExample, `contains("tattoo", "ttt")`, false)
	//test_xpath_eval(t, empty_example, `contains("", ())`, true)
	testXpathElements(t, bookExample, `//book[contains(title, "Potter")]`, 9)
	testXpathElements(t, bookExample, `//book[contains(year, "2005")]`, 3, 9)
	assertPanic(t, func() { selectNode(htmlExample, "//*[contains(0, 0)]") })
}

func Test_func_count(t *testing.T) {
	testXpathEval(t, bookExample, `count(//book)`, float64(4))
	testXpathEval(t, bookExample, `count(//book[3]/author)`, float64(5))
}

func Test_func_ends_with(t *testing.T) {
	testXpathEval(t, emptyExample, `ends-with("tattoo", "tattoo")`, true)
	testXpathEval(t, emptyExample, `ends-with("tattoo", "atto")`, false)
	testXpathElements(t, bookExample, `//book[ends-with(@category,'ing')]`, 3)
	testXpathElements(t, bookExample, `//book[ends-with(./price,'.99')]`, 9, 15)
	assertPanic(t, func() { selectNode(htmlExample, `//*[ends-with(0, 0)]`) }) // arg must be start with string
	assertPanic(t, func() { selectNode(htmlExample, `//*[ends-with(name(), 0)]`) })
}

func Test_func_last(t *testing.T) {
	testXpathElements(t, bookExample, `//bookstore[last()]`, 2)
	testXpathElements(t, bookExample, `//bookstore/book[last()]`, 25)
	testXpathElements(t, bookExample, `(//bookstore/book)[last()]`, 25)
	testXpathElements(t, bookExample, `(//bookstore/book[year = 2005])[last()]`, 9)
	testXpathElements(t, bookExample, `//bookstore/book[year = 2005][last()]`, 9)
	testXpathElements(t, htmlExample, `//ul/li[last()]`, 15)
	testXpathElements(t, htmlExample, `(//ul/li)[last()]`, 15)
}

func Test_func_local_name(t *testing.T) {
	testXpathEval(t, bookExample, `local-name(bookstore)`, "bookstore")
	testXpathEval(t, myBookExample, `local-name(//mybook:book)`, "book")
}

func Test_func_starts_with(t *testing.T) {
	testXpathEval(t, employeeExample, `starts-with("tattoo", "tat")`, true)
	testXpathEval(t, employeeExample, `starts-with("tattoo", "att")`, false)
	testXpathElements(t, bookExample, `//book[starts-with(title,'Everyday')]`, 3)
	assertPanic(t, func() { selectNode(htmlExample, `//*[starts-with(0, 0)]`) })
	assertPanic(t, func() { selectNode(htmlExample, `//*[starts-with(name(), 0)]`) })
}

func Test_func_string(t *testing.T) {
	testXpathEval(t, emptyExample, `string(1.23)`, "1.23")
	testXpathEval(t, emptyExample, `string(3)`, "3")
	testXpathEval(t, bookExample, `string(//book/@category)`, "cooking")
}

func Test_func_string_join(t *testing.T) {
	//(t, empty_example, `string-join(('Now', 'is', 'the', 'time', '...'), '')`, "Now is the time ...")
	testXpathEval(t, emptyExample, `string-join("some text", ";")`, "some text")
	testXpathEval(t, bookExample, `string-join(//book/@category, ";")`, "cooking;children;web;web")
}

func Test_func_string_length(t *testing.T) {
	testXpathEval(t, emptyExample, `string-length("Harp not on that string, madam; that is past.")`, float64(45))
	testXpathEval(t, emptyExample, `string-length(normalize-space(' abc '))`, float64(3))
	testXpathEval(t, htmlExample, `string-length(//title/text())`, float64(len("My page")))
	testXpathEval(t, htmlExample, `string-length(//html/@lang)`, float64(len("en")))
	testXpathCount(t, employeeExample, `//employee[string-length(@id) > 0]`, 3) // = //employee[@id]
}

func Test_func_substring(t *testing.T) {
	testXpathEval(t, emptyExample, `substring("motor car", 6)`, " car")
	testXpathEval(t, emptyExample, `substring("metadata", 4, 3)`, "ada")
	//test_xpath_eval(t, empty_example, `substring("12345", 5, -3)`, "") // ?? it should be 1 ??
	//test_xpath_eval(t, empty_example, `substring("12345", 1.5, 2.6)`, "234")
	//test_xpath_eval(t, empty_example, `substring("12345", 0, 3)`, "12") // panic??
	//test_xpath_eval(t, empty_example, `substring("12345", 5, -3)`, "1")
	testXpathEval(t, htmlExample, `substring(//title/child::node(), 1)`, "My page")
	assertPanic(t, func() { selectNode(emptyExample, `substring("12345", 5, -3)`) }) // Should be supported a negative value
	assertPanic(t, func() { selectNode(emptyExample, `substring("12345", 5, "")`) })
}

func Test_func_substring_after(t *testing.T) {
	testXpathEval(t, emptyExample, `substring-after("tattoo", "tat")`, "too")
	testXpathEval(t, emptyExample, `substring-after("tattoo", "tattoo")`, "")
}

func Test_func_substring_before(t *testing.T) {
	testXpathEval(t, emptyExample, `substring-before("tattoo", "attoo")`, "t")
	testXpathEval(t, emptyExample, `substring-before("tattoo", "tatto")`, "")
}

func Test_func_sum(t *testing.T) {
	testXpathEval(t, emptyExample, `sum(1 + 2)`, float64(3))
	testXpathEval(t, emptyExample, `sum(1.1 + 2)`, float64(3.1))
	testXpathEval(t, bookExample, `sum(//book/price)`, float64(149.93))
	testXpathElements(t, bookExample, `//book[sum(./price) > 40]`, 15)
	assertPanic(t, func() { selectNode(htmlExample, `//title[sum('Hello') = 0]`) })
}

func Test_func_translate(t *testing.T) {
	testXpathEval(t, emptyExample, `translate("bar","abc","ABC")`, "BAr")
	testXpathEval(t, emptyExample, `translate("--aaa--","abc-","ABC")`, "AAA")
	testXpathEval(t, emptyExample, `translate("abcdabc", "abc", "AB")`, "ABdAB")
	testXpathEval(t, emptyExample, `translate('The quick brown fox', 'brown', 'red')`, "The quick red fdx")
}

func Test_func_matches(t *testing.T) {
	testXpathEval(t, emptyExample, `matches("abracadabra", "bra")`, true)
	testXpathEval(t, emptyExample, `matches("abracadabra", "(?i)^A.*A$")`, true)
	testXpathEval(t, emptyExample, `matches("abracadabra", "^a.*a$")`, true)
	testXpathEval(t, emptyExample, `matches("abracadabra", "^bra")`, false)
	assertPanic(t, func() { selectNode(htmlExample, `//*[matches()]`) })                   // arg len check failure
	assertPanic(t, func() { selectNode(htmlExample, "//*[matches(substring(), 0)]") })     // first arg processing failure
	assertPanic(t, func() { selectNode(htmlExample, "//*[matches(@href, substring())]") }) // second arg processing failure
	assertPanic(t, func() { selectNode(htmlExample, "//*[matches(@href, 0)]") })           // second arg not string
	assertPanic(t, func() { selectNode(htmlExample, "//*[matches(@href, '[invalid')]") })  // second arg invalid regexp
	// testing unexpected the regular expression.
	_, err := Compile(`//*[matches(., '^[\u0621-\u064AA-Za-z\-]+')]`)
	assertErr(t, err)
	_, err = Compile(`//*[matches(., '//*[matches(., '\w+`)
	assertErr(t, err)
}

func Test_func_number(t *testing.T) {
	testXpathEval(t, emptyExample, `number(10)`, float64(10))
	testXpathEval(t, emptyExample, `number(1.11)`, float64(1.11))
	testXpathEval(t, emptyExample, `number("10") > 10`, false)
	testXpathEval(t, emptyExample, `number("10") = 10`, true)
	testXpathEval(t, emptyExample, `number("123") < 1000`, true)
	testXpathEval(t, emptyExample, `number(//non-existent-node) = 0`, false)
	assertTrue(t, math.IsNaN(MustCompile(`number(//non-existent-node)`).Evaluate(createNavigator(emptyExample)).(float64)))
	assertTrue(t, math.IsNaN(MustCompile(`number("123a")`).Evaluate(createNavigator(emptyExample)).(float64)))
}

func Test_func_position(t *testing.T) {
	testXpathElements(t, bookExample, `//book[position() = 1]`, 3)
	testXpathElements(t, bookExample, `//book[(position() mod 2) = 0]`, 9, 25)
	testXpathElements(t, bookExample, `//book[position() = last()]`, 25)
	testXpathElements(t, bookExample, `//book/*[position() = 1]`, 4, 10, 16, 26)
	// Test Failed
	//test_xpath_elements(t, book_example, `(//book/title)[position() = 1]`, 3)
}

func Test_func_replace(t *testing.T) {
	testXpathEval(t, emptyExample, `replace('aa-bb-cc','bb','ee')`, "aa-ee-cc")
	testXpathEval(t, emptyExample, `replace("abracadabra", "bra", "*")`, "a*cada*")
	testXpathEval(t, emptyExample, `replace("abracadabra", "a", "")`, "brcdbr")
	// The below xpath expressions is not supported yet
	//
	//test_xpath_eval(t, empty_example, `replace("abracadabra", "a.*a", "*")`, "*")
	//test_xpath_eval(t, empty_example, `replace("abracadabra", "a.*?a", "*")`, "*c*bra")
	//test_xpath_eval(t, empty_example, `replace("abracadabra", ".*?", "$1")`, "*c*bra") // error, because the pattern matches the zero-length string
	//test_xpath_eval(t, empty_example, `replace("AAAA", "A+", "b")`, "b")
	//test_xpath_eval(t, empty_example, `replace("AAAA", "A+?", "b")`, "bbb")
	//test_xpath_eval(t, empty_example, `replace("darted", "^(.*?)d(.*)$", "$1c$2")`, "carted")
	//test_xpath_eval(t, empty_example, `replace("abracadabra", "a(.)", "a$1$1")`, "abbraccaddabbra")
}

func Test_func_reverse(t *testing.T) {
	//test_xpath_eval(t, employee_example, `reverse(("hello"))`, "hello") // Not passed
	testXpathElements(t, employeeExample, `reverse(//employee)`, 13, 8, 3)
	testXpathElements(t, employeeExample, `//employee[reverse(.) = reverse(.)]`, 3, 8, 13)
	assertPanic(t, func() { selectNode(htmlExample, "reverse(concat())") }) // invalid node-sets argument.
	assertPanic(t, func() { selectNode(htmlExample, "reverse()") })         //  missing node-sets argument.
}

func Test_func_round(t *testing.T) {
	testXpathEval(t, employeeExample, `round(2.5)`, 3) // int
	testXpathEval(t, employeeExample, `round(2.5)`, 3)
	testXpathEval(t, employeeExample, `round(2.4999)`, 2)
}

func Test_func_namespace_uri(t *testing.T) {
	testXpathEval(t, myBookExample, `namespace-uri(//mybook:book)`, "http://www.contoso.com/books")
	testXpathElements(t, myBookExample, `//*[namespace-uri()='http://www.contoso.com/books']`, 3, 9)
}

func Test_func_normalize_space(t *testing.T) {
	const testStr = "\t    \rloooooooonnnnnnngggggggg  \r \n tes  \u00a0 t strin \n\n \r g "
	const expectedStr = `loooooooonnnnnnngggggggg tes t strin g`
	testXpathEval(t, employeeExample, `normalize-space("`+testStr+`")`, expectedStr)

	testXpathEval(t, emptyExample, `normalize-space(' abc ')`, "abc")
	testXpathEval(t, bookExample, `normalize-space(//book/title)`, "Everyday Italian")
	testXpathEval(t, bookExample, `normalize-space(//book[1]/title)`, "Everyday Italian")
}

func Test_func_lower_case(t *testing.T) {
	testXpathEval(t, emptyExample, `lower-case("ABc!D")`, "abc!d")
	testXpathCount(t, employeeExample, `//name[@from="ca"]`, 0)
	testXpathElements(t, employeeExample, `//name[lower-case(@from) = "ca"]`, 9)
	//test_xpath_eval(t, employee_example, `//employee/name/lower-case(text())`, "opal kole", "max miller", "beccaa moss")
}

func Benchmark_NormalizeSpaceFunc(b *testing.B) {
	b.ReportAllocs()
	const strForNormalization = "\t    \rloooooooonnnnnnngggggggg  \r \n tes  \u00a0 t strin \n\n \r g "
	for i := 0; i < b.N; i++ {
		_ = normalizeSpaceFunc(testQuery(strForNormalization), nil)
	}
}

func Benchmark_ConcatFunc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatFunc(testQuery("a"), testQuery("b"))(nil, nil)
	}
}
