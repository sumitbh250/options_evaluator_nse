import enum
from functools import wraps

class TradeType(enum.Enum):
  Null = 1
  Buy = 2
  Sell = 3

class CallType(enum.Enum):
  Null = 1
  Call = 2
  Put = 3

class Trade():
  def __init__(self, trade_data = None):
    self.strike_price = 0 if trade_data is None else trade_data['strikePrice']
    self.premium = 0 #if trade_data is None else trade_data['lastPrice']
    self.open_interest = 0 if trade_data is None else trade_data['openInterest']
    self.traded_volume = 0 if trade_data is None else trade_data['totalTradedVolume']
    self.type = TradeType.Null
    self.call_type = CallType.Null

class PEBuyTrade(Trade):
  def __init__(self, trade_data = None):
    super(PEBuyTrade, self).__init__(trade_data)
    self.premium = 0 if trade_data is None else trade_data['askPrice']
    self.type = TradeType.Buy
    self.call_type = CallType.Put
  def profit_amount(self, expiry_price):
    if (expiry_price < self.strike_price):
      return (self.strike_price - expiry_price - self.premium)
    else:
      return (-1 * self.premium)
  def __str__(self):
    return "PE Buy| Strike Price: " + str(self.strike_price) + " Premium: " + \
      str(self.premium) + " TV: "+ str(self.traded_volume)
  def __hash__(self):
    return hash((1, self.strike_price))

class PESellTrade(Trade):
  def __init__(self, trade_data = None):
    super(PESellTrade, self).__init__(trade_data)
    self.premium = 0 if trade_data is None else trade_data['bidprice']
    self.type = TradeType.Sell
    self.call_type = CallType.Put
  def profit_amount(self, expiry_price):
    if (expiry_price < self.strike_price):
      return (-1 * (self.strike_price - expiry_price - self.premium))
    else:
      return self.premium
  def __str__(self):
    return "PE Sell| Strike Price: " + str(self.strike_price) + " Premium: " + \
      str(self.premium) + " TV: "+ str(self.traded_volume)
  def __hash__(self):
    return hash((2, self.strike_price))

class CEBuyTrade(Trade):
  def __init__(self, trade_data = None):
    super(CEBuyTrade, self).__init__(trade_data)
    self.premium = 0 if trade_data is None else trade_data['askPrice']
    self.type = TradeType.Buy
    self.call_type = CallType.Call
  def profit_amount(self, expiry_price):
    if (expiry_price < self.strike_price):
      return (-1 * self.premium)
    else:
      return (expiry_price - self.strike_price - self.premium)
  def __str__(self):
    return "CE Buy| Strike Price: " + str(self.strike_price) + " Premium: " + \
      str(self.premium) + " TV: "+ str(self.traded_volume)
  def __hash__(self):
    return hash((3, self.strike_price))

class CESellTrade(Trade):
  def __init__(self, trade_data = None):
    super(CESellTrade, self).__init__(trade_data)
    self.premium = 0 if trade_data is None else trade_data['bidprice']
    self.type = TradeType.Sell
    self.call_type = CallType.Call
  def profit_amount(self, expiry_price):
    if (expiry_price < self.strike_price):
      return self.premium
    else:
      return (-1 * (expiry_price - self.strike_price - self.premium))
  def __str__(self):
    return "CE Sell| Strike Price: " + str(self.strike_price) + " Premium: " + \
      str(self.premium) + " TV: "+ str(self.traded_volume)
  def __hash__(self):
    return hash((4, self.strike_price))

class NullTrade(Trade):
  def profit_amount(self, expiry_price):
    return 0
  def __str__(self):
    return "Null Trade"
  def __hash__(self):
    return hash((0, 0))

def retry_on_exception(retries = 3):
  def deco_retry(f):
    @wraps(f)
    def f_retry(*args, **kwargs):
      retries_left = retries
      while retries_left > 1:
        try:
          return f(*args, **kwargs)
        except Exception as ex:
          print(ex, retries_left)
          retries_left -= 1
      return f(*args, **kwargs)
    return f_retry
  return deco_retry
