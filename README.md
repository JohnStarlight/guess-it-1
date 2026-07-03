# guess-it-1

A program that, given numbers on standard input one by one, prints for each
one the range in which it expects the **next** number to fall.

## Repository layout

```
guess-it-1/
├── student/                 <- the submission (copy this into the tester root)
│   ├── main.go              <- the guessing program (Go)
│   └── script.sh            <- executable entry point, run from the tester root
└── guess-it-dockerized/     <- the provided tester (web UI on port 3000)
```

## How the guesser works

The tester rewards each correct guess with `round(10^7 / (1 + width) / (n - 1))`
points, so the narrower the range, the more each hit is worth.

The graded data sets are ~99 % uniformly distributed integers over
`[100, 200]`, polluted with ~1 % astronomically large outliers. The program:

1. **Filters outliers** using the median and the MAD (median absolute
   deviation), which unlike the mean/standard deviation are immune to the
   huge outliers.
2. **Estimates the support** `[min, max]` of the distribution from the
   filtered values.
3. **Chooses the range width** by maximizing the exact expected score: for a
   uniform distribution over `s` integers, a range covering `k` of them hits
   with probability `k/s` and earns `round(10^7 / (k·(n-1)))` per hit.
   Because of the rounding, some widths are strictly better than others —
   for the tester's data sets the optimum covers 94 of the 101 values
   (9 points per hit at a ~93 % hit rate).

This scores ≈ 104 000–106 000 on every graded data file, which beats every
provided AI guesser on expectation.

## Testing with the provided tester

```console
cd guess-it-dockerized
docker compose up -d
```

The compose file mounts `../student` into the container, so the tester always
runs the current version of the program. Then open
`http://localhost:3000/?guesser=<name>` (e.g. `median`, `linear-regr`, `nic`,
see `guess-it-dockerized/README.md` for the full list) and click a
`Test Data` button. The first run compiles the Go program and takes a few
extra seconds; subsequent runs are instant.

Stop the tester with:

```console
docker compose down
```
