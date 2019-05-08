fees
=======


[![Build Status](http://img.shields.io/travis/fonero/fnod.svg)](https://travis-ci.org/fonero/fnod)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/fonero-project/fnod/fees)

Package fees provides fonero-specific methods for tracking and estimating fee
rates for new transactions to be mined into the network. Fee rate estimation has
two main goals:

- Ensuring transactions are mined within a target _confirmation range_
  (expressed in blocks);
- Attempting to minimize fees while maintaining be above restriction.

This package was started in order to resolve issue fonero/fnod#1412 and related.
See that issue for discussion of the selected approach.

This package was developed for fnod, a full-node implementation of Fonero which
is under active development.  Although it was primarily written for
fnod, this package has intentionally been designed so it can be used as a
standalone package for any projects needing the functionality provided.

## Installation and Updating

```bash
$ go get -u github.com/fonero-project/fnod/fees
```

## License

Package fnoutil is licensed under the [copyfree](http://copyfree.org) ISC
License.
