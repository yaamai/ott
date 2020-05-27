normal-match:
  $ echo a
  a

regex-match:
  $ echo aaa11
  aaa\d+ (re)

  $ echo aaaXX
  aaa\d+ (re)

regex-multiline:
  $ echo -e "aaaa\nbbbb\ncccc"
  [ab]+ (re)
  [ab]+ (re)
