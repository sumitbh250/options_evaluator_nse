package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WebSession struct {
	timeoutSecs int32
	retryCount  int32
	client      *http.Client
	maxFDChan   chan bool
}

func NewWebSession(timeoutSecs int32, retryCount int32, maxFileDescriptors int32) *WebSession {
	n := new(WebSession)
	n.timeoutSecs = timeoutSecs
	n.retryCount = retryCount
	n.maxFDChan = make(chan bool, maxFileDescriptors)
	jar, _ := cookiejar.New(nil)
	n.client = &http.Client{
		Timeout: time.Duration(timeoutSecs) * time.Second,
		Jar:     jar,
	}
	n.initNseSession()
	n.initZerodhaSession()
	return n
}

func getHttpGetRequest(strUrl string, header http.Header) *http.Request {
	request, err := http.NewRequest("GET", strUrl, nil)
	if err != nil {
		log.Fatal("Error Parsing URL", strUrl, err)
	}
	request.Header = header
	request.Close = true
	return request
}

func getHttpPostRequest(strUrl string, header http.Header, data url.Values) *http.Request {
	request, err := http.NewRequest("POST", strUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal("Error Parsing URL", strUrl, err)
	}
	request.Header = header
	request.Close = true
	// spew.Dump(request)
	return request
}

func (ws *WebSession) getDataFromWeb(request *http.Request) ([]byte, error) {
	var response *http.Response
	var err error
	defer func() {
		if response != nil {
			response.Body.Close()
		}
	}()
	retryCount := ws.retryCount
	for retryCount <= ws.retryCount {
		ws.maxFDChan <- true
		response, err = ws.client.Do(request)
		<-ws.maxFDChan
		if err == nil {
			break
		}
		if response != nil {
			response.Body.Close()
		}
		retryCount++
		time.Sleep(time.Duration(retryCount) * time.Second)
	}
	if err != nil {
		fmt.Println("Error fetching data for:", request.URL, err, response, retryCount)
		return nil, err
	}

	if response.StatusCode != 200 {
		fmt.Println("Error fetching data for:", request.URL, err, response, retryCount)
		return nil, errors.New("Response code: " + strconv.Itoa(response.StatusCode))
	}
	// if request.Method == "POST" {
	// 	spew.Dump(response)
	// }
	body, err := ioutil.ReadAll(response.Body)
	// if request.Method == "POST" {
	// 	fmt.Printf("%s\n", body)
	// }
	if err != nil {
		log.Fatal("Error reading data for:", request.URL, err, response.Body)
		return nil, err
	}
	return body, err
}

func (ws *WebSession) initNseSession() {
	// header := http.Header{
	// 	"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"},
	// 	"accept-language": {"en,gu;q=0.9,hi;q=0.8"},
	// 	"accept-encoding": {"gzip, deflate, br"},
	// 	"accept":          {"application/json, text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	// }

	header := http.Header{
		"authority":        {"www.nseindia.com"},
		"accept":           {"application/json, text/javascript, */*; q=0.01"},
		"accept-language":  {"en-US,en;q=0.9,hi;q=0.8"},
		"dnt":              {"1"},
		"referer":          {"https://www.nseindia.com/"},
		"sec-fetch-mode":   {"cors"},
		"sec-fetch-site":   {"same-origin"},
		"sec-gpc":          {"1"},
		"user-agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"},
		"x-requested-with": {"XMLHttpRequest"},
	}

	_, err := ws.getDataFromWeb(getHttpGetRequest("https://www.nseindia.com", header))
	if err != nil {
		log.Fatal("Couldn't connect to NSE.", err)
	}
}

type CurrentPrice struct {
	CurrentPrice float64 `json:"underlyingValue,omitempty"`
}

func (ws *WebSession) FetchCurrentPrice(stock string) (float64, error) {
	header := http.Header{
		"authority":        {"www.nseindia.com"},
		"accept":           {"application/json, text/javascript, */*; q=0.01"},
		"accept-language":  {"en-US,en;q=0.9,hi;q=0.8"},
		"dnt":              {"1"},
		"referer":          {"https://www.nseindia.com/"},
		"sec-fetch-mode":   {"cors"},
		"sec-fetch-site":   {"same-origin"},
		"sec-gpc":          {"1"},
		"user-agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"},
		"x-requested-with": {"XMLHttpRequest"},
	}
	url := "https://www.nseindia.com/api/quote-derivative?symbol=" + url.QueryEscape(stock)
	body, err := ws.getDataFromWeb(getHttpGetRequest(url, header))
	var currentPrice CurrentPrice
	err = json.Unmarshal(body, &currentPrice)
	if err != nil {
		fmt.Printf("%s\n", body)
		log.Fatal("Error decoding price response:", stock, body)
		return 0, err
	}
	// fmt.Println(currentPrice)
	return currentPrice.CurrentPrice, nil
}

