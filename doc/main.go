package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func getPage(urlStr string) string {
	log.Printf("Fetching %s ...", urlStr)
	res, err := http.Get(urlStr)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func readTemplate(filename string) string {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return string(fileBytes)
}

func extractComment(body string) (string, string) {

	reFirstComment := regexp.MustCompile("/\\*(.|[\r\n])*?\\*/")

	comment := reFirstComment.FindString(body)
	if comment == "" {
		log.Print("no comment")
	} else {
		body = strings.Replace(body, comment, "", 1)
		comment = strings.Replace(comment, "/* ", "", 1)
		comment = strings.Replace(comment, "*/", "", 1)
	}

	return comment, body
}

func main() {

	var tmplFilename = flag.String("in", "README.md.tmpl", "filename of the template")
	tmplBody := readTemplate(*tmplFilename)

	startTag := "<exampleInclude>"
	endTag := "</exampleInclude>"
	re := regexp.MustCompile(startTag + "[^<]*" + endTag)

	for _, statement := range re.FindAllString(tmplBody, -1) {

		// the example URL
		urlStr := strings.TrimPrefix(statement, startTag)
		urlStr = strings.TrimSuffix(urlStr, endTag)

		// example body
		pageBody := getPage(urlStr)
		exampleComment, exampleCode := extractComment(pageBody)

		exampleStr := ""
		exampleStr += exampleComment + "\n\n"
		exampleStr += "~~~ go\n"
		exampleStr += exampleCode + "\n"
		exampleStr += "~~~\n"

		tmplBody = strings.Replace(tmplBody, statement, exampleStr, -1)
	}
	fmt.Printf("%s", tmplBody)
}
