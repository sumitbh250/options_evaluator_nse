package main

import (
	"fmt"
	"math"
  _ "net/http/pprof"

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

func getStockResult(stock string, lotSize float64) {
	currentPrice, _ := Web.FetchCurrentPrice(stock)
	jsonCalls, _ := Web.FetchOptionsData(stock)
	rangeToCheck := getTradeRange(currentPrice, jsonCalls)
	var tradesArr []TradeIfc
	tradesArr = append(tradesArr, NewNullTrade())
	tradesArr = addValidOptionCallsToTradeArr(tradesArr, jsonCalls, currentPrice)
	jsonOptFutCalls, _ := Web.FetchOptionsAndFuturesData(stock)
	// spew.Dump(jsonOptFutCalls)
	tradesArr = addValidFutCallsToTradeArr(tradesArr, jsonOptFutCalls)
	for tradeCombo := range validTradeComboIter(fixedTradesArr, tradesArr, NumTrades, stock) {
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

func getTradeRange(currentPrice float64, allCalls *AllCallsJson) Range {
	lowerClosest, higherClosest := float64(-1), math.MaxFloat64
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
		if callDetails.StrikePrice < currentPrice && lowerClosest < callDetails.StrikePrice {
			lowerClosest = callDetails.StrikePrice
		} else if callDetails.StrikePrice > currentPrice && higherClosest > callDetails.StrikePrice {
			higherClosest = callDetails.StrikePrice
		}
	}
	lowerBound := math.Floor(currentPrice - (currentPrice * AnalysisRangeLowerPct))
	upperBound := math.Ceil(currentPrice + (currentPrice * AnalysisRangeUpperPct))
  safetyStart := math.Floor(currentPrice - (currentPrice * SafetyRangeLowerPct))
	safetyEnd := math.Ceil(currentPrice + (currentPrice * SafetyRangeUpperPct))
	profitStart := math.Floor(currentPrice - (currentPrice * ProfitRangeLowerPct))
	profitEnd := math.Floor(currentPrice + (currentPrice * ProfitRangeUpperPct))
	return Range{lowerBound, upperBound, (higherClosest-lowerClosest), profitStart, profitEnd, safetyStart, safetyEnd}
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
	hasFutureTrade := false
	for _, trade := range tradeCombo {
		if trade.GetTradeType() == TradeType_SELL {
			hasSellTrade = true
		} else if (trade.GetTradeType() == TradeType_BUY && trade.GetCallType() != CallType_FUTURE) {
			amountInvested += (trade.GetPremium() * lotSize)
		}
		if trade.GetCallType() == CallType_FUTURE {
			hasFutureTrade = true
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
	if hasSellTrade || hasFutureTrade {
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

func addValidOptionCallsToTradeArr(tradesArr []TradeIfc, allCalls *AllCallsJson, currentPrice float64) []TradeIfc {
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

func addValidFutCallsToTradeArr(tradesArr []TradeIfc, optFutCalls *OptionsFutRecords) []TradeIfc {
	for _, call := range optFutCalls.Stocks {
		if call.Metadata.InstrumentType != "Stock Futures" && call.Metadata.InstrumentType != "Index Futures" {
			continue
		}
		if call.Metadata.ExpiryDate != NseExpiryDate {
			continue
		}
		buyTrade := NewFutureBuyTrade(call)
		sellTrade := NewFutureSellTrade(call)
		if buyTrade.Premium > 0 {
			tradesArr = append(tradesArr, buyTrade)
		}
		if buyTrade.Premium > 0 {
			tradesArr = append(tradesArr, sellTrade)
		}
	}
	return tradesArr
}

func validTradeComboIter(fixedTrades []TradeIfc, trades []TradeIfc, numTrades int32, stock string) chan []TradeIfc {
	chnl := make(chan []TradeIfc)
	i, j := 0, 0
	go func() {
		numTrades -= int32(len(fixedTrades))
		for tradeCombo := range combinationsWithReplacement(trades, numTrades) {
			// spew.Dump(tradeCombo)
			tradeCombo = append(tradeCombo, fixedTrades...)
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

func isTradeComboValid(tradeCombo []TradeIfc) bool {
	numBuy, numSell, numFut, totalCalls := 0, 0, 0, 0
	ceBuyStrikes := make(map[float64]bool)
	ceSellStrikes := make(map[float64]bool)
	peBuyStrikes := make(map[float64]bool)
	peSellStrikes := make(map[float64]bool)
	futBuyTrade := false
	futSellTrade := false
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
			} else if trade.GetCallType() == CallType_FUTURE {
				if futBuyTrade {
					return false
				}
				numFut++
				futSellTrade = true
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
			} else if trade.GetCallType() == CallType_FUTURE {
				if futSellTrade {
					return false
				}
				numFut++
				futBuyTrade = true
			}
		}
	}

	if totalCalls == 0 {
		return false
	}

	// if numFut == 0 {
	// 	return false
	// }
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
