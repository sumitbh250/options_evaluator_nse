import requests
import json
import pandas as pd

options_stocks = [('AARTIIND', 425), ('ACC', 500), ('ADANIENT',2000), ('ADANIPORTS',2500), ('AMARAJABAT',1000), ('AMBUJACEM',3000), ('APOLLOHOSP',500), ('APOLLOTYRE',5000), ('ASHOKLEY', 9000), ('ASIANPAINT', 300), ('AUROPHARMA', 650), ('AXISBANK', 1200), ('BAJAJ-AUTO', 250), ('BAJAJFINSV', 125), ('BAJFINANCE', 250), ('BALKRISIND', 400), ('BANDHANBNK', 1800), ('BANKBARODA', 11700), ('BATAINDIA', 550), ('BEL', 7600), ('BERGEPAINT', 1100), ('BHARATFORG', 1500), ('BHARTIARTL', 1851), ('BHEL', 21000), ('BIOCON', 2300), ('BOSCHLTD', 50), ('BPCL', 1800), ('BRITANNIA', 200), ('CADILAHC', 2200), ('CANBK', 5400), ('CHOLAFIN', 2500), ('CIPLA', 1300), ('COALINDIA', 4200), ('COFORGE', 375), ('COLPAL', 700), ('CONCOR', 1563), ('CUMMINSIND', 1200), ('DABUR', 1250), ('DIVISLAB', 200), ('DLF', 3300), ('DRREDDY', 125), ('EICHERMOT', 350), ('ESCORTS', 550), ('EXIDEIND', 3600), ('FEDERALBNK', 10000), ('GAIL', 6100), ('GLENMARK', 1150), ('GMRINFRA', 22500), ('GODREJCP', 1000), ('GODREJPROP', 650), ('GRASIM', 950), ('HAVELLS', 1000), ('HCLTECH', 700), ('HDFC', 300), ('HDFCAMC', 200), ('HDFCBANK', 550), ('HDFCLIFE', 1100), ('HEROMOTOCO', 300), ('HINDALCO', 4300), ('HINDPETRO', 2700), ('HINDUNILVR', 300), ('IBULHSGFIN', 3100), ('ICICIBANK', 1375), ('ICICIGI', 425), ('ICICIPRULI', 1500), ('IDEA', 70000), ('IDFCFIRSTB', 19000), ('IGL', 1375), ('INDIGO', 500), ('INDUSINDBK', 900), ('INDUSTOWER', 2800), ('INFY', 600), ('IOC', 6500), ('ITC', 3200), ('JINDALSTEL', 5000), ('JSWSTEEL', 2700), ('JUBLFOOD', 250), ('KOTAKBANK', 400), ('L&TFH', 8924), ('LALPATHLAB', 250), ('LICHSGFIN', 2000), ('LT', 575), ('LUPIN', 850), ('M&M', 1400), ('M&MFIN', 4000), ('MANAPPURAM', 6000), ('MARICO', 2000), ('MARUTI', 100), ('MCDOWELL-N', 1250), ('MFSL', 1300), ('MGL', 600), ('MINDTREE', 800), ('MOTHERSUMI', 7000), ('MRF', 10), ('MUTHOOTFIN', 750), ('NATIONALUM', 17000), ('NAUKRI', 250), ('NESTLEIND', 50), ('NMDC', 6700), ('NTPC', 5700), ('ONGC', 7700), ('PAGEIND', 30), ('PEL', 550), ('PETRONET', 3000), ('PFC', 6200), ('PIDILITIND', 500), ('PNB', 16000), ('POWERGRID', 4000), ('PVR', 407), ('RAMCOCEM', 850), ('RBLBANK', 2900), ('RECLTD', 6000), ('RELIANCE', 250), ('SAIL', 19000), ('SBILIFE', 750), ('SBIN', 3000), ('SHREECEM', 50), ('SIEMENS', 550), ('SRF', 125), ('SRTRANSFIN', 800), ('SUNPHARMA', 1400), ('SUNTV', 1500), ('TATACHEM', 2000), ('TATACONSUM', 1350), ('TATAMOTORS', 5700), ('TATAPOWER', 13500), ('TATASTEEL', 1700), ('TCS', 300), ('TECHM', 1200), ('TITAN', 750), ('TORNTPHARM', 3000), ('TORNTPOWER', 3000), ('TVSMOTOR', 1400), ('UBL', 700), ('ULTRACEMCO', 200), ('UPL', 1300), ('VEDL', 6200), ('VOLTAS', 1000), ('WIPRO', 3200), ('ZEEL', 3000)]

options_indices = [('NIFTY', 75, (14000, 15201, 50))] #, ('FINNIFTY', 40), ('BANKNIFTY', 25)]

import requests

