package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func AnalyzeResults() {
	fmt.Println("Results printing")
	// Reverse sort the results
	sort.Slice(Results, func(i, j int) bool {
		if ResultSortType == SortType_AVG_PROFIT {
			return Results[i].avgProfit > Results[j].avgProfit
		} else if ResultSortType == SortType_MIN_PROFIT {
			return Results[i].minProfit > Results[j].minProfit
		} else {
			return false
		}
	})
	idx := -1
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Enter index out of ", len(Results))
		byteText, _, _ := reader.ReadLine()
		text := string(byteText)
		fmt.Println("Input is: ", text)
		if text == "" {
			idx++
		} else if i, err := strconv.Atoi(text); err != nil {
			log.Fatalf("Invalid index")
		} else {
			idx = i
		}
		fmt.Println("Printing trade at index", idx)
		printResultAnalysis(Results[idx])
	}
}

func printResultAnalysis(r *Result) {
	fmt.Println("Symbol", r.symbol, "Current Price", r.currentPrice,
		"Lot Size", r.lotSize, "Range", r.stockRange)
	fmt.Println("Profit Ratio", r.profitRatio, "Avg Profit", r.avgProfit,
		"Min Profit", r.minProfit)
	fmt.Println("Amount Invested", r.amountInvested)
	for _, trade := range r.tradeCombo {
		trade.PrintDetails()
	}
	step := r.stockRange.step
	lowerBound := math.Floor(r.currentPrice - (r.currentPrice * PrintAnalysisRangeLowerPct))
	upperBound := math.Ceil(r.currentPrice + (r.currentPrice * PrintAnalysisRangeUpperPct))
	for expiryPrice := lowerBound; expiryPrice <= upperBound; expiryPrice += step {
		profit := float64(0)
		for _, trade := range r.tradeCombo {
			profit += trade.ProfitAmount(expiryPrice)
		}
		if profit > 0 {
			fmt.Println("\033[92m", expiryPrice, "    ", int(profit*r.lotSize), "\033[00m")
		} else {
			fmt.Println("\033[91m", expiryPrice, "    ", int(profit*r.lotSize), "\033[00m")
		}
	}
}

func SendResultAnalysisViaTelegram() {
	for _, chatId := range ChatIds {
		msg := tgbotapi.NewMessage(chatId, "hello")
		Bot.Send(msg)
	}
}
