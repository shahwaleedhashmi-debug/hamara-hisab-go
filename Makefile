build:
	GOOS=linux GOARCH=amd64 go build -o netlify/functions/transactions/transactions ./netlify/functions/transactions/
	GOOS=linux GOARCH=amd64 go build -o netlify/functions/deposit/deposit ./netlify/functions/deposit/
	GOOS=linux GOARCH=amd64 go build -o netlify/functions/reports/reports ./netlify/functions/reports/
	GOOS=linux GOARCH=amd64 go build -o netlify/functions/data/data ./netlify/functions/data/

.PHONY: build
