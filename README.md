goxxx  [![Build Status](https://travis-ci.org/vaz-ar/goxxx.svg)](https://travis-ci.org/vaz-ar/goxxx)
=====

IRC bot written in Go.

Install
=======

Once you have a working installation of Go (Go version >= 1.5), you just need to run:

```
$ go get github.com/vaz-ar/goxxx
```

Build
=======

Under `$GOPATH/src/github.com/vaz-ar/goxxx`, run:

```
$ make
```

Usage
=====

Minimal command to start goxxx is the following:

```
$ goxxx -channel "#my_channel" -config config.ini
```

Where `config.ini` is a configuration file (see `sample_config.ini` to write your own file).

To get help about program usage, just run:
```
$ goxxx
```

### Configuration file
- By default goxxx will search for a file named `goxxx.ini` in the directory where it is started.
- You can also specify a path for the configuration file via the `-config` flag.

### Log file
- The log file will be created in the directory where goxxx is started, and will be named `goxxx_logs.txt`.

### Troubles
If command fails silently, search in `goxxx_logs.txt` for similar lines:

```
2016/09/28 17:33:42 unable to open database file
2016/09/28 17:33:42 Error while applying migrations, exiting ...
```

You need to set correct permissions on `./storage/`:

```
$ chmod u+rw storage
```

Commands
=====

Currently implemented commands:

### invoke
- !invoke \<nick\> \[\<message\>\] => Send an email to an user, with an optionnal message

### memo
- !memo/!m \<nick\> \<message\> => Leave a memo for another user
- !memostat/!ms => Get the list of the unread memos (List only the memos you left)

### pictures
- !p/!pic \<search terms\> => Search in the database for pictures matching \<search terms\>
- !addpic \<url\> \<tag\> \[#NSFW\] => Add a picture in the database for \<tag\> (\<url\> must have an image extension)
- !rmpic \<url\> \<tag\> => Remove a picture in the database for \<tag\> (Admin only command)

### quote
- !q/!quote \<nick\> \[\<part of message\>\]
- !aq/!addquote \<nick\> \<part of message\>
- !rmq/!rmquote \<nick\> \<part of the quote\> (Admins only)

### search
- !d/!dg/!ddg \<terms to search\> => Search on DuckduckGo
- !w/!wiki \<terms to search\> => Search on Wikipedia EN
- !wf/!wfr \<terms to search\> => Search on Wikipedia FR
- !u/!ud \<terms to search\> => Search on Urban Dictionnary

### url
- !url \<search terms\>=> Return links with titles matching \<search terms\>

### xkcd
- !xkcd \[\<comic number\>\] => Return the XKCD comic corresponding to the number. If number is not specified, returns the last comic.

Tests
=====

to run the tests:
```
$ make test
```
It will run the tests for all the packages.


Development / Contributions
=====

Pull requests are welcome.
