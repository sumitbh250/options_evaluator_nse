import requests
import json
from options_classes import *

class ProStocksMargin():
  def __init__(self):
    self.init_session()

  @retry_on_exception()
  def init_session(self):
    url_oc = "https://www.prostocks.com/equity-fo-span-margin-calculator.html"
    self.headers = {
      'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:84.0) Gecko/20100101 Firefox/84.0',
      'Accept': 'application/json, text/javascript, */*; q=0.5',
      'Accept-Language': 'en-US,en;q=0.5',
      'Referer': 'https://www.prostocks.com/equity-fo-span-margin-calculator.html',
      'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
      'X-Requested-With': 'XMLHttpRequest',
      'Origin': 'https://www.prostocks.com',
      'Connection': 'keep-alive',
      'TE': 'Trailers',
    }
    self.session = requests.Session()
    request = self.session.get(url_oc, headers=self.headers, timeout=5)
    self.cookies = dict(request.cookies)
    self.span_url = "https://www.prostocks.com/index.php?option=com_custom&view=fomargincalc"

  @retry_on_exception()
  def get_margin_for_trades(self, trade_arr, symbol, lot_size, expiry_dt):
    data = [('product', 'option')]
    trades_freq = dict()
    for trade in trade_arr:
      if trade.type == TradeType.Null or trade.call_type == CallType.Null:
        continue
      elif trade in trades_freq:
        trades_freq[trade] += 1
      else:
        trades_freq[trade] = 1
    count = 0
    old_option_data = []
    for trade, freq in trades_freq.items():
      qty = str(lot_size * freq)
      buy_or_sell = 'buy' if trade.type == TradeType.Buy else 'sell' if trade.type == TradeType.Sell else ''
      ce_or_pe = 'C' if trade.call_type == CallType.Call else 'P' if trade.call_type == CallType.Put else ''
      if count == 0:
        count += 1
        data.append(('contract_detail', symbol + "+" + expiry_dt + "+" + "181"))
        data.append(('option_type', ce_or_pe))
        data.append(('strike_price', int(trade.strike_price)))
        data.append(('qty', qty))
        data.append(('trade', buy_or_sell))
        data.append(('calculate', 'true'))
        data.append(('tmpl', 'component'))
        data.append(('oldFutureData', '[]'))
      else:
        old_option_data.append("\""+symbol+"+"+expiry_dt+"+"+"181"+str(qty)+"+"+buy_or_sell+"+"+ce_or_pe+"+"+str(trade.strike_price)+"\"")
    data.append(('oldOptionData', '['+ ','.join(old_option_data) +']'))
    response = self.session.post(self.span_url, headers=self.headers, timeout=5, cookies=self.cookies, data=data)
    print(data)
    print(response.json())
    dajs = response.json()
    margin = 0
    if 'tableResult' in dajs:
      if 'initialMargin' in dajs['tableResult']:
        margin += dajs['tableResult']['initialMargin']
      if 'exposureMargin' in dajs['tableResult']:
        margin += dajs['tableResult']['exposureMargin']
      return margin
    print(data)
    print(response.json())
    return 999999999

class ZerodhaMargin():
  def __init__(self):
    self.init_session()

  @retry_on_exception()
  def init_session(self):
    url_oc = "https://zerodha.com/"
    self.headers = {
      'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:84.0) Gecko/20100101 Firefox/84.0',
      'Accept': 'application/json, text/javascript, */*; q=0.01',
      'Accept-Language': 'en-US,en;q=0.5',
      'Referer': 'https://zerodha.com/',
      'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
      'X-Requested-With': 'XMLHttpRequest',
      'Origin': 'https://zerodha.com',
      'Connection': 'keep-alive',
      'TE': 'Trailers',
    }
    self.session = requests.Session()
    request = self.session.get(url_oc, headers=self.headers, timeout=5)
    self.cookies = dict(request.cookies)
    self.span_url = "https://zerodha.com/margin-calculator/SPAN"

  @retry_on_exception()
  def get_margin_for_trades(self, trade_arr, symbol, lot_size, expiry_dt):
    data = [('action', 'calculate')]
    trades_freq = dict()
    for trade in trade_arr:
      if trade.type == TradeType.Null or trade.call_type == CallType.Null:
        continue
      elif trade in trades_freq:
        trades_freq[trade] += 1
      else:
        trades_freq[trade] = 1
    for trade, freq in trades_freq.items():
      data.append(('exchange[]', 'NFO'))
      data.append(('product[]', 'OPT'))
      scrip = symbol + expiry_dt
      data.append(('scrip[]', scrip))
      ce_or_pe = 'CE' if trade.call_type == CallType.Call else 'PE' if trade.call_type == CallType.Put else ''
      data.append(('option_type[]', ce_or_pe))
      data.append(('strike_price[]', int(trade.strike_price)))
      data.append(('qty[]', lot_size*freq))
      buy_or_sell = 'buy' if trade.type == TradeType.Buy else 'sell' if trade.type == TradeType.Sell else ''
      data.append(('trade[]', buy_or_sell))
    response = self.session.post(self.span_url, headers=self.headers, timeout=5, cookies=self.cookies, data=data)
    # dajs = response.json()
    response.encoding = 'UTF-8'
    dajs = json.loads(response.text)
    if 'total' in dajs and 'total' in dajs['total']:
      return dajs['total']['total']
    print(data)
    print(dajs)
    return 999999999


def main():
  zerodha_margin = ZerodhaMargin()
  session = zerodha_margin.session
  headers = zerodha_margin.headers
  cookies = zerodha_margin.cookies
  url = "https://zerodha.com/margin-calculator/SPAN"
  data = [
  ('action', 'calculate'),
  ('exchange[]', 'NFO'),
  ('product[]', 'OPT'),
  ('scrip[]', 'NIFTY29APR'),
  ('option_type[]', 'CE'),
  ('strike_price[]', '15000'),
  ('qty[]', '75'),
  ('trade[]', 'buy'),
  ('exchange[]', 'NFO'),
  ('product[]', 'OPT'),
  ('scrip[]', 'NIFTY29APR'),
  ('option_type[]', 'CE'),
  ('strike_price[]', '15400'),
  ('qty[]', '75'),
  ('trade[]', 'buy'),
  ('exchange[]', 'NFO'),
  ('product[]', 'OPT'),
  ('scrip[]', 'NIFTY21FEB'),
  ('option_type[]', 'CE'),
  ('strike_price[]', '14700'),
  ('qty[]', '75'),
  ('trade[]', 'sell'),
  ]
  response = session.post(url, headers=headers, timeout=5, cookies=cookies, data=data)
  print(response)
  print(response.json()['total']['total'])

if __name__ == "__main__":
    # main()
    main2()