url_oc = "https://www.nseindia.com"
headers = {'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, '
                         'like Gecko) '
                         'Chrome/80.0.3987.149 Safari/537.36',
           'accept-language': 'en,gu;q=0.9,hi;q=0.8', 'accept-encoding': 'gzip, deflate, br'}
session = requests.Session()
request = session.get(url_oc, headers=headers, timeout=5)
cookies = dict(request.cookies)

def get_next_url():
  # url_template = "https://www.nseindia.com/api/option-chain-equities?symbol="
  # for stock, lot_size in options_stocks:
  #   yield (stock, lot_size, url_template+stock)
  url_template = "https://www.nseindia.com/api/option-chain-indices?symbol="
  for index, lot_size, ranges in options_indices:
    yield (index, lot_size, ranges, url_template+index)

def fetch_latest_price(stock):
  url_template = "https://www.nseindia.com/api/quote-derivative?symbol="
  url = url_template + stock
  session = requests.Session()
  response = session.get(url, headers=headers, timeout=5, cookies=cookies)
  dajs = response.json()
  return dajs['underlyingValue']

def calculate_profit(ranges, ce_strike, ce_premium, pe_strike, pe_premium):
  total_buckets = 0
  total_profitable_buckets = 0
  for expiry_price in range(ranges[0], ranges[1], ranges[2]):
    profit = 0
    if expiry_price < ce_strike:
      profit -= ce_premium
    else:
      profit += (expiry_price - ce_strike - ce_premium)
    if expiry_price < pe_strike:
      profit += (pe_strike - expiry_price - pe_premium)
    else:
      profit -= pe_premium
    if profit > 0:
      total_profitable_buckets += 1
    total_buckets += 1
  return (total_profitable_buckets/total_buckets)

def fetch_oi(expiry_dt, stock, lot_size, ranges, url):
  global results
  print(stock)
  try:
    latest_price = fetch_latest_price(stock)
    response = session.get(url, headers=headers, timeout=5, cookies=cookies)
    dajs = response.json()
    ce_values = [data['CE'] for data in dajs['records']['data'] if "CE" in data and data['expiryDate'] == expiry_dt and abs(latest_price - data['strikePrice']) * 100/latest_price < 3]
    pe_values = [data['PE'] for data in dajs['records']['data'] if "PE" in data and data['expiryDate'] == expiry_dt and abs(data['strikePrice'] - latest_price) * 100/latest_price < 3]
    if len(ce_values) == 0 or len(pe_values) == 0:
      return
    ce_dt = pd.DataFrame(ce_values).sort_values(['strikePrice'])
    pe_dt = pd.DataFrame(pe_values).sort_values(['strikePrice'])
    for _, ce in ce_dt.iterrows():
      for _, pe in pe_dt.iterrows():
        if pe['lastPrice'] == 0 or ce['lastPrice'] == 0:
          continue
        price_diff = max(pe['strikePrice'], ce['strikePrice']) - min(pe['strikePrice'], ce['strikePrice'])
        premium_sum = ce['lastPrice'] + pe['lastPrice']
        #max_loss = (premium_sum - price_diff) * lot_size
        profit_ratio = calculate_profit(ranges, ce['strikePrice'], ce['lastPrice'], pe['strikePrice'], pe['lastPrice'])
        results = results.append({'Symbol': stock, 'Current Price': latest_price,
          'Lot Size': lot_size, 'CE Strike': ce['strikePrice'],
          'CE Premium': ce['lastPrice'], 'CE OI': ce['openInterest'] ,
          'PE Strike': pe['strikePrice'],
          'PE Premium': pe['lastPrice'], 'PE OI': pe['openInterest'],
          'Profit Ratio': profit_ratio,
          #'Max Loss': max_loss,
          'Total Premium': (premium_sum * lot_size)}, ignore_index=True)
  except Exception as ex:
    print(ex)

def main():
  results = pd.DataFrame(columns=['Symbol', 'Current Price', 'Lot Size', 'CE Strike', 'CE Premium', 'CE OI', 'PE Strike', 'PE Premium', 'PE OI',
    #'Max Loss',
    'Profit Ratio', 'Total Premium'])
  for stock, lot_size, ranges, url in get_next_url():
    expiry_dt = '11-Feb-2021'
    fetch_oi(expiry_dt, stock, lot_size, ranges, url)
  #results = results.sort_values(by=['Max Loss'])
  results = results.sort_values(by=['Profit Ratio'], ascending=False)

  pd.set_option('display.max_rows', None)
  pd.set_option('display.max_columns', None)
  pd.set_option('display.width', 2000)
  pd.set_option('display.float_format', '{:20,.2f}'.format)
  pd.set_option('display.max_colwidth', None)
  print(results)

if __name__ == "__main__":
    main()
