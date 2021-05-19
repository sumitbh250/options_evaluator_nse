package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
  _ "net/http/pprof"
	"net/http"
  "net/url"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
	// "github.com/davecgh/go-spew/spew"
)

type Range struct {
	start float64
	end float64
	step float64
	profitStart float64
	profitEnd float64
	safetyStart float64
	safetyEnd float64
}

type Result struct {
	symbol string
	currentPrice float64
	lotSize float64
	stockRange Range
	tradeCombo []TradeIfc
	profitRatio float64
	minProfit float64
	avgProfit float64
	amountInvested float64
}

func getTradeRange(currentPrice float64, allCalls *AllCallsJson) Range {
	firstLowest, secondLowest := math.MaxFloat64, math.MaxFloat64
	for _, call := range allCalls.CallsArr {
	  var callDetails *SingleCallDetailsJson
		if call.CE != nil {
			callDetails = call.CE
		}
		if call.PE != nil {
			callDetails = call.PE
		}
		if callDetails.ExpiryDate != NseExpiryDate {
			continue
		}
		if callDetails.StrikePrice < firstLowest {
			secondLowest = firstLowest
			firstLowest = callDetails.StrikePrice
		} else if callDetails.StrikePrice < secondLowest {
			secondLowest = callDetails.StrikePrice
		}
	}
	lowerBound := math.Floor(currentPrice - (currentPrice * AnalysisRangeLowerPct))
	upperBound := math.Ceil(currentPrice + (currentPrice * AnalysisRangeUpperPct))
  safetyStart := math.Floor(currentPrice - (currentPrice * SafetyRangeLowerPct))
	safetyEnd := math.Ceil(currentPrice + (currentPrice * SafetyRangeUpperPct))
	profitStart := math.Floor(currentPrice - (currentPrice * ProfitRangeLowerPct))
	profitEnd := math.Floor(currentPrice + (currentPrice * ProfitRangeUpperPct))
	return Range{lowerBound, upperBound, (secondLowest-firstLowest), profitStart, profitEnd, safetyStart, safetyEnd}
}

func findTradeResult(tradeCombo []TradeIfc, stockRange Range, lotSize float64, symbol string, currentPrice float64) {
	var result *Result = nil
	defer func() {
		if result != nil {
			ResultMutex.Lock()
			Results = append(Results, result)
			ResultMutex.Unlock()
		}
	}()

	var amountInvested float64 = 0
	hasSellTrade := false
	for _, trade := range tradeCombo {
		if trade.GetTradeType() == TradeType_SELL {
			hasSellTrade = true
		} else if trade.GetTradeType() == TradeType_BUY {
			amountInvested += (trade.GetPremium() * lotSize)
		}
	}
	if amountInvested > MaxInvestmentAmount {
		return
	}
	var lossCount float64 = 0
	var total float64 = 0
	minProfit := math.MaxFloat64
	var avgProfit float64 =  0
	// profitArr []int
	for expiryPrice := stockRange.start; expiryPrice <= stockRange.end; expiryPrice += stockRange.step {
		var profit float64 =  0
		// total++
		for _, trade := range tradeCombo {
			profit += trade.ProfitAmount(expiryPrice)
		}
		if profit >= 0 {
			// profitArr = append(profitArr, profit)
			if stockRange.profitStart <= expiryPrice && stockRange.profitEnd >= expiryPrice {
				if profit * lotSize < ProfitRangeMinProfit {
					return
				}
				total++
				if profit < minProfit {
					minProfit = profit
				}
				avgProfit += profit
			} else if stockRange.safetyStart <= expiryPrice && stockRange.safetyEnd >= expiryPrice {
				if profit * lotSize < SafetyRangeMinProfit {
					return
				}
				// total++
				// if profit < minProfit {
				// 	minProfit = profit
				// }
				// avgProfit += profit
			} else if profit * lotSize < AnalysisRangeMinProfit {
				return
			}
		} else {
			lossCount++
			if stockRange.profitStart <= expiryPrice && stockRange.profitEnd >= expiryPrice {
				if profit * lotSize < (-1 * ProfitRangeMaxLoss) {
					return
				}
			} else if stockRange.safetyStart <= expiryPrice && stockRange.safetyEnd >= expiryPrice {
				if profit * lotSize < (-1 * SafetyRangeMaxLoss) {
					return
				}
			} else if profit * lotSize < (-1 * AnalysisRangeMaxLoss) {
				return
			}
		}
	}
	if hasSellTrade {
		amountInvested += Web.GetMarginForTrades(tradeCombo, symbol, lotSize, ZerodhaExpiryDate)
	}
	if amountInvested > MaxInvestmentAmount {
		return
	}
	minProfit = (minProfit * lotSize * 100)/amountInvested
	// avgProfit = (((avgProfit * lotSize * 100) - (40 * NumTrades))/(total*amountInvested))
	avgProfit = ((avgProfit * lotSize * 100)/(total*amountInvested))
	profitRatio := ((total - lossCount) * 100)/total
	if avgProfit < MinAvgProfitPct {
		return
	}
	result = &Result{symbol: symbol, minProfit: minProfit, avgProfit: avgProfit,
		profitRatio: profitRatio, stockRange: stockRange, lotSize: lotSize,
		amountInvested: amountInvested, currentPrice: currentPrice, tradeCombo: tradeCombo}
}

