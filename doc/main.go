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

func main() {

	var tmplFilename = flag.String("in", "README.md.tmpl", "filename of the template")
	tmplBody := readTemplate(*tmplFilename)

	startTag := "<webInclude>"
	endTag := "</webInclude>"
	re := regexp.MustCompile(startTag + "[^<]*" + endTag)
	for _, statement := range re.FindAllString(tmplBody, -1) {

		urlStr := strings.TrimPrefix(statement, startTag)
		urlStr = strings.TrimSuffix(urlStr, endTag)

		pageBody := getPage(urlStr)

		tmplBody = strings.Replace(tmplBody, statement, pageBody, -1)
	}
	fmt.Printf("%s", tmplBody)
}
