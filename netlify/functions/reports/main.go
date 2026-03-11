package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const firebaseURL = "https://shah-hisab-default-rtdb.firebaseio.com"

type Transaction struct {
	FirebaseKey      string  `json:"firebase_key,omitempty"`
	Trs              int64   `json:"trs"`
	Tstamp           string  `json:"tstamp"`
	Des              string  `json:"des"`
	Amount           float64 `json:"amount"`
	Ac               int     `json:"ac"`
	IncomeExpense    string  `json:"income_expense"`
	CashCredit       string  `json:"cash_credit"`
	CommonIndividual string  `json:"common_individual"`
	Ammi             float64 `json:"ammi"`
	Alka             float64 `json:"alka"`
	Jahanzeb         float64 `json:"jahanzeb"`
	Memoona          float64 `json:"memoona"`
	Waleed           float64 `json:"waleed"`
}

type ShareholderBalance struct {
	Name     string  `json:"name"`
	Income   float64 `json:"income"`
	Expense  float64 `json:"expense"`
	Deposits float64 `json:"deposits"`
	Balance  float64 `json:"balance"`
}

type YearlySummary struct {
	Year    string  `json:"year"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

type ReportsResponse struct {
	TotalIncome       float64              `json:"total_income"`
	TotalExpense      float64              `json:"total_expense"`
	CashOnHand        float64              `json:"cash_on_hand"`
	TxnCount          int                  `json:"txn_count"`
	Last10            []Transaction        `json:"last_10"`
	Shareholders      []ShareholderBalance `json:"shareholders"`
	YearlySummaries   []YearlySummary      `json:"yearly_summaries"`
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
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

	// Fetch transactions
	txns, err := fetchTxns()
	if err != nil {
		return errResp(500, err.Error()), nil
	}

	// Fetch deposits
	deposits, err := fetchDeposits()
	if err != nil {
		deposits = map[string]map[string]float64{}
	}

	// Sort by Trs descending
	sort.Slice(txns, func(i, j int) bool { return txns[i].Trs > txns[j].Trs })

	var totalInc, totalExp float64
	yearMap := map[string]*YearlySummary{}
	shIncome := map[string]float64{"ammi": 0, "alka": 0, "jahanzeb": 0, "memoona": 0, "waleed": 0}
	shExpense := map[string]float64{"ammi": 0, "alka": 0, "jahanzeb": 0, "memoona": 0, "waleed": 0}

	for _, t := range txns {
		year := ""
		if len(t.Tstamp) >= 4 {
			year = t.Tstamp[:4]
		}
		if _, ok := yearMap[year]; !ok {
			yearMap[year] = &YearlySummary{Year: year}
		}

		if t.IncomeExpense == "income" {
			totalInc += t.Amount
			yearMap[year].Income += t.Amount
			shIncome["ammi"] += t.Ammi
			shIncome["alka"] += t.Alka
			shIncome["jahanzeb"] += t.Jahanzeb
			shIncome["memoona"] += t.Memoona
			shIncome["waleed"] += t.Waleed
		} else {
			totalExp += t.Amount
			yearMap[year].Expense += t.Amount
			shExpense["ammi"] += t.Ammi
			shExpense["alka"] += t.Alka
			shExpense["jahanzeb"] += t.Jahanzeb
			shExpense["memoona"] += t.Memoona
			shExpense["waleed"] += t.Waleed
		}
	}

	// Compute yearly net
	years := make([]YearlySummary, 0, len(yearMap))
	for _, y := range yearMap {
		y.Net = round2(y.Income - y.Expense)
		y.Income = round2(y.Income)
		y.Expense = round2(y.Expense)
		years = append(years, *y)
	}
	sort.Slice(years, func(i, j int) bool { return years[i].Year > years[j].Year })

	// Shareholder deposits
	shDeposits := map[string]float64{"ammi": 0, "alka": 0, "jahanzeb": 0, "memoona": 0, "waleed": 0}
	for _, d := range deposits {
		for k := range shDeposits {
			if v, ok := d[k]; ok {
				shDeposits[k] += v
			}
		}
	}

	shNames := []string{"ammi", "alka", "jahanzeb", "memoona", "waleed"}
	shBalances := make([]ShareholderBalance, 0, 5)
	for _, name := range shNames {
		inc := round2(shIncome[name])
		exp := round2(shExpense[name])
		dep := round2(shDeposits[name])
		shBalances = append(shBalances, ShareholderBalance{
			Name:     name,
			Income:   inc,
			Expense:  exp,
			Deposits: dep,
			Balance:  round2(inc - exp - dep),
		})
	}

	last10 := txns
	if len(last10) > 10 {
		last10 = last10[:10]
	}

	report := ReportsResponse{
		TotalIncome:     round2(totalInc),
		TotalExpense:    round2(totalExp),
		CashOnHand:      round2(totalInc - totalExp),
		TxnCount:        len(txns),
		Last10:          last10,
		Shareholders:    shBalances,
		YearlySummaries: years,
	}

	out, _ := json.Marshal(report)
	return events.APIGatewayProxyResponse{StatusCode: 200, Headers: corsHeaders(), Body: string(out)}, nil
}

func fetchTxns() ([]Transaction, error) {
	resp, err := http.Get(firebaseURL + "/txns.json")
	if err != nil {
		return nil, fmt.Errorf("firebase fetch failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawMap); err != nil || rawMap == nil {
		return []Transaction{}, nil
	}

	txns := make([]Transaction, 0, len(rawMap))
	for key, val := range rawMap {
		var t Transaction
		if err := json.Unmarshal(val, &t); err == nil {
			t.FirebaseKey = key
			txns = append(txns, t)
		}
	}
	return txns, nil
}

func fetchDeposits() (map[string]map[string]float64, error) {
	resp, err := http.Get(firebaseURL + "/deposits.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var rawMap map[string]map[string]float64
	if err := json.Unmarshal(body, &rawMap); err != nil {
		return map[string]map[string]float64{}, nil
	}
	return rawMap, nil
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
