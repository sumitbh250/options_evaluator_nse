package main

const (
	NseExpiryDate = "24-Jun-2021"
	ZerodhaExpiryDate = "21JUN"

	// ResultSortType = SortType_MIN_PROFIT
	ResultSortType = SortType_AVG_PROFIT

	// This is the range which is used for safety calculation
	SafetyRangeMinProfit = 0
	SafetyRangeMaxLoss = 0
	SafetyRangeLowerPct = 0.05
	SafetyRangeUpperPct = 0.3

	// This is the range for profit calculation.
	MinAvgProfitPct = 0
	ProfitRangeMinProfit = 00
	ProfitRangeMaxLoss = 0
	ProfitRangeLowerPct = 0.05
	ProfitRangeUpperPct = 0.1

	// This range is used for analysis
	AnalysisRangeMinProfit = 0
	AnalysisRangeMaxLoss = 2000000
	AnalysisRangeLowerPct = 0.5
	AnalysisRangeUpperPct = 0.5

	// This is used for visual analysis
	PrintAnalysisRangeLowerPct = 0.5
	PrintAnalysisRangeUpperPct = 0.5

	// This range is for chosing the valid trades around current price.
	StrikeRangeLowerPct = 0.3
	StrikeRangeUpperPct = 0.3
	ApplyStrikeRange = false

	// Trade variables.
	MinTotalTradedVolume = 1
	MaxInvestmentAmount = 100000
	NumTrades = 3

	// Http variables.
	HttpTimeoutSecs = 30
	HttpRetryCount = 4
	MaxFileDescriptors = 30
)

var fixedTradesArr []TradeIfc
// aa := new(PEBuyTrade)
// bb := new(PESellTrade)
// aa.Premium = 3.3
// aa.StrikePrice = 170
// bb.Premium = 5.4
// bb.StrikePrice = 180
// fixedTradesArr = append(fixedTradesArr, aa, bb)
