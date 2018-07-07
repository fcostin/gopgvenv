gopgvenv
========

[![travis-ci build status](https://travis-ci.org/fcostin/gopgvenv.svg?branch=master)](https://travis-ci.org/fcostin/gopgvenv)
[![go report card](https://goreportcard.com/badge/github.com/fcostin/gopgvenv)](https://goreportcard.com/report/github.com/fcostin/gopgvenv)


`gopgvenv` is a command-line tool to provision disposable PostgreSQL DB environments.

Usage
-----

```
someuser@somebox:~/somewhere$ gopgvenv sh -c "psql --dbname \$PGURL -c 'select now();'"
pgBinDir= /usr/lib/postgresql/9.5/bin
workDir= /tmp/gopgvenv_802082755
pgDataDir= /tmp/gopgvenv_802082755/pgdata
pgSockDir= /tmp/gopgvenv_802082755/pgsock
port= 33497
postgresOptions= -i -h localhost -p 33497 -k /tmp/gopgvenv_802082755/pgsock -F
              now
-------------------------------
 2018-07-07 21:36:02.418061+10
(1 row)
```

Features
--------

*   Static binary [YES]
*   Linux support [YES]
*   Windows support [TODO; ASPIRATIONAL]

License
-------

BSD

Prior Art
---------

*   [debian pg_virtualenv](https://manpages.debian.org/jessie/postgresql-common/pg_virtualenv.1.en.html)
*   [jd/pifpaf](https://github.com/jd/pifpaf)

