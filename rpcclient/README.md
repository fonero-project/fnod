rpcclient
=========

[![Build Status](http://img.shields.io/travis/fonero/fnod.svg)](https://travis-ci.org/fonero/fnod)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/fonero-project/fnod/rpcclient)

rpcclient implements a Websocket-enabled Fonero JSON-RPC client package written
in [Go](http://golang.org/).  It provides a robust and easy to use client for
interfacing with a Fonero RPC server that uses a fnod compatible Fonero
JSON-RPC API.

## Status

This package is currently under active development.  It is already stable and
the infrastructure is complete.  However, there are still several RPCs left to
implement and the API is not stable yet.

## Documentation

* [API Reference](http://godoc.org/github.com/fonero-project/fnod/rpcclient)
* [fnod Websockets Example](https://github.com/fonero-project/fnod/tree/master/rpcclient/examples/fnodwebsockets)
  Connects to a fnod RPC server using TLS-secured websockets, registers for
  block connected and block disconnected notifications, and gets the current
  block count
* [fnowallet Websockets Example](https://github.com/fonero-project/fnod/tree/master/rpcclient/examples/fnowalletwebsockets)  
  Connects to a fnowallet RPC server using TLS-secured websockets, registers for
  notifications about changes to account balances, and gets a list of unspent
  transaction outputs (utxos) the wallet can sign

## Major Features

* Supports Websockets (fnod/fnowallet) and HTTP POST mode (bitcoin core-like)
* Provides callback and registration functions for fnod/fnowallet notifications
* Supports fnod extensions
* Translates to and from higher-level and easier to use Go types
* Offers a synchronous (blocking) and asynchronous API
* When running in Websockets mode (the default):
  * Automatic reconnect handling (can be disabled)
  * Outstanding commands are automatically reissued
  * Registered notifications are automatically reregistered
  * Back-off support on reconnect attempts

## Installation

```bash
$ go get -u github.com/fonero-project/fnod/rpcclient
```

## License

Package rpcclient is licensed under the [copyfree](http://copyfree.org) ISC
License.
