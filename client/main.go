package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type QuoteResponse struct {
	Bid string `json:"bid"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	// Create a context with timeout of 300 milliseconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel() // Cancel the task, when timeout is reached

	body := getQuoteResponseFromApi(ctx)

	// Print the response body in terminal
	println("-> Body from Request: ", string(body))

	// Try to convert response body to the expected payload
	quote := convertResponsePayload(body)

	// Try to write the result in "cotacao.txt"
	writeResultInFile(quote)
}

func getQuoteResponseFromApi(ctx context.Context) []byte {
	client := http.Client{}

	// Create a request for the server, with the created context
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	doPanicIfAnErrorExist(err)

	// Execute the request
	res, err := client.Do(req)
	doPanicIfAnErrorExist(err)
	defer res.Body.Close()

	// Reads the body response
	body, err := io.ReadAll(res.Body)
	doPanicIfAnErrorExist(err)
	return body
}

func convertResponsePayload(body []byte) QuoteResponse {
	var quote QuoteResponse
	err := json.Unmarshal(body, &quote)
	doPanicIfAnErrorExist(err)
	return quote
}

func writeResultInFile(quote QuoteResponse) {
	// Reads the "cotacao.txt" file
	file, err := os.Create("cotacao.txt")
	doPanicIfAnErrorExist(err)
	defer file.Close()

	// Write in file the current quote value
	_, err = file.WriteString(fmt.Sprintf("Dólar: %v", quote.Bid))
	doPanicIfAnErrorExist(err)
}

func doPanicIfAnErrorExist(err error) {
	if err != nil {
		panic(err)
	}
}
