package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("its working")
	mux := http.NewServeMux()
	serv := http.Server{
		Addr: ":8080",
		Handler: mux,
	}	
	serv.ListenAndServe()

	fmt.Println("does this print?")
}
