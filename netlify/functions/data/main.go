package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Shareholder struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	ShareA float64 `json:"share_a"`
	ShareB float64 `json:"share_b"`
}

type Account struct {
	Ac          int    `json:"ac"`
	AcName      string `json:"acname"`
	IncomeFlag  int    `json:"income_flag"` // 0=expense, 1=income, 2=special
	Category    string `json:"category"`
}

type DataResponse struct {
	Shareholders []Shareholder `json:"shareholders"`
	Accounts     []Account     `json:"accounts"`
}

var shareholders = []Shareholder{
	{1, "Ammi", 0.20, 0.25},
	{2, "Alka", 0.20, 0.25},
	{3, "Jahanzeb", 0.20, 0.125},
	{4, "Memoona", 0.20, 0.125},
	{5, "Waleed", 0.20, 0.25},
}

var accounts = []Account{
	{100, "BANK DEPOSIT", 2, "0"},
	{101, "GENERAL INCOME", 1, "0"},
	{102, "GENERAL EXPENSE", 0, "0"},
	{103, "SALARY", 0, "1"},
	{104, "RENT", 0, "1"},
	{105, "ELECTRICITY", 0, "1"},
	{106, "GAS", 0, "1"},
	{107, "WATER", 0, "1"},
	{108, "INTERNET", 0, "1"},
	{109, "PHONE", 0, "1"},
	{110, "FOOD & GROCERIES", 0, "1"},
	{111, "TRANSPORT", 0, "1"},
	{112, "MEDICAL", 0, "1"},
	{113, "EDUCATION", 0, "1"},
	{114, "CLOTHING", 0, "1"},
	{115, "HOUSEHOLD", 0, "1"},
	{116, "ENTERTAINMENT", 0, "1"},
	{117, "MAINTENANCE", 0, "1"},
	{118, "LOAN GIVEN", 0, "2"},
	{119, "LOAN RECEIVED", 1, "2"},
	{120, "LOAN REPAID", 0, "2"},
	{121, "LOAN RECOVERY", 1, "2"},
	{122, "MISCELLANEOUS", 0, "0"},
}

func corsHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Content-Type":                 "application/json",
	}
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{StatusCode: 200, Headers: corsHeaders()}, nil
	}

	resp := DataResponse{
		Shareholders: shareholders,
		Accounts:     accounts,
	}
	out, _ := json.Marshal(resp)
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: corsHeaders(), Body: string(out)}, nil
}

func main() {
	lambda.Start(handler)
}
