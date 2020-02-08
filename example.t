# meta
#  a: 100
#  b: 100
echo-a:
  $ echo -e "a\nb"
  a
  b
  $ echo -e "c\nd"
  aaaa
  d
multiline:
  $ export B=200
  $ echo a &&
  > date &&
  > echo $B
  b

setup-per-run:
  $ date "+%Y"

setup-per-file:
  $ date "+%m"

setup-per-case:
  $ date "+%d"
