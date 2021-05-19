import pandas as pd
import math
import sys
from options_classes import *

def prRed(skk): print("\033[91m {}\033[00m" .format(skk))
def prGreen(skk): print("\033[92m {}\033[00m" .format(skk))

def main():
  if len(sys.argv) != 2:
    sys.exit(1)
  df_file = './' + sys.argv[1] + '.pkl'
  df_text_file = './' + sys.argv[1] + '.txt'
  results = pd.read_pickle(df_file)
  # results = results.sort_values(by=['Max Loss'])
  # results = results.sort_values(by=['Final Profit', 'Profit Ratio'], ascending=False, ignore_index=True)
  # results = results.sort_values(by=['Total Premium', 'Profit Ratio', 'Final Profit'], ascending=[True, False, False], ignore_index=True)
  results = results.sort_values(by=['Profit Ratio', 'Final Profit'], ascending=False, ignore_index=True)
  # results = results.groupby('Symbol').head(1).reset_index(drop=True)

  pd.set_option('display.max_rows', None)
  pd.set_option('display.max_columns', None)
  pd.set_option('display.width', 2000)
  pd.set_option('display.float_format', '{:20,.2f}'.format)
  pd.set_option('display.max_colwidth', None)
  with open(df_text_file, 'w') as f:
    f.write(results.to_string())

  print(results.index)
  index = -1
  profit_arr = []
  while True:
    print("Enter index of the data frame")
    inp_str = input()
    if inp_str == "":
      index+=1
    else:
      index = int(inp_str)
    res = results.iloc[index]
    current_price = res['Current Price']
    range_limit_low = 0.2 * current_price
    range_limit_high = 0.2 * current_price
    ranges = (math.floor(current_price - range_limit_low),
              math.ceil(current_price + range_limit_high),
              res['Ranges'][2])
    # ranges = res['Ranges']
    taken_trades = res['Trades']
    lot_size = res['Lot Size']
    print(res)
    for expiry_price in range(ranges[0], ranges[1], ranges[2]):
      profit = int(sum([trade.profit_amount(expiry_price) for trade in taken_trades]) * lot_size)
      if profit <= 0:
        prRed(str(expiry_price) + " " + str(profit))
      else:
        prGreen(str(expiry_price) + " " + str(profit))

if __name__ == "__main__":
    main()
