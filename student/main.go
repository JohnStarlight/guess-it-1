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
// For values uniformly distributed over a support of s integers, a range
// covering k of them hits with probability k/s, so the expected score is
//
//	(k/s) * (n-1) * round(10^7 / (k * (n-1)))
//
// Because of the per-guess rounding, the expected score is NOT flat in k:
// some k get a "free" round-up (e.g. k=94 gives 9 points per hit instead
// of the flat 8), so we pick k by maximizing that expression exactly.
const (
	expectedGuesses = 12499 // n-1 for the tester's data sets
	warmup          = 50    // guesses answered with a safe wide range
	outlierMult     = 8.0   // reject values beyond this many MADs from the median
	madEvery        = 64    // recompute the MAD every this many inserts
)

// stats keeps every inlier value in a sorted slice so the median,
// the MAD and the support bounds are cheap to query.
type stats struct {
	sorted []float64
	mad    float64 // cached median absolute deviation
	dirty  int     // inserts since the MAD was last refreshed
}

func (s *stats) add(v float64) {
	i := sort.SearchFloat64s(s.sorted, v)
	s.sorted = append(s.sorted, 0)
	copy(s.sorted[i+1:], s.sorted[i:])
	s.sorted[i] = v
	s.dirty++
}

func medianOf(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func (s *stats) median() float64 { return medianOf(s.sorted) }

// refreshMAD recomputes the cached median absolute deviation, a spread
// measure that, unlike the standard deviation, is immune to the
// astronomically large outliers present in some data sets.
func (s *stats) refreshMAD() {
	med := s.median()
	devs := make([]float64, len(s.sorted))
	for i, v := range s.sorted {
		devs[i] = math.Abs(v - med)
	}
	sort.Float64s(devs)
	s.mad = medianOf(devs)
	s.dirty = 0
}

func (s *stats) isOutlier(v float64) bool {
	if s.dirty >= madEvery {
		s.refreshMAD()
	}
	return math.Abs(v-s.median()) > outlierMult*math.Max(s.mad, 1)
}

// dropOutliers rebuilds the sorted slice keeping inliers only. Called once
// after the warmup, in case an outlier slipped in before the statistics
// were stable enough to filter.
func (s *stats) dropOutliers() {
	s.refreshMAD()
	med := s.median()
	keep := s.sorted[:0]
	for _, v := range s.sorted {
		if math.Abs(v-med) <= outlierMult*math.Max(s.mad, 1) {
			keep = append(keep, v)
		}
	}
	s.sorted = keep
	s.refreshMAD()
}

// bestCover returns how many values of a support of size s the range
// should cover to maximize the expected score under the tester's
// per-guess rounding.
func bestCover(s int) int {
	bestK, bestScore := s, 0.0
	for k := 1; k <= s; k++ {
		perHit := math.Round(10_000_000 / (float64(k) * expectedGuesses))
		if score := float64(k) / float64(s) * perHit; score > bestScore {
			bestK, bestScore = k, score
		}
	}
	return bestK
}

func main() {
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var s stats
	seen := 0

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		v, err := strconv.ParseFloat(line, 64)
		if err != nil {
			continue
		}
		seen++

		if seen <= warmup {
			s.add(v)
			if seen == warmup {
				s.dropOutliers()
			}
			// Not enough data for stable bounds yet: answer with a
			// generous range around the median.
			med := int(math.Round(s.median()))
			fmt.Fprintf(out, "%d %d\n", med-50, med+50)
			continue
		}

		if !s.isOutlier(v) {
			s.add(v)
		}

		// The inliers are uniformly spread, so their min/max estimate the
		// support precisely; cover the best k of its s integers, centered.
		min := int(math.Round(s.sorted[0]))
		max := int(math.Round(s.sorted[len(s.sorted)-1]))
		support := max - min + 1
		k := bestCover(support)
		lo := min + (support-k)/2
		fmt.Fprintf(out, "%d %d\n", lo, lo+k-1)
	}
}
