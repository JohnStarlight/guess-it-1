package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// The tester rewards each correct guess with
//
//	round(10_000_000 / (1 + width) / (n - 1))
//
// n being the size of the data set (12 500 lines in every provided set).
//
// Every provided data set follows the same model: a straight line
// (slope 0 for data sets 1-3, slope 1 for 4-5) plus independent noise,
// mostly uniform over ~101 integers, polluted with up to ~2 % of
// astronomically large outliers. The program therefore:
//
//  1. estimates the slope robustly (median of the consecutive
//     differences, immune to the outliers),
//  2. de-trends the values and drops the outliers using the median and
//     the MAD of the residuals,
//  3. picks the guessing window over the residual distribution by
//     maximizing the expected score exactly: a window covering count
//     residuals out of total earns (count/total) * round(10^7/(k*(n-1)))
//     per guess for a width of k integers. Because of the per-guess
//     rounding some widths are strictly better than others (e.g. k=94
//     over a uniform support of 101 earns 9 points per hit instead of
//     the flat 8).
const (
	expectedGuesses = 12499 // n-1 for the tester's data sets
	warmup          = 100   // guesses answered with a safe wide range
	outlierMult     = 8.0   // reject residuals beyond this many MADs from the median
	refreshEvery    = 128   // re-estimate the model every this many values
	maxSpan         = 4096  // safety cap for the residual histogram size
)

func median(sorted []int) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return float64(sorted[n/2-1]+sorted[n/2]) / 2
	}
	return float64(sorted[n/2])
}

// model is the cached prediction state, re-estimated every refreshEvery
// values: the next value at index x is expected in
// [slope*x + offset, slope*x + offset + width - 1].
type model struct {
	slope  int
	offset int
	width  int
}

// fit estimates the model from all values seen so far.
func fit(values []int) model {
	n := len(values)

	// Robust slope (Theil-Sen with a long baseline): the median of the
	// slopes of pairs half the data apart. The long gap divides the noise
	// away and the median is immune to the outliers. Consecutive
	// differences would NOT work: their median has a standard error of
	// about half a unit, so the rounded slope would flap between values.
	gap := n / 2
	slopes := make([]float64, 0, n-gap)
	for i := 0; i+gap < n; i++ {
		slopes = append(slopes, float64(values[i+gap]-values[i])/float64(gap))
	}
	sort.Float64s(slopes)
	mid := slopes[len(slopes)/2]
	slope := int(math.Round(mid))

	// De-trend, then reject outliers with median +- outlierMult * MAD.
	res := make([]int, n)
	for i, v := range values {
		res[i] = v - slope*i
	}
	sort.Ints(res)
	med := median(res)
	devs := make([]int, n)
	for i, r := range res {
		devs[i] = int(math.Abs(float64(r) - med))
	}
	sort.Ints(devs)
	cut := outlierMult * math.Max(median(devs), 1)

	inMin, inMax := math.MaxInt64, math.MinInt64
	var inliers []int
	for _, r := range res {
		if math.Abs(float64(r)-med) <= cut {
			inliers = append(inliers, r)
			if r < inMin {
				inMin = r
			}
			if r > inMax {
				inMax = r
			}
		}
	}
	span := inMax - inMin + 1
	if span > maxSpan { // paranoia; the MAD cut keeps real data far below this
		span = maxSpan
	}

	// Histogram of the inlier residuals, smoothed with a small moving
	// average: a single bin that got lucky must not look like a better
	// bet than it will be on future draws.
	hist := make([]float64, span)
	for _, r := range inliers {
		if b := r - inMin; b < span {
			hist[b]++
		}
	}
	smooth := make([]float64, span)
	const rad = 3
	for i := range hist {
		lo, hi := i-rad, i+rad
		if lo < 0 {
			lo = 0
		}
		if hi >= span {
			hi = span - 1
		}
		var sum float64
		for j := lo; j <= hi; j++ {
			sum += hist[j]
		}
		smooth[i] = sum / float64(hi-lo+1)
	}
	prefix := make([]float64, span+1)
	for i := 0; i < span; i++ {
		prefix[i+1] = prefix[i] + smooth[i]
	}

	// Exhaustively pick the window (position and width) with the highest
	// expected score under the tester's per-guess rounding. Even smoothed,
	// the count of the winning window overstates its future hit rate
	// (the maximum over thousands of candidates is biased upward by
	// sampling luck, worst for tiny windows), so penalize each count by a
	// few standard deviations (~sqrt(count)) before comparing.
	bestScore, bestLo, bestK := -1.0, inMin, span
	for k := 1; k <= span; k++ {
		perHit := math.Round(10_000_000 / (float64(k) * expectedGuesses))
		if perHit == 0 {
			break
		}
		for a := 0; a+k <= span; a++ {
			count := prefix[a+k] - prefix[a]
			if score := (count - 2.5*math.Sqrt(count)) * perHit; score > bestScore {
				bestScore, bestLo, bestK = score, inMin+a, k
			}
		}
	}
	return model{slope: slope, offset: bestLo, width: bestK}
}

func main() {
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var values []int
	var m model

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		f, err := strconv.ParseFloat(line, 64)
		if err != nil {
			continue
		}
		values = append(values, int(math.Round(f)))
		t := len(values) // index of the value to predict

		if t <= warmup {
			// Not enough data for a stable model yet: answer with a
			// generous range around the median of what we have seen.
			cp := append([]int(nil), values...)
			sort.Ints(cp)
			med := int(math.Round(median(cp)))
			fmt.Fprintf(out, "%d %d\n", med-50, med+50)
			continue
		}

		if t == warmup+1 || (t-warmup)%refreshEvery == 0 {
			m = fit(values)
		}

		lo := m.slope*t + m.offset
		fmt.Fprintf(out, "%d %d\n", lo, lo+m.width-1)
	}
}
