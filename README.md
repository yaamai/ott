# ott

Shell script runner with built-in diff checker.

## Example
```
# ott -h
Usage of ott:
  -format string
        output format (text/json) (default "text")
  -log string
        log level (default "warn")
  -mode string
        output mode (diff/actual/expected) (default "diff")
  -session-cmd string
        session command (default "sh")
  -session-mode string
        session parse mode (shell/python) (default "shell")

# cat examples/minimal.t
test-echo-a:
  $ echo a
  b

# /ott -session-cmd bash examples/minimal.t
test-echo-a:
  $ echo a
  -b
  +a

# cat examples/regex.t
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

# ./ott -session-cmd bash examples/regex.t
normal-match:
  $ echo a
  a
regex-match:
  $ echo aaa11
  aaa\d+ (re)
  $ echo aaaXX
  -aaa\d+ (re)
  +aaaXX
regex-multiline:
  $ echo -e "aaaa\nbbbb\ncccc"
  [ab]+ (re)
  [ab]+ (re)
  +cccc
```

## Install
```
git clone https://github.com/yaamai/ott.git
go build
mv ott /usr/local/bin
```
