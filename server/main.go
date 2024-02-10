package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type USDBRL struct {
	Bid string `json:"bid"`
}

type Quote struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type QuoteResponse struct {
	Bid string `json:"bid"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	// Create WebServer
	http.HandleFunc("/cotacao", getQuote)
	http.ListenAndServe(":8080", nil)
}

func getQuote(w http.ResponseWriter, r *http.Request) {
	// Verify if the http Method is "GET"
	if r.Method != "GET" {
		GivenAnErrorResponse(errors.New("Invalid HTTP Method"), http.StatusNotFound, w)
	}

	log.Println("Starting request ...")
	defer log.Println("... Request finished")

	// Create the Client with context to do the request
	client := http.Client{Timeout: time.Duration(time.Millisecond * 200)}

	// Do the Request
	res, err := client.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	GivenAnErrorResponse(err, http.StatusInternalServerError, w)
	defer res.Body.Close()

	// Try to read the body
	body, err := io.ReadAll(res.Body)
	GivenAnErrorResponse(err, http.StatusInternalServerError, w)

	// Converts it to Struct
	var quote Quote
	err = json.Unmarshal(body, &quote)
	GivenAnErrorResponse(err, http.StatusInternalServerError, w)

	log.Println("Obtained response from API: ", quote)

	successMsg := "Request completed successfully"

	log.Println(successMsg)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(QuoteResponse{Bid: quote.USDBRL.Bid})
}

func GivenAnErrorResponse(err error, status int, w http.ResponseWriter) {
	if err != nil {
		msg := fmt.Sprint("Error: ", err.Error())

		log.Fatalln(msg)
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(ErrorResponse{Message: msg})
	}
}
