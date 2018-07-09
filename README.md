gopgvenv
========

[![travis-ci build status](https://travis-ci.org/fcostin/gopgvenv.svg?branch=master)](https://travis-ci.org/fcostin/gopgvenv)
[![appveyor build status](https://ci.appveyor.com/api/projects/status/leyema0c7yowbken/branch/master?svg=true)](https://ci.appveyor.com/project/fcostin/gopgvenv/branch/master)
[![go report card](https://goreportcard.com/badge/github.com/fcostin/gopgvenv)](https://goreportcard.com/report/github.com/fcostin/gopgvenv)


`gopgvenv` is a command-line tool to provision disposable PostgreSQL DB environments.

Usage (Linux)
-------------

```
someuser@somebox:~/somewhere$ ./gopgvenv "psql --dbname=\$PGURL -c 'select now();'"
pgBinDir= /usr/lib/postgresql/9.5/bin
workDir= /tmp/gopgvenv_976354706
pgDataDir= /tmp/gopgvenv_976354706/pgdata
pgSockDir= /tmp/gopgvenv_976354706/pgsock
port= 42600
waiting for server to start....LOG:  database system was shut down at 2018-07-09 23:45:27 AEST
LOG:  MultiXact member wraparound protections are now enabled
LOG:  autovacuum launcher started
LOG:  database system is ready to accept connections
 done
server started
              now
-------------------------------
 2018-07-09 23:45:29.477944+10
(1 row)

LOG:  received fast shutdown request
LOG:  aborting any active transactions
LOG:  autovacuum launcher shutting down
LOG:  shutting down
LOG:  database system is shut down
```

Features
--------

*   Static binary [YES]
*   Linux support [YES]
*   Windows support [YES]

License
-------

BSD

Prior Art
---------

*   [debian pg_virtualenv](https://manpages.debian.org/jessie/postgresql-common/pg_virtualenv.1.en.html)
*   [jd/pifpaf](https://github.com/jd/pifpaf)

