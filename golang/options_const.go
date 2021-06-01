package main

type SortType int32
const (
	SortType_MIN_PROFIT SortType = iota
	SortType_AVG_PROFIT
)

type AnalysisType int32
const (
	AnalysisType_TERMINAL AnalysisType = iota
	AnalysisType_TELEGRAM
)

const (
	NseExpiryDate = "24-Jun-2021"
	ZerodhaExpiryDate = "21JUN"

	// ResultSortType = SortType_MIN_PROFIT
	ResultSortType = SortType_AVG_PROFIT

	// This is the range which is used for safety calculation
	SafetyRangeMinProfit = 0
	SafetyRangeMaxLoss = 0
	SafetyRangeLowerPct = 0.04
	SafetyRangeUpperPct = 0.04

	// This is the range for profit calculation.
	MinAvgProfitPct = 0
	ProfitRangeMinProfit = 00
	ProfitRangeMaxLoss = 0
	ProfitRangeLowerPct = 0.03
	ProfitRangeUpperPct = 0.03

	// This range is used for analysis
	AnalysisRangeMinProfit = 0
	AnalysisRangeMaxLoss = 0000000000000
	AnalysisRangeLowerPct = 0.5
	AnalysisRangeUpperPct = 0.5

	// This is used for visual analysis
	PrintAnalysisRangeLowerPct = 0.22
	PrintAnalysisRangeUpperPct = 0.22

	// This range is for chosing the valid trades around current price.
	StrikeRangeLowerPct = 0.3
	StrikeRangeUpperPct = 0.3
	ApplyStrikeRange = false

	// Trade variables.
	MinTotalTradedVolume = 1000
	MaxInvestmentAmount = 100000
	NumTrades = 1

	// Http variables.
	HttpTimeoutSecs = 30
	HttpRetryCount = 4
	MaxFileDescriptors = 30

	ResultAnalysisType = AnalysisType_TELEGRAM
)

var fixedTradesArr []TradeIfc
// aa := new(PEBuyTrade)
// bb := new(PESellTrade)
// aa.Premium = 3.3
// aa.StrikePrice = 170
// bb.Premium = 5.4
// bb.StrikePrice = 180
// fixedTradesArr = append(fixedTradesArr, aa, bb)