type SingleCallDetailsJson struct {
	StrikePrice       float64 `json:"strikePrice"`
	ExpiryDate        string  `json:"expiryDate"`
	OpenInterest      float64 `json:"openInterest"`
	TotalTradedVolume float64 `json:"totalTradedVolume"`
	AskPrice          float64 `json:"askPrice"`
	BidPrice          float64 `json:"bidprice"`
	LastPrice         float64 `json:"lastPrice"`
}

type SingleCallJson struct {
	CE *SingleCallDetailsJson `json:"CE"`
	PE *SingleCallDetailsJson `json:"PE"`
}

type AllCallsJson struct {
	CallsArr []*SingleCallJson `json:"data"`
}

type OptionsRecords struct {
	Data AllCallsJson `json:"records"`
}

func (ws *WebSession) FetchOptionsData(stock string) (*AllCallsJson, error) {
	header := http.Header{
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"},
		"accept-language": {"en,gu;q=0.9,hi;q=0.8"},
		"accept-encoding": {"gzip, deflate, br"},
		"accept":          {"application/json, text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	}
	stockUrl := ""
	indicesTemplate := "https://www.nseindia.com/api/option-chain-indices?symbol="
	stocksTemplate := "https://www.nseindia.com/api/option-chain-equities?symbol="
	if _, ok := OptionsIndiceLotDict[stock]; ok {
		stockUrl = indicesTemplate + url.QueryEscape(stock)
	} else {
		stockUrl = stocksTemplate + url.QueryEscape(stock)
	}
	body, err := ws.getDataFromWeb(getHttpGetRequest(stockUrl, header))
	var optionsRecords OptionsRecords
	err = json.Unmarshal(body, &optionsRecords)
	if err != nil {
		log.Fatal("%s\n", body)
		//log.Fatal("Error decoding options response:", url, body)
		return nil, err
	}
	// spew.Dump(optionsRecords)
	return &optionsRecords.Data, nil
}

type MetadataInfo struct {
	InstrumentType          string  `json:instrumentType`
	ExpiryDate              string  `json:expiryDate`
	OptionType              string  `json:optionType`
	StrikePrice             float64 `json:strikePrice`
	Identifier              string  `json:identifier`
	NumberOfContractsTraded float64 `json:numberOfContractsTraded`
}

type OrderInfo struct {
	Price float64 `json:price`
}

type CarryOfCostPriceInfo struct {
	BestBuy   float64 `json:bestBuy`
	BestSell  float64 `json:bestSell`
	LastPrice float64 `json:lastPrice`
}

type CarryOfCostInfo struct {
	Price *CarryOfCostPriceInfo `json:price`
	Carry *CarryOfCostPriceInfo `json:carry`
}

type MarketDeptOrderBookInfo struct {
	Bid         []*OrderInfo     `json:bid`
	Ask         []*OrderInfo     `json:ask`
	CarryOfCost *CarryOfCostInfo `json:carryOfCost`
}

type StocksData struct {
	Metadata            MetadataInfo            `json:metadata`
	MarketDeptOrderBook MarketDeptOrderBookInfo `json:marketDeptOrderBook`
}

type OptionsFutRecords struct {
	Stocks []*StocksData `json:stocks`
}

func (ws *WebSession) FetchOptionsAndFuturesData(stock string) (*OptionsFutRecords, error) {
	header := http.Header{
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"},
		"accept-language": {"en,gu;q=0.9,hi;q=0.8"},
		"accept-encoding": {"gzip, deflate, br"},
		"accept":          {"application/json, text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	}
	urlTemplate := "https://www.nseindia.com/api/quote-derivative?symbol="
	stockUrl := urlTemplate + url.QueryEscape(stock)
	body, err := ws.getDataFromWeb(getHttpGetRequest(stockUrl, header))
	var optionsFutRecords OptionsFutRecords
	err = json.Unmarshal(body, &optionsFutRecords)
	if err != nil {
		log.Fatal("%s\n", body)
		//log.Fatal("Error decoding options response:", url, body)
		return nil, err
	}
	// spew.Dump(optionsRecords)
	return &optionsFutRecords, nil
}

func (ws *WebSession) initZerodhaSession() {
	header := http.Header{
		"User-Agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:84.0) Gecko/20100101 Firefox/84.0"},
		"Accept":           {"application/json, text/javascript, */*; q=0.01"},
		"Accept-Language":  {"en-US,en;q=0.5"},
		"Referer":          {"https://zerodha.com/"},
		"Content-Type":     {"application/x-www-form-urlencoded; charset=UTF-8"},
		"X-Requested-With": {"XMLHttpRequest"},
		"Origin":           {"https://zerodha.com"},
		"Connection":       {"keep-alive"},
		"TE":               {"Trailers"},
		"DNT":              {"1"},
	}
	_, err := ws.getDataFromWeb(getHttpGetRequest("https://zerodha.com/", header))
	if err != nil {
		log.Fatal("Couldn't connect to Zerodha.", err)
	}
}

type TotalRet struct {
	Total float64 `json:"total"`
}

type OptionsMarginRet struct {
	Total TotalRet `json:"total"`
}

func (ws *WebSession) GetMarginForTrades(trades []TradeIfc, symbol string,
	lotSize float64, expiryDate string) float64 {

	header := http.Header{
		"User-Agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:84.0) Gecko/20100101 Firefox/84.0"},
		"Accept":           {"application/json"},
		"Accept-Language":  {"en-US,en;q=0.5"},
		"Referer":          {"https://zerodha.com/"},
		"Content-Type":     {"application/x-www-form-urlencoded; charset=UTF-8"},
		"X-Requested-With": {"XMLHttpRequest"},
		"Origin":           {"https://zerodha.com"},
		"Connection":       {"keep-alive"},
		"TE":               {"Trailers"},
	}
	strUrl := "https://zerodha.com/margin-calculator/SPAN"
	data := url.Values{}
	data.Set("action", "calculate")
	tradesFreq := make(map[TradeIfc]int)
	for _, trade := range trades {
		if trade.GetTradeType() == TradeType_NULL || trade.GetCallType() == CallType_NULL {
			continue
		} else if val, ok := tradesFreq[trade]; ok {
			tradesFreq[trade] = val + 1
		} else {
			tradesFreq[trade] = 1
		}
	}
	for trade, freq := range tradesFreq {
		data.Add("exchange[]", "NFO")
		data.Add("scrip[]", symbol+expiryDate)
		if trade.GetCallType() == CallType_CALL {
			data.Add("product[]", "OPT")
			data.Add("option_type[]", "CE")
		} else if trade.GetCallType() == CallType_PUT {
			data.Add("product[]", "OPT")
			data.Add("option_type[]", "PE")
		} else if trade.GetCallType() == CallType_FUTURE {
			data.Add("product[]", "FUT")
			data.Add("option_type[]", "CE")
		}
		if trade.GetTradeType() == TradeType_BUY {
			data.Add("trade[]", "buy")
		} else if trade.GetTradeType() == TradeType_SELL {
			data.Add("trade[]", "sell")
		}
		data.Add("qty[]", strconv.Itoa(int(lotSize)*freq))
		strStrike := strconv.FormatFloat(trade.GetStrikePrice(), 'f', -1, 64)
		if float64(int(trade.GetStrikePrice())) == trade.GetStrikePrice() {
			strStrike = strconv.Itoa(int(trade.GetStrikePrice()))
		}
		data.Add("strike_price[]", strStrike)
	}
	body, err := ws.getDataFromWeb(getHttpPostRequest(strUrl, header, data))
	if err != nil {
		// fmt.Println("Unable to fetch data from: ", strUrl, symbol, err)
		return 9999999999
	}
	var optionsMarginRet OptionsMarginRet
	err = json.Unmarshal(body, &optionsMarginRet)
	if err != nil {
		fmt.Println("Unable to parse returned data from: ", strUrl, symbol, err, body)
		fmt.Printf("%s     %s\n", body, data)
		return 9999999999
	}
	return optionsMarginRet.Total.Total
}
