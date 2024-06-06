package xpath

import (
	"testing"
)

func TestLogical(t *testing.T) {
	testXpathElements(t, bookExample, `//book[1 + 1]`, 9)
	testXpathElements(t, bookExample, `//book[1 * 2]`, 9)
	testXpathElements(t, bookExample, `//book[5 div 2]`, 9) // equal to `//book[2]`
	testXpathElements(t, bookExample, `//book[3 div 2]`, 3)
	testXpathElements(t, bookExample, `//book[3 - 2]`, 3)
	testXpathElements(t, bookExample, `//book[price > 35]`, 15, 25)
	testXpathElements(t, bookExample, `//book[price >= 30]`, 3, 15, 25)
	testXpathElements(t, bookExample, `//book[price < 30]`, 9)
	testXpathElements(t, bookExample, `//book[price <= 30]`, 3, 9)
	testXpathElements(t, bookExample, `//book[count(author) > 1]`, 15)
	testXpathElements(t, bookExample, `//book[position() mod 2 = 0]`, 9, 25)
}

func TestPositions(t *testing.T) {
	testXpathElements(t, employeeExample, `/empinfo/employee[2]`, 8)
	testXpathElements(t, employeeExample, `//employee[position() = 2]`, 8)
	testXpathElements(t, employeeExample, `/empinfo/employee[2]/name`, 9)
	testXpathElements(t, employeeExample, `//employee[position() > 1]`, 8, 13)
	testXpathElements(t, employeeExample, `//employee[position() <= 2]`, 3, 8)
	testXpathElements(t, employeeExample, `//employee[last()]`, 13)
	testXpathElements(t, employeeExample, `//employee[position() = last()]`, 13)
	testXpathElements(t, bookExample, `//book[@category = "web"][2]`, 25)
	testXpathElements(t, bookExample, `(//book[@category = "web"])[2]`, 25)
}

func TestPredicates(t *testing.T) {
	testXpathElements(t, employeeExample, `//employee[name]`, 3, 8, 13)
	testXpathElements(t, employeeExample, `/empinfo/employee[@id]`, 3, 8, 13)
	testXpathElements(t, bookExample, `//book[@category = "web"]`, 15, 25)
	testXpathElements(t, bookExample, `//book[author = "J K. Rowling"]`, 9)
	testXpathElements(t, bookExample, `//book[./author/text() = "J K. Rowling"]`, 9)
	testXpathElements(t, bookExample, `//book[year = 2005]`, 3, 9)
	testXpathElements(t, bookExample, `//year[text() = 2005]`, 6, 12)
	testXpathElements(t, employeeExample, `/empinfo/employee[1][@id=1]`, 3)
	testXpathElements(t, employeeExample, `/empinfo/employee[@id][2]`, 8)
}

func TestOperators(t *testing.T) {
	testXpathElements(t, employeeExample, `//designation[@discipline and @experience]`, 5, 10)
	testXpathElements(t, employeeExample, `//designation[@discipline or @experience]`, 5, 10, 15)
	testXpathElements(t, employeeExample, `//designation[@discipline | @experience]`, 5, 10, 15)
	testXpathElements(t, employeeExample, `/empinfo/employee[@id != "2" ]`, 3, 13)
	testXpathElements(t, employeeExample, `/empinfo/employee[@id and @id = "2"]`, 8)
	testXpathElements(t, employeeExample, `/empinfo/employee[@id = "1" or @id = "2"]`, 3, 8)
}

func TestNestedPredicates(t *testing.T) {
	testXpathElements(t, employeeExample, `//employee[./name[@from]]`, 8)
	testXpathElements(t, employeeExample, `//employee[.//name[@from = "CA"]]`, 8)
}
