# guess-it

## Usage

You will first need copy the `student/` folder (provided by the student) in
which you will see the student guessing program along with a file called
`script.sh`. This file should be an executable shell script that runs the
student program if you are in the root folder `guess-it/`. The filesystem
should look somethings like this:

```console
─ guess-it/
├── ai/
│   ├── big-range
│   └── ...
├── index.html
├── index.js
└── ...
└── student/
    ├── ...
    └── script.sh

```

To test the student program, these commands should be ran to have the
dependencies needed and to start the webpage on the port 3000:

```console
docker compose up
```

To run it in the background and keep your terminal free:

```console
docker compose up -d
```

To stop the server properly when you are done (avoids it getting stuck):

```console
docker compose down
```

If the server is unresponsive and `Ctrl+C` alone does not work:

```console
docker compose down --remove-orphans
```

> Note: the first test run after each `docker compose up` will take a few
> extra seconds while the student program compiles. Subsequent runs are instant.

After opening your browser of preference in the port
[3000](http://localhost:3000/), if you try clicking on any of the `Test Data`
buttons, you will notice that in the Dev Tool/ Console there is a message which
tells you that you need another guesser besides the student.

Adding a guesser is simple. You need to add in the URL a guesser, in other
words, the name of one of the files present in the `ai/` folder:

```console
?guesser=<name_of_guesser>
```

Here are the ready-to-use URLs for each available guesser:

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

After that, choose which of the random data set to test. After that you can
wait for the program to test all of the values (boooooring), or you can click
`Quick` to skip the waiting and be presented with the results.

Since the website uses big data sets, we advise you to clear the displays
clicking on the `Clean` button after each test.
