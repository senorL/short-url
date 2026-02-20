package main

import "net/http"

func main() {

	http.Handle("/bili", http.RedirectHandler("https://www.bilibili.com", http.StatusFound))
	http.Handle("/github", http.RedirectHandler("https://www.github.com", http.StatusFound))

	http.ListenAndServe(":8080", nil)
}
