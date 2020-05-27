package main

func calcDiff(a, b []string, equal func(string, string) bool) []string {
	ret := []string{}
	ses := calcSES(a, b, equal)
	x, y := 0, 0
	for _, es := range ses {
		switch es {
		case 0:
			ret = append(ret, a[x])
			x, y = x+1, y+1
		case -1:
			ret = append(ret, "-" + a[x])
			x += 1
		case 1:
			ret = append(ret, "+" + b[y])
			y += 1
		}
	}

	return ret
}

func calcSES(a, b []string, equal func(string, string) bool) []int {
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

	// Shortest Edit Script hold map
	// -1: del-from-a, 0: copy, 1: add-from-b
	oldses := map[int][]int{}
	sesdown := func (k, x int, prev bool) {
		if prev {
			d := make([]int, len(oldses[max+k+1]))
			copy(d, oldses[max+k+1])
			oldses[max+k] = d
		}
		oldses[max+k] = append(oldses[max+k], 1)
	}
	sesright := func (k, x int, prev bool) {
		if prev {
			d := make([]int, len(oldses[max+k-1]))
			copy(d, oldses[max+k-1])
			oldses[max+k] = d
		}
		oldses[max+k] = append(oldses[max+k], -1)
	}
	sescopy := func (k, x int) {
		oldses[max+k] = append(oldses[max+k], 0)
	}

	for d := 0; d <= max+1; d++ {
		for k := -d; k <= d; k += 2 {
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
			for x < n && y < m && equal(a[x], b[y]) {
				x += 1
				y += 1
				sescopy(k, y)
			}

			setv(k, x)
			if x >= n && y >= m {
				return oldses[max+k]
			}
		}
	}

	return nil
}

