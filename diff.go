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

	oldses := map[int][]int{}
	sesdown := func (k, x int, prev bool) {
		if prev {
			d := make([]int, len(oldses[max+k+1]))
			copy(d, oldses[max+k+1])
			oldses[max+k] = d
			fmt.Println("down2-from:", max+k+1, oldses[max+k])
		}
		oldses[max+k] = append(oldses[max+k], 1)
		fmt.Println("down2", x)
	}
	sesright := func (k, x int, prev bool) {
		if prev {
			d := make([]int, len(oldses[max+k-1]))
			copy(d, oldses[max+k-1])
			oldses[max+k] = d
			fmt.Println("right-from:", max+k-1, oldses[max+k])
		}
		oldses[max+k] = append(oldses[max+k], -1)
		fmt.Println("right", x)
	}
	sescopy := func (k, x int) {
		oldses[max+k] = append(oldses[max+k], 0)
		fmt.Println("copy", x)
	}

	for d := 0; d <= max+1; d++ {
		fmt.Println("dloop: ", d)
		for k := -d; k <= d; k += 2 {
			fmt.Println("kloop: ", k)
			fmt.Println("v: ", v)
			fmt.Println("ses: ", oldses)

			x := 0
			if d == 0 {
				x = getv(k+1)
			} else if k == -d || k != d && getv(k-1) < getv(k+1){
				x = getv(k+1)
				sesdown(k, x, true)
			} else {
				x = getv(k-1) + 1
				sesright(k, x, true)
			}
			y := x - k
			fmt.Println("pos: ", x, y)
			for x < n && y < m && a[x] == b[y] {
				x += 1
				y += 1
				sescopy(k, y)
			}

			setv(k, x)
			fmt.Println("save:", max+k)

			if x >= n && y >= m {
				fmt.Println("SES:", oldses[max+k])
				return d
			}
		}
	}

	return 0
}

