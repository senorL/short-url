package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var urlSwitch = make(map[string]string)
var count = 0

type ResultData struct {
	ShortCode string
	Original  string
}

func main() {

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		count++

		shortCode := fmt.Sprintf("%d", count)
		urlSwitch[shortCode] = url
		t, _ := template.ParseFiles("success.html")

		data := ResultData{
			ShortCode: shortCode,
			Original:  url,
		}
		t.Execute(w, data)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		if urlPath == "/" {
			funcMap := template.FuncMap{"turl": TruncateURL}
			t, _ := template.New("index.html").Funcs(funcMap).ParseFiles("index.html")
			t.Execute(w, urlSwitch)
		} else {
			shortCode := urlPath[1:]

			newUrl, ok := urlSwitch[shortCode]
			if ok {
				http.Redirect(w, r, newUrl, http.StatusFound)
			} else {
				http.NotFound(w, r)
			}
		}

	})

	http.ListenAndServe(":8080", nil)
}

func TruncateURL(url string) string {
	if len(url) > 30 {
		url = url[:30] + "..."
	}
	return url
}
