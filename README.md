Fonero
=========

This repository contains `fnod`, a Fonero full node implementation
written in Go (Golang).

It acts as a chain daemon for the [Fonero](https://fonero.org)
cryptocurrency. `fnod` maintains the entire past transactional ledger
of Fonero and allows relaying of transactions to other Fonero
nodes across the world. To read more about Fonero please see the
[project documentation](/doc/trunk/docs/overview.md).

To send or receive funds and join Proof-of-Stake mining, you will also
need to use a `fnowallet`.

Requirements
------------

[Go](http://golang.org) 1.11 or newer.

Installation
------------

```
go get -u -v github.com/fonero-project/fnod/...
```


Testing
-------

`$ make test`

(Runs `./run_tests.sh`.)

Issue Tracker
-------------

The [integrated issue tracker](/ticket) is used for this project.

Documentation
-------------

The documentation is a work-in-progress. It is located in the
[docs](/dir?ci=trunk&name=docs) folder. See for example:

-   [Development notes](docs/development_notes.md)
-   [Fonero updater](docs/updater.md)
-   [Proof-of-Work mechanism](docs/proof_of_work.md)

License
-------

Fonero is licensed under the [copyfree](http://copyfree.org) ISC
License.
