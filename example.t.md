# test A

  ```
  # echo a
  a
  ```

# test B

```
# date
(rc)
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
docker.io/library/nginx/1.19.3 (has)
```

# curl

```
# curl -vLo/dev/null https://www.google.com
```

# todo
- step-wide matcher
  - [ ] (rc) matcher
  - [ ] (has) matcher
- line-based matcher
  - [ ] /.../ style regex match
  - [ ] normal match
- other
  - [ ] named pipe (for curl verbose output match and main output match)
  - [ ] `prev` pipe (execute command and buffering, read from buffer with pipe(|) style)