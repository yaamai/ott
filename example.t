# can write comment or meta for file here
# meta
#  file: a

# can write comment or meta for test-case("echo-a") here
# meta
#  a: 100
#  b: 100
echo-a:
  # can write test-step(below) comment
  $ echo -e "a\nb"
  a
  b

  # and here
  $ echo -e "c\nd"
  aaaa
  d
multiline:
  $ export B=200

  # can write multi-line command
  $ echo a &&
  > date &&
  > echo $B
  b

# test-case name prefixed with `(setup|teardown)-per-(run|file|case)`
# running each timing
setup-per-run-echo:
  $ echo n
  n
