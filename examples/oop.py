class SimpleDict:
  def __init__(self): 
    self._data = {}
  def __setitem__(self, key, value):
    self._data[str(key)] = value
  def __getitem__(self, key):
    return self._data[str(key)]
sd = SimpleDict()
sd["foo"] = 100
sd[5] = "bar"
sd["foo"]
try:
  print(str(sd.foo))
except:
  # print(Er)
  print(str(sd["foo"]))