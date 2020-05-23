miniaml:
  $ print("a")
  a

has-diff:
  $ print("b")
  a

multi-line:
  $ for i in range(3):
  >   print(i)
  1
  2
  3
