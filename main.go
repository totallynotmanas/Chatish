package main

import (
	"fmt"
	"net/http"
)

func main(){
	// Serve the static folder (for JS, CSS later)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w,r, "templeates/index.html")
	})

	fmt.Println("Starting Chatish Server at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}