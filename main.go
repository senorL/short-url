package main

import (
	"fmt"
	"net/http"
)

var urlSwitch = make(map[string]string)
var count = 0

func main() {

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")

		count++
		shortCode := fmt.Sprintf("%d", count)

		urlSwitch[shortCode] = url
		fmt.Fprintf(w, "生成成功！你的短链接是: http://localhost:8080/%s\n", shortCode)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		if urlPath == "/" {
			http.ServeFile(w, r, "index.html")
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
