goxxx  [![Build Status](https://travis-ci.org/vaz-ar/goxxx.svg)](https://travis-ci.org/vaz-ar/goxxx)
=====

IRC bot written in Go.

Install
=======

Once you have a working installation of Go, you just need to run:

```
$ go get github.com/vaz-ar/goxxx
```

Build
=======
```
$ make
```

Usage
=====

To get help about program usage, just run:
```
$ goxxx
```

### Configuration file
- By default goxxx will search for a file named `goxxx.ini` in the directory where it is started.
- You can also specify a path for the configuration file via the `-config` flag.

### Log file
- The log file will be created in the directory where goxxx is started, and will be named `goxxx_logs.txt`.


Commands
=====

Currently implemented commands:

### memo
- !memo/!m \<nick\> \<message\> => Leave a memo for another user
- !memostat/!ms => Get the list of the unread memos (List only the memos you left)

### pictures
- !p/!pic \<search terms\> => Search in the database for pictures matching \<search terms\>
- !addpic \<url\> \<tag\> \[#NSFW\] => Add a picture in the database for \<tag\> (\<url\> must have an image extension)
- !rmpic \<url\> \<tag\> => Remove a picture in the database for \<tag\> (Admin only command)

### url
- !url \<search terms\>=> Return links with titles matching \<search terms\>

### search
- !d/!dg/!ddg \<terms to search\> => Search on DuckduckGo
- !w/!wiki \<terms to search\> => Search on Wikipedia EN
- !wf/!wfr \<terms to search\> => Search on Wikipedia FR
- !u/!ud \<terms to search\> => Search on Urban Dictionnary

### invoke
- !invoke \<nick\> \[\<message\>\] => Send an email to an user, with an optionnal message

### xkcd
- !xkcd \[\<comic number\>\] => Return the XKCD comic corresponding to the number. If number is not specified, returns the last comic.

### quote
- !q/!quote \<nick\> \[\<part of message\>\] (If \<part of message\> is not supplied, it will list the quotes for \<nick\>)
- !rmq/!rmquote \<nick\> \<part of the quote\> (Admins only)



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