func isTradeComboValid(tradeCombo []TradeIfc) bool {
	numBuy, numSell, totalCalls := 0, 0, 0
	ceBuyStrikes := make(map[float64]bool)
	ceSellStrikes := make(map[float64]bool)
	peBuyStrikes := make(map[float64]bool)
	peSellStrikes := make(map[float64]bool)
	for _, trade := range(tradeCombo) {
		if trade.GetTradeType() == TradeType_SELL {
			numSell++
			totalCalls++
			if trade.GetCallType() == CallType_CALL {
				if exists := ceBuyStrikes[trade.GetStrikePrice()]; exists {
					return false
				}
				ceSellStrikes[trade.GetStrikePrice()] = true
			} else if trade.GetCallType() == CallType_PUT {
				if exists := peBuyStrikes[trade.GetStrikePrice()]; exists {
					return false
				}
				peSellStrikes[trade.GetStrikePrice()] = true
			}
		} else if trade.GetTradeType() == TradeType_BUY {
			numBuy++
			totalCalls++
			if trade.GetCallType() == CallType_CALL {
				if exists := ceSellStrikes[trade.GetStrikePrice()]; exists {
					return false
				}
				ceBuyStrikes[trade.GetStrikePrice()] = true
			} else if trade.GetCallType() == CallType_PUT {
				if exists := peSellStrikes[trade.GetStrikePrice()]; exists {
					return false
				}
				peBuyStrikes[trade.GetStrikePrice()] = true
			}
		}
	}

	if totalCalls == 0 {
		return false
	}

	// if numBuy <= numSell {
	// 	return false
	// }
	// if numSell > numBuy {
	// 	return false
	// }
	// if totalCalls < 2 {
	// 	return false
	// } else if totalCalls == 2 && (numSell == 2 || numBuy == 2) {
	// 	return false
	// } else if totalCalls == 3 && (numBuy >= 2 || numSell >= 2) {
	// 	return false
	// } else if totalCalls == 4 && (numSell > 3 || numBuy > 3) {
	// 	return false
	// }
	return true
}

func validTradeComboIter(trades []TradeIfc, numTrades int32, stock string) chan []TradeIfc {
	chnl := make(chan []TradeIfc)
	i, j := 0, 0
	go func() {
		for tradeCombo := range combinationsWithReplacement(trades, numTrades) {
			// spew.Dump(tradeCombo)
			if isTradeComboValid(tradeCombo) {
				chnl <- tradeCombo
				i++
			}
			j++
		}
		fmt.Println(stock, len(trades), i, j)
		close(chnl)
	}()
	return chnl
}

func checkValidity(callDetails *SingleCallDetailsJson, currentPrice float64) bool {
	if callDetails.ExpiryDate != NseExpiryDate {
		return false
	}
	if callDetails.TotalTradedVolume < MinTotalTradedVolume {
		return false
	}
	if ApplyStrikeRange {
		lowerBound := math.Floor(currentPrice - (currentPrice * StrikeRangeLowerPct))
		upperBound := math.Ceil(currentPrice + (currentPrice * StrikeRangeUpperPct))
		if callDetails.StrikePrice < lowerBound || callDetails.StrikePrice > upperBound {
			return false
		}
	}
	return true
}

