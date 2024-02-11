package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type USDBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type QuoteResponse struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type QuoteHandler struct {
	Db *gorm.DB
}

func NewQuoteHandler(db *gorm.DB) *QuoteHandler {
	return &QuoteHandler{
		Db: db,
	}
}

func main() {
	// Open conection to the Database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	doPanicIfAnErrorExist(err)

	db.AutoMigrate(&USDBRL{})

	quoteHandler := NewQuoteHandler(db)

	// Create WebServer
	http.HandleFunc("/cotacao", quoteHandler.getQuote)
	http.ListenAndServe(":8080", nil)
}

func (h *QuoteHandler) getQuote(w http.ResponseWriter, r *http.Request) {
	// Verify if the http Method is "GET"
	if r.Method != "GET" {
		GivenAnErrorResponse(errors.New("Invalid HTTP Method"), http.StatusNotFound, w)
		return
	}

	log.Println("Starting request ...")
	defer log.Println("... Request finished")

	body, err := getQuoteFromApi(w)
	if err != nil {
		GivenAnErrorResponse(err, http.StatusInternalServerError, w)
		return
	}

	quote, err := convertResponsePayload(body, w)
	if err != nil {
		GivenAnErrorResponse(err, http.StatusInternalServerError, w)
		return
	}

	log.Println("Obtained response from API: ", quote)

	// Save to database
	err = saveQuoteOnDatabase(quote, h.Db, w)
	if err != nil {
		GivenAnErrorResponse(err, http.StatusInternalServerError, w)
		return
	}

	successMsg := "Request completed successfully"

	log.Println(successMsg)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quote)
}

func getQuoteFromApi(w http.ResponseWriter) ([]byte, error) {
	// Create Context with 200ms timeout for Api Request
	ctx, cancelReqCtx := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancelReqCtx()

	// Create the Client
	client := http.Client{}

	// Do the Request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	// Executes the request with context
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Try to read the body
	return io.ReadAll(res.Body)
}

func convertResponsePayload(body []byte, w http.ResponseWriter) (*USDBRL, error) {
	// Converts it to Struct
	var quote QuoteResponse
	err := json.Unmarshal(body, &quote)
	return &quote.USDBRL, err
}

func saveQuoteOnDatabase(quote *USDBRL, db *gorm.DB, w http.ResponseWriter) error {
	// Create Context with 200ms timeout for Database Query
	ctx, cancelDbQuery := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelDbQuery()

	err := db.WithContext(ctx).Create(quote).Error
	if err != nil {
		return err
	}

	log.Println("Quote saved on database: ", quote)
	return nil
}

func GivenAnErrorResponse(err error, status int, w http.ResponseWriter) {
	msg := fmt.Sprint("Error: ", err.Error())

	log.Println(msg)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: msg})
}

func doPanicIfAnErrorExist(err error) {
	if err != nil {
		panic(err)
	}
}
