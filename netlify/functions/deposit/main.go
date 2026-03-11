package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const firebaseURL = "https://shah-hisab-default-rtdb.firebaseio.com"

type DepositRequest struct {
	Ammi     float64 `json:"ammi"`
	Alka     float64 `json:"alka"`
	Jahanzeb float64 `json:"jahanzeb"`
	Memoona  float64 `json:"memoona"`
	Waleed   float64 `json:"waleed"`
	Total    float64 `json:"total"`
}

type DepositRecord struct {
	Tstamp   string  `json:"tstamp"`
	Ammi     float64 `json:"ammi"`
	Alka     float64 `json:"alka"`
	Jahanzeb float64 `json:"jahanzeb"`
	Memoona  float64 `json:"memoona"`
	Waleed   float64 `json:"waleed"`
	Total    float64 `json:"total"`
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

	if req.HTTPMethod != "POST" {
		return events.APIGatewayProxyResponse{StatusCode: 405, Headers: corsHeaders(), Body: `{"error":"method not allowed"}`}, nil
	}

	var dep DepositRequest
	if err := json.Unmarshal([]byte(req.Body), &dep); err != nil {
		return errResp(400, "invalid JSON: "+err.Error()), nil
	}

	record := DepositRecord{
		Tstamp:   time.Now().Format("2006-01-02 15:04:05"),
		Ammi:     dep.Ammi,
		Alka:     dep.Alka,
		Jahanzeb: dep.Jahanzeb,
		Memoona:  dep.Memoona,
		Waleed:   dep.Waleed,
		Total:    dep.Total,
	}

	recJSON, _ := json.Marshal(record)
	resp, err := http.Post(
		firebaseURL+"/deposits.json",
		"application/json",
		strings.NewReader(string(recJSON)),
	)
	if err != nil {
		return errResp(500, "firebase write failed: "+err.Error()), nil
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	out, _ := json.Marshal(record)
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
