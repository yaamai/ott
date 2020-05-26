package main

import (
	"fmt"
)

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

/*
 * a=abc(len=n,idx=x)
 * b=bbc(len=m,idx=y)
 * y = x - k
 * k is intercept
 * v[k] = k's edit count
 */
func calc_diff(a, b []string) int {
	n, m := len(a), len(b)
	max := m + n

	// hold v (-max to +max)
	v := make([]int, max*2 + 1)
	setv := func (k int, val int) {
		v[max + k] = val
	}
	getv := func (k int) int {
		return v[max + k]
	}

	oldses := map[int][]string{}
	sesdown := func (ses *[]string, k, y int, prev bool) {
		if prev {
			*ses = oldses[max+k+1]
		}
		*ses = append(*ses, "down")
		fmt.Println("down")
	}
	sesright := func (ses *[]string, k, y int, prev bool) {
		if prev {
			*ses = oldses[max+k-1]
		}
		*ses = append(*ses, "right")
		fmt.Println("right")
	}
	sescopy := func (ses *[]string, k, y int) {
		*ses = append(*ses, "copy")
		fmt.Println("copy")
	}

	for d := 0; d <= max+1; d++ {
		fmt.Println("dloop: ", d)
		for k := -d; k <= d; k += 2 {
			fmt.Println("kloop: ", k)
			fmt.Println("v: ", v)

			y := 0
			ses := []string{}
			if d == 0 {
				y = 0
			} else if k == -d {
				// lower-k has same previous k's v
				y = getv(k+1) + 1
				sesdown(&ses, k, y, true)
			} else if k == d {
				// upper-k
				y = getv(k-1)
				sesright(&ses, k, y, true)
			} else {
				lower := getv(k+1) + 1
				upper := getv(k-1)
				// FIXME: maybe cant' determine best-path without snake
				if upper > lower {
					y = upper
					sesdown(&ses, k, y, true)
				} else {
					y = lower
					sesright(&ses, k, y, true)
				}
			}

			x := y + k
			fmt.Println("pos: ", x, y)
			for x < n && y < m && a[x] == b[y] {
				x += 1
				y += 1
				sescopy(&ses, k, y)
			}

			setv(k, y)
			oldses[max+k] = ses

			if x >= m && y >= n {
				fmt.Println("SES:", ses)
				return d
			}
		}
	}

	return 0
}

func calc_diff_2(a, b []string) int {
	m, n := len(a), len(b)
	v := make([]int, 2*(m+n)+1)
	hist := map[int][]int{}
	offset := m + n

	for d := 0; d <= (m+n); d++ {
		fmt.Println("dloop: ", d)
		for k := -d; k <= d; k += 2 {
			fmt.Println("kloop: ", k)
			fmt.Println("hist: ", hist)
			ses := []int{}
			i := 0
			if d == 0 {
				i = 0
			} else if k == -d {
				i = v[offset + k + 1] + 1
				ses = hist[offset + k + 1]
				ses = append(ses, []int{1}...)
				fmt.Println("down")
			} else if k == d {
				i = v[offset + k - 1]
				ses = hist[offset + k - 1]
				ses = append(ses, []int{2}...)
				fmt.Println("right")
			} else {
				upper := v[offset + k + 1] + 1
				lower := v[offset + k - 1]
				if upper > lower {
					i = upper
					ses = hist[offset + k + 1]
					ses = append(ses, []int{1}...)
					fmt.Println("down")
				} else {
					i = lower
					ses = hist[offset + k - 1]
					ses = append(ses, []int{2}...)
					fmt.Println("right")
				}
			}
			fmt.Println("i: ", i)
			fmt.Println("v: ", v)

			for (i < m && (i+k) < n && a[i] == b[i + k]) {
				fmt.Println("snake: ", i, k, m, n)
				ses = append(ses, []int{0}...)
				i += 1
			}

			if (k == (n - m) && i == m) {
				fmt.Println("SES: ", ses)
				return d
			}

			v[offset + k] = i
			fmt.Println("Vd(k): =", i, d, k)
			hist[offset + k] = ses
		}
	}

	return 0
}


/*
 * a=abc(len=n,idx=x)
 * b=bbc(len=m,idx=y)
 * y = x - k
 * k is intercept
 * v[k] = k's edit count
 */
