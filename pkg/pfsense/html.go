package pfsense

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type htmlTableRow[T any] interface {
	setByHTMLTableRow(int) error
	setByHTMLTableCol(int, string) error
	*T
}

func scrapeHTMLTable[T any, PT htmlTableRow[T]](sel *goquery.Selection) []T {
	var t []T
	sel.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
		var (
			r    T
			err  error
			flag bool
		)

		pr := PT(&r)

		err = pr.setByHTMLTableRow(rowIndex)
		if err != nil {
			return
		}

		row.Find("td").EachWithBreak(func(colIndex int, col *goquery.Selection) bool {
			text := strings.TrimSpace(col.Text())
			if text == "" {
				return true
			}

			err = pr.setByHTMLTableCol(colIndex, text)
			if err != nil {
				flag = true
				return false
			}

			return true
		})

		if flag {
			return
		}

		t = append(t, r)
	})

	return t
}

func scrapeValidationErrors(doc *goquery.Document) error {
	inputErrorList := doc.FindMatcher(goquery.Single("div.input-errors:has(p:contains('input errors')) ul"))

	if inputErrorList.Length() != 0 {
		var inputErrors []string
		inputErrorList.Find("li").Each(func(i int, e *goquery.Selection) {
			inputErrors = append(inputErrors, strings.TrimSpace(e.Text()))
		})
		return fmt.Errorf("%w, '%s'", ErrServerValidation, strings.Join(inputErrors, ", "))
	}
	return nil
}
