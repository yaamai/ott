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
date:
  $ date
  aaa
multiline:
  $ export B=200
  $ echo a &&\
  > date &&\
  > echo $B
  b

