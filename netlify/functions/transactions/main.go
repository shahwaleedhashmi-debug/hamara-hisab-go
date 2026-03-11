package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const firebaseURL = "https://shah-hisab-default-rtdb.firebaseio.com"

type Shareholder struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	ShareA float64 `json:"share_a"`
	ShareB float64 `json:"share_b"`
	ShareC float64 `json:"share_c"`
}

var shareholders = []Shareholder{
	{1, "ammi", 0.25, 0.124, 0},
	{2, "alka", 0.125, 0.146, 0},
	{3, "jahanzeb", 0.25, 0.292, 0.5},
	{4, "memoona", 0.125, 0.146, 0},
	{5, "waleed", 0.25, 0.292, 0.5},
}

type Transaction struct {
	ID               interface{} `json:"id,omitempty"`
	FirebaseKey      string      `json:"firebase_key,omitempty"`
	Trs              int         `json:"trs"`
	Tstamp           string      `json:"tstamp"`
	Des              string      `json:"des"`
	Amount           float64     `json:"amount"`
	Ac               int         `json:"ac"`
	IncomeExpense    string      `json:"income_expense"`
	CashCredit       string      `json:"cash_credit"`
	CommonIndividual string      `json:"common_individual"`
	Ammi             float64     `json:"ammi"`
	Alka             float64     `json:"alka"`
	Jahanzeb         float64     `json:"jahanzeb"`
	Memoona          float64     `json:"memoona"`
	Waleed           float64     `json:"waleed"`
}

type NewTransactionRequest struct {
	Des              string  `json:"des"`
	Amount           float64 `json:"amount"`
	Ac               int     `json:"ac"`
	IncomeExpense    string  `json:"income_expense"`
	CashCredit       string  `json:"cash_credit"`
	CommonIndividual string  `json:"common_individual"`
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func calcSplit(amount float64, ci string) (ammi, alka, jahanzeb, memoona, waleed float64) {
	switch ci {
	case "A":
		// Equal split: Ammi 25%, Alka 12.5%, Jahanzeb 25%, Memoona 12.5%, Waleed 25%
		ammi = round2(amount * 0.25)
		alka = round2(amount * 0.125)
		jahanzeb = round2(amount * 0.25)
		memoona = round2(amount * 0.125)
		waleed = round2(amount * 0.25)
	case "B":
		// B split: Ammi 12.4%, Alka 14.6%, Jahanzeb 29.2%, Memoona 14.6%, Waleed 29.2%
		ammi = round2(amount * 0.124)
		alka = round2(amount * 0.146)
		jahanzeb = round2(amount * 0.292)
		memoona = round2(amount * 0.146)
		waleed = round2(amount * 0.292)
	case "C":
		// C split: Jahanzeb 50%, Waleed 50%
		jahanzeb = round2(amount * 0.5)
		waleed = round2(amount * 0.5)
	case "1":
		ammi = amount
	case "2":
		alka = amount
	case "3":
		jahanzeb = amount
	case "4":
		memoona = amount
	case "5":
		waleed = amount
	}
	return
}

func corsHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Content-Type":                 "application/json",
	}
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{StatusCode: 200, Headers: corsHeaders()}, nil
	}

	if req.HTTPMethod == "GET" {
		return handleGet()
	}
	if req.HTTPMethod == "POST" {
		return handlePost(req.Body)
	}

	return events.APIGatewayProxyResponse{StatusCode: 405, Headers: corsHeaders(), Body: `{"error":"method not allowed"}`}, nil
}

func handleGet() (events.APIGatewayProxyResponse, error) {
	resp, err := http.Get(firebaseURL + "/txns.json")
	if err != nil {
		return errResp(500, "firebase fetch failed: "+err.Error()), nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Firebase returns a map of key→txn; convert to array
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawMap); err != nil || rawMap == nil {
		return events.APIGatewayProxyResponse{StatusCode: 200, Headers: corsHeaders(), Body: "[]"}, nil
	}

	txns := make([]Transaction, 0, len(rawMap))
	for key, val := range rawMap {
		var t Transaction
		if err := json.Unmarshal(val, &t); err == nil {
			t.FirebaseKey = key
			txns = append(txns, t)
		}
	}

	out, _ := json.Marshal(txns)
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: corsHeaders(), Body: string(out)}, nil
}

func handlePost(body string) (events.APIGatewayProxyResponse, error) {
	var req NewTransactionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return errResp(400, "invalid JSON: "+err.Error()), nil
	}

	ammi, alka, jahanzeb, memoona, waleed := calcSplit(req.Amount, req.CommonIndividual)

	txn := Transaction{
		Trs:              int(time.Now().UnixMilli()),
		Tstamp:           time.Now().Format("2006-01-02 15:04:05"),
		Des:              req.Des,
		Amount:           req.Amount,
		Ac:               req.Ac,
		IncomeExpense:    req.IncomeExpense,
		CashCredit:       req.CashCredit,
		CommonIndividual: req.CommonIndividual,
		Ammi:             ammi,
		Alka:             alka,
		Jahanzeb:         jahanzeb,
		Memoona:          memoona,
		Waleed:           waleed,
	}

	txnJSON, _ := json.Marshal(txn)

	resp, err := http.Post(
		firebaseURL+"/txns.json",
		"application/json",
		strings.NewReader(string(txnJSON)),
	)
	if err != nil {
		return errResp(500, "firebase write failed: "+err.Error()), nil
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var fbResp map[string]string
	json.Unmarshal(respBody, &fbResp)
	txn.FirebaseKey = fbResp["name"]

	out, _ := json.Marshal(txn)
	return events.APIGatewayProxyResponse{StatusCode: 201, Headers: corsHeaders(), Body: string(out)}, nil
}

func errResp(code int, msg string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Headers:    corsHeaders(),
		Body:       fmt.Sprintf(`{"error":%q}`, msg),
	}
}

func main() {
	lambda.Start(handler)
}
