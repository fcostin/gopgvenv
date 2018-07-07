plan: golang pgvenv
===================

goal
----

small command line tool to provide env with temporary disposable postgres db

distinguishing features:

*	static binary
*	cross platform (windows, linux)

business case
-------------

na. hobby project not product.

license
-------

bsd

hidden conspiracy
-----------------

learn basic cross-platform golang command line tool scriping

Why golang?

*	bored of python
*	$NEXTJOB
*	wget static binary is a deployment love story for dev tools
*	support linux AND windows

dev notes
---------

*	mimic the basic debian `pg_virtualenv` interface or the `pifpaf` interface: `venv options -- subcommand`
*	dev in linux first


### under the hood, do everything by exec-ing:

*	`pg_config --bindir`
*	`pg_ctl`:
*	`pg_isready (if necessary)`

```
pg_ctl initdb [-s] [-D datadir] [-o initdb-options]
pg_ctl start [-w] [-t seconds] [-s] [-D datadir] [-l filename] [-o options] [-p path] [-c]
pg_ctl stop [-W] [-t seconds] [-s] [-D datadir] [-m s[mart] | f[ast] | i[mmediate]]
pg_ctl status [-D datadir]
pg_ctl kill signal_name process_id
```

### windows.

*	can get windows dev VM from microsoft and use appveyor for CI
*	https://github.com/golang/go/blob/master/src/syscall/exec_windows.go

prior art
---------

*	[debian pg_virtualenv](https://manpages.debian.org/jessie/postgresql-common/pg_virtualenv.1.en.html)
*	[jd/pifpaf](https://github.com/jd/pifpaf)

