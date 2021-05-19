package main

const (
	NseExpiryDate = "27-May-2021"
	ZerodhaExpiryDate = "21MAY"

	ResultSortType = SortType_MIN_PROFIT
	// ResultSortType = SortType_AVG_PROFIT

	// This is the range which is used for safety calculation
	SafetyRangeMinProfit = 0
	SafetyRangeMaxLoss = 0
	SafetyRangeLowerPct = 0.035
	SafetyRangeUpperPct = 0.035

	// This is the range for profit calculation.
	MinAvgProfitPct = 3
	ProfitRangeMinProfit = 00
	ProfitRangeMaxLoss = 0
	ProfitRangeLowerPct = 0.03
	ProfitRangeUpperPct = 0.03

	// This range is used for analysis
	AnalysisRangeMinProfit = 0
	AnalysisRangeMaxLoss = 1000000000000
	AnalysisRangeLowerPct = 0.15
	AnalysisRangeUpperPct = 0.15

	// This is used for visual analysis
	PrintAnalysisRangeLowerPct = 0.15
	PrintAnalysisRangeUpperPct = 0.15

	// // This is the range which is used for safety calculation
	// SafetyRangeMinProfit = 0
	// SafetyRangeMaxLoss = 0
	// SafetyRangeLowerPct = 0.2
	// SafetyRangeUpperPct = 0.2

	// // This is the range for profit calculation.
	// MinAvgProfitPct = 0
	// ProfitRangeMinProfit = 00
	// ProfitRangeMaxLoss = 50000
	// ProfitRangeLowerPct = 0.03
	// ProfitRangeUpperPct = 0.03

	// // This range is used for analysis
	// AnalysisRangeMinProfit = 0
	// AnalysisRangeMaxLoss = 25000
	// AnalysisRangeLowerPct = 0.4
	// AnalysisRangeUpperPct = 0.4

	// // This is used for visual analysis
	// PrintAnalysisRangeLowerPct = 0.2
	// PrintAnalysisRangeUpperPct = 0.2

	// This range is for chosing the valid trades around current price.
	StrikeRangeLowerPct = 0.3
	StrikeRangeUpperPct = 0.3
	ApplyStrikeRange = false

	// Trade variables.
	MinTotalTradedVolume = 1000
	MaxInvestmentAmount = 50000
	NumTrades = 3

	// Http variables.
	HttpTimeoutSecs = 30
	HttpRetryCount = 4
	MaxFileDescriptors = 30
)
