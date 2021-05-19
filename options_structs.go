package main

import "fmt"

type TradeType int32
const (
	TradeType_NULL TradeType = iota
	TradeType_BUY
	TradeType_SELL
)

type CallType int32
const (
	CallType_NULL CallType = iota
	CallType_CALL
	CallType_PUT
)

type Trade struct {
	StrikePrice float64
	Premium float64
	OpenInterest float64
	TradedVolume float64
  TradeType TradeType
	CallType CallType
}
type TradeIfc interface {
	ProfitAmount(expiryPrice float64) float64
	GetTradeType() TradeType
	GetCallType() CallType
	GetPremium() float64
	GetStrikePrice() float64
	PrintDetails()
}

type NullTrade struct {
	Trade
}
func NewNullTrade() *NullTrade {
	trade := new(NullTrade)
	trade.Premium = 0
	trade.OpenInterest = 0
	trade.TradedVolume = 0
	trade.StrikePrice = 0
	trade.TradeType = TradeType_NULL
	trade.CallType = CallType_NULL
	return trade
}
func (t *NullTrade) ProfitAmount(expiryPrice float64) float64 {
	return 0
}
func (t *NullTrade) GetTradeType() TradeType {
	return t.TradeType
}
func (t *NullTrade) GetCallType() CallType {
	return t.CallType
}
func (t *NullTrade) GetPremium() float64 {
	return t.Premium
}
func (t *NullTrade) GetStrikePrice() float64 {
	return t.StrikePrice
}
func (t *NullTrade) PrintDetails() {
	fmt.Println("Null Trade")
}

type PEBuyTrade struct {
	Trade
}
func NewPEBuyTrade(call *SingleCallDetailsJson) *PEBuyTrade {
	trade := new(PEBuyTrade)
	trade.Premium = call.AskPrice
	trade.OpenInterest = call.OpenInterest
	trade.TradedVolume = call.TotalTradedVolume
	trade.StrikePrice = call.StrikePrice
	trade.TradeType = TradeType_BUY
	trade.CallType = CallType_PUT
	return trade
}
func (t *PEBuyTrade) ProfitAmount(expiryPrice float64) float64 {
	if (expiryPrice < t.StrikePrice) {
		return (t.StrikePrice - expiryPrice - t.Premium)
	} else {
		return (-1 * t.Premium)
	}
}
func (t *PEBuyTrade) GetTradeType() TradeType {
	return t.TradeType
}
func (t *PEBuyTrade) GetCallType() CallType {
	return t.CallType
}
func (t *PEBuyTrade) GetPremium() float64 {
	return t.Premium
}
func (t *PEBuyTrade) GetStrikePrice() float64 {
	return t.StrikePrice
}
func (t *PEBuyTrade) PrintDetails() {
	fmt.Println("PE Buy", t.StrikePrice, t.Premium, t.TradedVolume, t.OpenInterest)
}

type PESellTrade struct {
	Trade
}
func NewPESellTrade(call *SingleCallDetailsJson) *PESellTrade {
	trade := new(PESellTrade)
	trade.Premium = call.BidPrice
	trade.OpenInterest = call.OpenInterest
	trade.TradedVolume = call.TotalTradedVolume
	trade.StrikePrice = call.StrikePrice
	trade.TradeType = TradeType_SELL
	trade.CallType = CallType_PUT
	return trade
}
func (t *PESellTrade) ProfitAmount(expiryPrice float64) float64 {
	if (expiryPrice < t.StrikePrice) {
		return (-1 * (t.StrikePrice - expiryPrice - t.Premium))
	} else {
		return t.Premium
	}
}
func (t *PESellTrade) GetTradeType() TradeType {
	return t.TradeType
}
func (t *PESellTrade) GetCallType() CallType {
	return t.CallType
}
func (t *PESellTrade) GetPremium() float64 {
	return t.Premium
}
func (t *PESellTrade) GetStrikePrice() float64 {
	return t.StrikePrice
}
func (t *PESellTrade) PrintDetails() {
	fmt.Println("PE Sell", t.StrikePrice, t.Premium, t.TradedVolume, t.OpenInterest)
}

type CEBuyTrade struct {
	Trade
}
func NewCEBuyTrade(call *SingleCallDetailsJson) *CEBuyTrade {
	trade := new(CEBuyTrade)
	trade.Premium = call.AskPrice
	trade.OpenInterest = call.OpenInterest
	trade.TradedVolume = call.TotalTradedVolume
	trade.StrikePrice = call.StrikePrice
	trade.TradeType = TradeType_BUY
	trade.CallType = CallType_CALL
	return trade
}
func (t *CEBuyTrade) ProfitAmount(expiryPrice float64) float64 {
	if (expiryPrice < t.StrikePrice) {
		return (-1 * t.Premium)
	} else {
		return (expiryPrice - t.StrikePrice - t.Premium)
	}
}
func (t *CEBuyTrade) GetTradeType() TradeType {
	return t.TradeType
}
func (t *CEBuyTrade) GetCallType() CallType {
	return t.CallType
}
func (t *CEBuyTrade) GetPremium() float64 {
	return t.Premium
}
func (t *CEBuyTrade) GetStrikePrice() float64 {
	return t.StrikePrice
}
func (t *CEBuyTrade) PrintDetails() {
	fmt.Println("CE Buy", t.StrikePrice, t.Premium, t.TradedVolume, t.OpenInterest)
}

type CESellTrade struct {
	Trade
}
func NewCESellTrade(call *SingleCallDetailsJson) *CESellTrade {
	trade := new(CESellTrade)
	trade.Premium = call.BidPrice
	trade.OpenInterest = call.OpenInterest
	trade.TradedVolume = call.TotalTradedVolume
	trade.StrikePrice = call.StrikePrice
	trade.TradeType = TradeType_SELL
	trade.CallType = CallType_CALL
	return trade
}
func (t *CESellTrade) ProfitAmount(expiryPrice float64) float64 {
	if (expiryPrice < t.StrikePrice) {
		return t.Premium
	} else {
		return (-1 * (expiryPrice - t.StrikePrice - t.Premium))
	}
}
func (t *CESellTrade) GetTradeType() TradeType {
	return t.TradeType
}
func (t *CESellTrade) GetCallType() CallType {
	return t.CallType
}
func (t *CESellTrade) GetPremium() float64 {
	return t.Premium
}
func (t *CESellTrade) GetStrikePrice() float64 {
	return t.StrikePrice
}
func (t *CESellTrade) PrintDetails() {
	fmt.Println("CE Sell", t.StrikePrice, t.Premium, t.TradedVolume, t.OpenInterest)
}
