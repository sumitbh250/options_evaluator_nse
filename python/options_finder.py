# from gevent import monkey
# monkey.patch_all()
import itertools
import multiprocessing
import requests
import json
import math
import pandas as pd
import sys
import concurrent.futures
from urllib.parse import quote_plus
from options_util import *
from options_classes import *
from options_margin import *

def get_next_url():
  url_template = "https://www.nseindia.com/api/option-chain-indices?symbol="
  for index, lot_size in options_indices:
    yield (index, lot_size, url_template+quote_plus(index))
  url_template = "https://www.nseindia.com/api/option-chain-equities?symbol="
  for stock, lot_size in options_stocks:
    yield (stock, lot_size, url_template+quote_plus(stock))

@retry_on_exception()
def fetch_latest_price(stock):
  url_template = "https://www.nseindia.com/api/quote-derivative?symbol="
  url = url_template + quote_plus(stock)
  session = requests.Session()
  response = session.get(url, headers=headers, timeout=5, cookies=cookies)
  # dajs = response.json()
  response.encoding = 'UTF-8'
  dajs = json.loads(response.text)
  return dajs['underlyingValue']

def calculate_profit_ratio(trade_arr, ranges, lot_size, current_price):
  topN = 10
  total = len(range(ranges[0], ranges[1], ranges[2]))
  max_losses_count = 0 * total
  max_losses_amount = -10000
  #profit_arr = []
  #only_loss_arr = []
  only_profit_arr = []
  losses_count = 0
  for expiry_price in range(ranges[0], ranges[1], ranges[2]):
    profit = sum([trade.profit_amount(expiry_price) for trade in trade_arr])
    # profit_arr.append(profit)
    if profit > 0:
      only_profit_arr.append(profit)
    else:
      losses_count += 1
      if losses_count > max_losses_count or profit <= max_losses_amount:
        return None, None
      #only_loss_arr.append(profit)
  #profit_arr.sort(reverse=True)
  only_profit_arr.sort(reverse=True)
  #only_loss_arr.sort(reverse=True)
  topN = min(topN, total)
  #avg_profit = sum(profit_arr[:topN])/topN
  profit_ratio = (total - losses_count)/total
  #max_profit = profit_arr[0] * lot_size
  min_profit = (sum(only_profit_arr) * lot_size)/len(only_profit_arr)
  # min_profit = only_profit_arr[-1] * lot_size if len(only_profit_arr) > 0 else 0
  #max_loss = only_loss_arr[-1] * lot_size if len(only_loss_arr) > 0 else 0
  return profit_ratio, min_profit

@retry_on_exception()
def get_trade_range(trade_arr, current_price):
  premium_arr = sorted(set([int(x.strike_price) for x in trade_arr if type(x) != NullTrade]))
  if len(premium_arr) < 2:
    return None
  range_limit_low = 0.15 * current_price
  range_limit_high = 0.15 * current_price
  ranges = (math.floor(current_price - range_limit_low),
            math.ceil(current_price + range_limit_high),
            (premium_arr[1]-premium_arr[0]))
  # print(ranges)
  return ranges

@retry_on_exception()
def get_options_data(url):
    response = session.get(url, headers=headers, timeout=5, cookies=cookies)
    return response

def find_trade_result(taken_trades, ranges, lot_size, latest_price, margin_expiry_dt, stock):
  profit_ratio, final_profit = calculate_profit_ratio(taken_trades, ranges, lot_size, latest_price)
  if profit_ratio == None  or final_profit == None:
    return None

  amount_invested = 0
  has_sell_trade = False
  for trade in taken_trades:
   if trade.type == TradeType.Sell:
     has_sell_trade = True
  if has_sell_trade:
    amount_invested += margin_calculator.get_margin_for_trades(taken_trades, stock, lot_size, margin_expiry_dt)
  amount_invested += sum([trade.premium * lot_size for trade in taken_trades if trade.type == TradeType.Buy])
  if amount_invested >= 100000:
    return None
  result = {
    'Symbol': stock,
    'Current Price': latest_price,
    'Lot Size': lot_size,
    'Ranges': ranges,
    'Trades': taken_trades,
    # 'Profit Ratio': profit_ratio,
    'Profit Ratio': (final_profit * 100)/amount_invested,
    # 'Profit Ratio': final_profit,
    'Final Profit': final_profit,
    'Total Premium': amount_invested}
  return result

def valid_trade_combo_iter(trade_arr, num_trades):
  for taken_trades in itertools.combinations_with_replacement(trade_arr, num_trades):
    num_buy = 0
    num_sell = 0
    total_calls = 0
    ce_buy_trade_strikes = set()
    ce_sell_trade_strikes = set()
    pe_buy_trade_strikes = set()
    pe_sell_trade_strikes = set()
    for trade in taken_trades:
      if trade.type == TradeType.Sell:
        num_sell += 1
        total_calls += 1
        if trade.call_type == CallType.Call:
          ce_sell_trade_strikes.add(trade.strike_price)
        elif trade.call_type == CallType.Put:
          pe_sell_trade_strikes.add(trade.strike_price)
      elif trade.type == TradeType.Buy:
        num_buy += 1
        total_calls += 1
        if trade.call_type == CallType.Call:
          ce_buy_trade_strikes.add(trade.strike_price)
        elif trade.call_type == CallType.Put:
          pe_buy_trade_strikes.add(trade.strike_price)
    if len(ce_buy_trade_strikes.intersection(ce_sell_trade_strikes)) > 0:
      continue
    if len(pe_buy_trade_strikes.intersection(pe_sell_trade_strikes)) > 0:
      continue
    if (total_calls < 2):
      continue
    elif (total_calls == 2 and (num_sell == 2 or num_buy == 2)):
      continue
    elif (total_calls == 3 and (num_sell >= 2 or num_buy > 2)):
      continue
    elif (total_calls == 4 and (num_sell > 2 or num_buy > 3)):
      continue
    yield taken_trades

