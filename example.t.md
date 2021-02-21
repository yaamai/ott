


# test-title

```
<< for i in $(seq 3); do
>   echo \# echo $i
>   echo $1
> done

# echo a
a

$ echo b
b

# for i in $(seq 3); do
> echo $i
> sleep 1
> done
1
2
3
```



# test A

  ```
  # echo a
  a
  ```

# test B

```
# aaaa
(rc==127)
```
// TODO: matcher(re)
// TODO: matcher(rc)
// TODO: matcher(ignore)
// TODO: matcher(has)

# test C

```
# for i in $(seq 3); do
> echo $i
> sleep 1
> done
```

# docker

```
# docker pull nginx:1.19.3
docker.io/library/nginx:1.19.3 (has)
```

# curl

```
# curl -Lo/dev/null https://www.google.com
```

# non-fenced code block
> aaa

    # uname -a
    (rc==1)

# exit
```
# exit 1
```


# todo
- step-wide matcher
  - [x] (rc) matcher
  - [x] (has) matcher
- line-based matcher
  - [ ] /.../ style regex match
  - [ ] normal match
- other
  - [ ] named pipe (for curl verbose output match and main output match)
  - [ ] `prev` pipe (execute command and buffering, read from buffer with pipe(|) style)
  - [ ] show diff (line-based)
  - [ ] show diff (word-based)
  - [x] non-fenced code block
  - [ ] ansi
  - [ ] per loop checker in for
    - [ ] use `set -e`?
    - [ ] generate command-line strings in bash then send to ott and execute each
      ```
      << for i in $(seq 3); do
      >   echo \# echo $i
      >   echo $1
      > done

      # echo 1
      1
      # echo 2
      2
      # echo 3
      3
      ```