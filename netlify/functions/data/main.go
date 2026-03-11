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
	ShareC float64 `json:"share_c"`
}

type Account struct {
	Ac         int    `json:"ac"`
	AcName     string `json:"acname"`
	IncomeFlag int    `json:"income_flag"` // 0=expense, 1=income, 2=special
	Category   string `json:"category"`
}

type DataResponse struct {
	Shareholders []Shareholder `json:"shareholders"`
	Accounts     []Account     `json:"accounts"`
}

var shareholders = []Shareholder{
	{1, "Ammi", 0.25, 0.124, 0},
	{2, "Alka", 0.125, 0.146, 0},
	{3, "Jahanzeb", 0.25, 0.292, 0.5},
	{4, "Memoona", 0.125, 0.146, 0},
	{5, "Waleed", 0.25, 0.292, 0.5},
}

var accounts = []Account{
	{100, "BANK DEPOSIT", 2, "0"},
	{101, "GENERAL EXPENSE", 0, "0"},
	{103, "SHAMIM", 0, "1"},
	{104, "NAILA", 0, "1"},
	{105, "AMMI 8 Acer", 0, "1"},
	{106, "MASJID", 0, "1"},
	{107, "ALMEEZAN", 1, "0"},
	{108, "OLD ALADEEL", 1, "0"},
	{109, "NEW ALADEEL", 1, "0"},
	{110, "NEW ALMEEZAN", 1, "0"},
	{111, "RENT", 1, "0"},
	{112, "ZAMEEN FROKHAT", 1, "0"},
	{113, "CNG", 1, "0"},
	{114, "GULZAR", 1, "2"},
	{115, "AZAM", 1, "2"},
	{116, "ABDUL REHMAN", 1, "2"},
	{117, "SHIKOO", 1, "2"},
	{118, "RASHEED", 1, "2"},
	{119, "MUNSHI", 1, "2"},
	{120, "ASHIQ BUT", 1, "2"},
	{121, "ASAD BUTTER", 1, "2"},
	{122, "EMBOD", 1, "0"},
	{123, "MASTER SB", 1, "2"},
	{124, "WELFAIR", 0, "1"},
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