func addValidCallsToTradeArr(tradesArr []TradeIfc, allCalls *AllCallsJson, currentPrice float64) []TradeIfc {
	for _, call := range allCalls.CallsArr {
		if call.CE != nil && checkValidity(call.CE, currentPrice) {
			buyTrade := NewCEBuyTrade(call.CE)
			sellTrade := NewCESellTrade(call.CE)
			if buyTrade.Premium > 0 {
				tradesArr = append(tradesArr, buyTrade)
			}
			if sellTrade.Premium > 0 {
				tradesArr = append(tradesArr, sellTrade)
			}
		}
		if call.PE != nil && checkValidity(call.PE, currentPrice) {
			buyTrade := NewPEBuyTrade(call.PE)
			sellTrade := NewPESellTrade(call.PE)
			if buyTrade.Premium > 0 {
				tradesArr = append(tradesArr, buyTrade)
			}
			if sellTrade.Premium > 0 {
				tradesArr = append(tradesArr, sellTrade)
			}
		}
	}
	return tradesArr
}

func getStockResult(stock string, lotSize float64, url string) {
	currentPrice, _ := Web.FetchCurrentPrice(stock)
	jsonCalls, _ := Web.FetchOptionsData(url)
	rangeToCheck := getTradeRange(currentPrice, jsonCalls)
	var tradesArr []TradeIfc
	tradesArr = append(tradesArr, NewNullTrade())
	// aa := new(PEBuyTrade)
	// bb := new(PESellTrade)
	// aa.Premium = 3.3
	// aa.StrikePrice = 170
	// bb.Premium = 5.4
	// bb.StrikePrice = 180
	// tradesArr = append(tradesArr, aa, bb)
	tradesArr = addValidCallsToTradeArr(tradesArr, jsonCalls, currentPrice)
	for tradeCombo := range validTradeComboIter(tradesArr, NumTrades, stock) {
		tradeComboCopy := make([]TradeIfc, len(tradeCombo))
		_ = copy(tradeComboCopy, tradeCombo)
		// spew.Dump(tradeCombo)
		WG.Add(1)
		go func () {
			defer WG.Done()
			findTradeResult(tradeComboCopy, rangeToCheck, lotSize, stock, currentPrice)
		}()
	}
}

var WG sync.WaitGroup
var ResultMutex sync.Mutex
var Results []*Result
var Web *WebSession

func printResultAnalysis(r *Result) {
	fmt.Println("Symbol", r.symbol, "Current Price", r.currentPrice,
	  "Lot Size", r.lotSize, "Range", r.stockRange)
	fmt.Println("Profit Ratio", r.profitRatio, "Avg Profit", r.avgProfit,
	  "Min Profit", r.minProfit)
	fmt.Println("Amount Invested", r.amountInvested)
	for _, trade := range(r.tradeCombo) {
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
			fmt.Println("\033[92m", expiryPrice, "    ",  int(profit * r.lotSize), "\033[00m")
		} else {
			fmt.Println("\033[91m", expiryPrice, "    ",  int(profit * r.lotSize), "\033[00m")
		}
	}
}

type SortType int32
const (
	SortType_MIN_PROFIT SortType = iota
	SortType_AVG_PROFIT
)

func analyzeResults() {
	fmt.Println("Results printing")
	// Reverse sort the results
	sort.Slice(Results, func(i, j int) bool {
		if ResultSortType == SortType_AVG_PROFIT {
			return Results[i].avgProfit > Results[j].avgProfit
		} else if  ResultSortType == SortType_MIN_PROFIT {
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

func init() {
	Web = NewWebSession(HttpTimeoutSecs, HttpRetryCount, MaxFileDescriptors)
}

func main() {

	fmt.Println("acsdsd")
	indicesTemplate := "https://www.nseindia.com/api/option-chain-indices?symbol="
	WG.Add(len(OptionsIndiceLotDict) + len(OptionsStockLotDict))
	for stock, lotSize := range OptionsIndiceLotDict {
		url := indicesTemplate + url.QueryEscape(stock)
		go func(stock string, lotSize float64) {
			defer WG.Done()
			getStockResult(stock, lotSize, url)
		}(stock, lotSize)
		// fmt.Println("\033[91m", stock, lotSize, url, "\033[00m")
	}
	stocksTemplate := "https://www.nseindia.com/api/option-chain-equities?symbol="
	// count:=0
	for stock, lotSize := range OptionsStockLotDict {
	  url := stocksTemplate + url.QueryEscape(stock)
		go func(stock string, lotSize float64) {
			defer WG.Done()
			getStockResult(stock, lotSize, url)
		}(stock, lotSize)
		// count++
		// if (count % 10 == 0) {
		// 	time.Sleep(0 * time.Second)
		// }
		// fmt.Println("\033[92m", stock, lotSize, url, "\033[00m")
	}

	go func() {
		for {
			fmt.Println(runtime.NumGoroutine())
			time.Sleep(30 * time.Second)
		}
	}()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	WG.Wait()
	analyzeResults()
}

