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
	NseExpiryDate     = "26-Oct-2023"
	ZerodhaExpiryDate = "23OCT"

	// ResultSortType = SortType_MIN_PROFIT
	ResultSortType = SortType_AVG_PROFIT

	// This is the range which is used for safety calculation
	SafetyRangeMinProfit = 0
	SafetyRangeMaxLoss   = 000
	SafetyRangeLowerPct  = 0.05
	SafetyRangeUpperPct  = 0.02

	// This is the range for profit calculation.
	MinAvgProfitPct      = 0
	ProfitRangeMinProfit = 00
	ProfitRangeMaxLoss   = 000
	ProfitRangeLowerPct  = 0.05
	ProfitRangeUpperPct  = 0.02

	// This range is used for analysis
	AnalysisRangeMinProfit = 0
	AnalysisRangeMaxLoss   = 3000000000000
	AnalysisRangeLowerPct  = 0.07
	AnalysisRangeUpperPct  = 0.07

	// This is used for visual analysis
	PrintAnalysisRangeLowerPct = 0.07
	PrintAnalysisRangeUpperPct = 0.07

	// This range is for chosing the valid trades around current price.
	StrikeRangeLowerPct = 0.1
	StrikeRangeUpperPct = 0.1
	ApplyStrikeRange    = false

	// Trade variables.
	MinTotalTradedVolume = 100
	MaxInvestmentAmount  = 160000
	NumTrades            = 2

	// Http variables.
	HttpTimeoutSecs    = 30
	HttpRetryCount     = 4
	MaxFileDescriptors = 10

	ResultAnalysisType = AnalysisType_TERMINAL
)

var fixedTradesArr []TradeIfc