def get_stock_result(arg):
  expiry_dt, margin_expiry_dt, stock, lot_size, url = arg[0], arg[1], arg[2], arg[3], arg[4]
  # print(stock)
  try:
    latest_price = fetch_latest_price(stock)
    response = get_options_data(url)
    # dajs = response.json()
    response.encoding = 'UTF-8'
    dajs = json.loads(response.text)
    ce_values = [data['CE'] for data in dajs['records']['data'] if "CE" in data and data['expiryDate'] == expiry_dt]# and abs(latest_price - data['strikePrice']) * 100/latest_price <= 10]
    pe_values = [data['PE'] for data in dajs['records']['data'] if "PE" in data and data['expiryDate'] == expiry_dt]# and abs(data['strikePrice'] - latest_price) * 100/latest_price <= 10]
    if len(ce_values) == 0 or len(pe_values) == 0:
      return []
    ce_dt = pd.DataFrame(ce_values).sort_values(['strikePrice'])
    pe_dt = pd.DataFrame(pe_values).sort_values(['strikePrice'])
    # trade_arr = []
    trade_arr = []
    for _, ce in ce_dt.iterrows():
      trade_arr.append(CEBuyTrade(ce))
      trade_arr.append(CESellTrade(ce))
    for _, pe in pe_dt.iterrows():
      trade_arr.append(PEBuyTrade(pe))
      trade_arr.append(PESellTrade(pe))
    ranges = get_trade_range(trade_arr, latest_price)
    trade_arr = list(filter(lambda trade: trade.premium !=0 and int(trade.strike_price) == trade.strike_price and trade.traded_volume > 0, trade_arr))
    trade_arr.append(NullTrade())
    if ranges == None:
      return []
    dictionary_list = []
    print(stock, len(trade_arr))
    # ii = 0
    # for taken_trades in valid_trade_combo_iter(trade_arr, 3):
    #   # ii += 1
    #   # print(ii, end='\r')
    #   result = find_trade_result(taken_trades, ranges, lot_size, latest_price, margin_expiry_dt, stock)
    #   dictionary_list.append(result)
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
      dictionary_list = list(executor.map(find_trade_result,
        valid_trade_combo_iter(trade_arr, 3),
        itertools.repeat(ranges), itertools.repeat(lot_size),
        itertools.repeat(latest_price), itertools.repeat(margin_expiry_dt),
        itertools.repeat(stock)))
    dictionary_list = [i for i in dictionary_list if i is not None]
    return dictionary_list
    #results = results.append(pd.DataFrame.from_dict(dictionary_list))
  except Exception as ex:
    # raise
    print(ex, stock)

@retry_on_exception()
def init_session():
  global cookies, headers, session
  url_oc = "https://www.nseindia.com"
  headers = {'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, '
                           'like Gecko) '
                           'Chrome/80.0.3987.149 Safari/537.36',
             'accept-language': 'en,gu;q=0.9,hi;q=0.8', 'accept-encoding': 'gzip, deflate, br'}
  session = requests.Session()
  request = session.get(url_oc, headers=headers, timeout=5)
  cookies = dict(request.cookies)


cookies = None
headers = None
session = None
margin_calculator = None
results = pd.DataFrame(columns=['Symbol', 'Current Price', 'Lot Size', 'Trades',
  'Ranges','Profit Ratio', 'Final Profit', 'Total Premium'])
MAX_PROCESSES = 4
MAX_THREADS = 10 # per process
CHUNKSIZE = 20

def proc_worker(ps):
  import concurrent.futures as cf
  with cf.ThreadPoolExecutor(max_workers=MAX_THREADS) as e:
      result = list(e.map(get_stock_result, ps))
  return result

def main():
  global margin_calculator, results
  if len(sys.argv) != 2:
    sys.exit('Profide filename to store dataframe')
  init_session()
  margin_calculator = ZerodhaMargin()
  #margin_calculator = ProStocksMargin()
  arr = []
  for stock, lot_size, url in get_next_url():
    expiry_dt = '29-Apr-2021'
    expiry_dt2 = '21APR'
    # expiry_dt2 = '20210429'
    arr.append((expiry_dt, expiry_dt2, stock, lot_size, url))
  with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
    all_stocks_result_list = list(executor.map(get_stock_result, arr))
    for per_stock_result_list in all_stocks_result_list:
      results = results.append(pd.DataFrame.from_dict(per_stock_result_list))
  # with concurrent.futures.ProcessPoolExecutor(max_workers=MAX_PROCESSES) as e:
  #   for chunk_result in e.map(proc_worker, (arr[i: i+CHUNKSIZE] for i in range(0, len(arr), CHUNKSIZE))):
  #     for per_stock_result in chunk_result:
  #           results = results.append(pd.DataFrame.from_dict(per_stock_result))
  df_file = './' + sys.argv[1] + '.pkl'
  results.to_pickle(df_file)

if __name__ == "__main__":
    main()
