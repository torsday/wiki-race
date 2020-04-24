package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"torsday.com/wiki-race/puregorace"
)

var port = flag.String("port", "8083", "Port for address")

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/wiki-race/goLang", puregorace.WikiRacePureGoHandler).Methods("GET")
	fmt.Printf("Port: %v \n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, router))
}
