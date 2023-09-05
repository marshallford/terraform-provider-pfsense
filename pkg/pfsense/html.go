package pfsense

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrapeHTMLValidationErrors(doc *goquery.Document) error {
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

func sanitizeHTMLMessage(text string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		return "", err
	}
	sanitize := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return sanitize.ReplaceAllString(doc.Text(), ""), nil
}
