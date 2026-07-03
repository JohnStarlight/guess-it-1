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

Every provided data set follows the same model: a straight line (slope 0 for
Data 1–3, slope 1 for Data 4–5) plus independent noise, mostly uniform over
~101 integers, polluted with up to ~2 % astronomically large outliers
(±10⁹ in Data 2). The program:

1. **Estimates the trend** with a robust long-baseline Theil–Sen slope (the
   median of the slopes of pairs half the data apart), immune to outliers.
2. **De-trends and filters outliers** using the median and the MAD (median
   absolute deviation) of the residuals, which unlike the mean/standard
   deviation are not wrecked by the huge outliers.
3. **Chooses the guessing window** over the residual distribution by
   maximizing the exact expected score: a window covering `k` integers earns
   `round(10^7 / (k·(n-1)))` per hit, and because of that rounding some
   widths are strictly better than others (covering 94 of the 101 support
   values earns 9 points per hit instead of the flat 8). The residual
   histogram is smoothed and the candidate counts are penalized by their
   sampling noise so the optimizer does not chase lucky bins.

This scores ≈ 103 000–106 000 (~90 % correct guesses) on every data file of
Data 1–5, which beats every provided AI guesser on expectation.

## Testing with the provided tester

```console
cd guess-it-dockerized
docker compose up -d
```

The compose file mounts `../student` into the container, so the tester always
runs the current version of the program. Then open the tester with the AI
guesser to compete against in the URL:

```
http://localhost:3000/?guesser=average
http://localhost:3000/?guesser=big-range
http://localhost:3000/?guesser=correlation-coef
http://localhost:3000/?guesser=huge-range
http://localhost:3000/?guesser=linear-regr
http://localhost:3000/?guesser=median
http://localhost:3000/?guesser=mse
http://localhost:3000/?guesser=nic
```

and click a `Test Data` button. The first run compiles the Go program and
takes a few extra seconds; subsequent runs are instant.

Stop the tester with:

```console
docker compose down
```