func calc_diff_reversed(a, b []string) int {
	n, m := len(a), len(b)
	max := m + n

	// hold v (-max to +max)
	v := make([]int, max*2 + 1)
	setv := func (k int, val int) {
		v[max + k] = val
	}
	getv := func (k int) int {
		return v[max + k]
	}

	oldses := map[int][]string{}
	sesdown := func (ses *[]string, k, y int, prev bool) {
		if prev {
			*ses = oldses[max+k+1]
		}
		*ses = append(*ses, "down")
		fmt.Println("down")
	}
	sesright := func (ses *[]string, k, y int, prev bool) {
		if prev {
			*ses = oldses[max+k-1]
		}
		*ses = append(*ses, "right")
		fmt.Println("right")
	}
	sescopy := func (ses *[]string, k, y int) {
		*ses = append(*ses, "copy")
		fmt.Println("copy")
	}

	for d := 0; d <= max+1; d++ {
		fmt.Println("dloop: ", d)
		for k := -d; k <= d; k += 2 {
			fmt.Println("kloop: ", k)
			fmt.Println("v: ", v)

			y := 0
			ses := []string{}
			if d == 0 {
				y = 0
			} else if k == -d {
				// lower-k has same previous k's v
				y = getv(k+1) + 1
				sesdown(&ses, k, y, true)
			} else if k == d {
				// upper-k
				y = getv(k-1)
				sesright(&ses, k, y, true)
			} else {
				lower := getv(k+1) + 1
				upper := getv(k-1)
				// FIXME: maybe cant' determine best-path without snake
				if upper > lower {
					y = upper
					sesdown(&ses, k, y, true)
				} else {
					y = lower
					sesright(&ses, k, y, true)
				}
			}

			x := y + k
			fmt.Println("pos: ", x, y)
			for x < m && y < n && a[y] == b[x] {
				x += 1
				y += 1
				sescopy(&ses, k, y)
			}

			setv(k, x)
			oldses[max+k] = ses

			if x >= m && y >= n {
				fmt.Println("SES:", ses)
				return d
			}
		}
	}

	return 0
}
/*
 * a=abc(len=n,idx=y)
 * b=bbc(len=m,idx=x)
 * y = x - k
 * k is intercept
 * v[k] = k's edit count
 */
func calc_diff_maybeok(a, b []string) int {
	n, m := len(a), len(b)
	max := m + n

	// hold v (-max to +max)
	v := make([]int, max*2 + 1)
	setv := func (k int, val int) {
		v[max + k] = val
	}
	getv := func (k int) int {
		return v[max + k]
	}

	oldses := map[int][]string{}

	for d := 0; d <= max+1; d++ {
		fmt.Println("dloop: ", d)
		for k := -d; k <= d; k += 2 {
			fmt.Println("kloop: ", k)
			fmt.Println("v: ", v)

			x := 0
			ses := []string{}
			if d == 0 {
				x = 0
			} else if k == -d {
				// lower-k has same previous k's v
				x = getv(k+1)
				ses = oldses[max+k+1]
				ses = append(ses, []string{"down"}...)
				fmt.Println("down")
			} else if k == d {
				// upper-k
				x = getv(k-1) + 1
				ses = oldses[max+k-1]
				ses = append(ses, []string{"right"}...)
				fmt.Println("right")
			} else {
				lower := getv(k-1) + 1
				upper := getv(k+1)
				// FIXME: maybe cant' determine best-path without snake
				if upper > lower {
					x = upper
					ses = oldses[max+k+1]
					ses = append(ses, []string{"down"}...)
					fmt.Println("down")
				} else {
					x = lower
					ses = oldses[max+k-1]
					ses = append(ses, []string{"right"}...)
					fmt.Println("right")
				}
			}

			y := x - k
			fmt.Println("pos: ", x, y)
			for x < n && y < m && a[x] == b[y] {
				x += 1
				y += 1
				fmt.Println("diagonal")
				ses = append(ses, []string{"copy"}...)
			}

			setv(k, x)
			oldses[max+k] = ses

			if x >= n && y >= m {
				fmt.Println("SES:", ses)
				return d }
		}
	}

	return 0
}
