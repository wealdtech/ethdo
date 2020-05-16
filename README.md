[![Tag](https://img.shields.io/github/tag/wealdtech/ethdo.svg)](https://github.com/wealdtech/ethdo/releases/)
[![License](https://img.shields.io/github/license/wealdtech/ethdo.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/ethdo?status.svg)](https://godoc.org/github.com/wealdtech/ethdo)
[![Travis CI](https://img.shields.io/travis/wealdtech/ethdo.svg)](https://travis-ci.org/wealdtech/ethdo)

A command-line tool for managing common tasks in Ethereum 2.

** Please note that this library uses standards that are not yet final, and as such may result in changes that alter public and private keys.  Do not use this library for production use just yet **

## Table of Contents

- [Install](#install)
  - [Docker](#docker)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`ethdo` is a standard Go program which can be installed with:

```sh
GO111MODULE=on go get github.com/wealdtech/ethdo
```

Note that `ethdo` requires at least version 1.13 of go to operate.  The version of go can be found with `go version`.

If this does not work please see the [troubleshooting](https://github.com/wealdtech/ethdo/blob/master/docs/troubleshooting.md) page.

### Docker

It is possible to build the tool using docker:

```sh
docker build -t ethdo .
```

You can run the tool using docker after that. Example:

```sh
docker run -it ethdo --help
```

## Usage

ethdo contains a large number of features that are useful for day-to-day interactions with the Ethereum 2 blockchain.

### Wallets and accounts

ethdo uses the [go-eth2-wallet](https://github.com/wealdtech/go-eth2-wallet) system to provide unified access to different wallet types.  When on the filesystem the locations of the created wallets and accounts are:

    - for Linux: $HOME/.config/ethereum2/wallets
    - for OSX: $HOME/Library/Application Support/ethereum2/wallets
    - for Windows: %APPDATA%\ethereum2\wallets

If using the filesystem store, the additional parameter `basedir` can be supplied to change this location.

All ethdo comands take the following parameters:

  - `store`: the name of the storage system for wallets.  This can be one of "filesystem" (for local storage of the wallet) or "s3" (for remote storage of the wallet on [Amazon's S3](https://aws.amazon.com/s3/) storage system), and defaults to "filesystem"
  - `storepassphrase`: the passphrase for the store.  If this is empty the store is unencrypted
  - `walletpassphrase`: the passphrase for the wallet.  This is required for some wallet-centric operations such as creating new accounts
  - `accountpassphrase`: the passphrase for the account.  This is required for some account-centric operations such as signing data

Accounts are specified in the standard "<wallet>/<account>" format, for example the account "savings" in the wallet "primary" would be referenced as "primary/savings".

### Configuration file and environment

ethdo supports a configuration file; by default in the user's home directory but changeable with the `--config` argument on the command line.  The configuration file provides values that override the defaults but themselves can be overridden with command-line arguments.

The default file name is `.ethdo.json` or `.ethdo.yml` depending on the encoding used (JSON or YAML, respectively).  An example `.ethdo.json` file is shown below:

```json
{
  "store": "s3",
  "storepassphrase": "s3 secret passphrse",
  "account": "Personal wallet/Operations",
  "verbose": true
}
```

ethdo also supports environment variables.  Environment variables are prefixed with "ETHDO_" and are upper-cased.  So for example to provide your account passphrase in an environment variable on a Unix system you could use:

```sh
export ETHDO_PASSPHRASE="my account passphrase"
```

### Output and exit status

If set, the `--quiet` argument will suppress all output.

If set, the `--verbose` argument will output additional information related to the command.  Details of the additional information is command-specific and explained in the command help below.

If set, the `--debug` argument will output additional information about the operation of ethdo as it carries out its work.

Commands will have an exit status of 0 on success and 1 on failure.  The specific definition of success is specified in the help for each command.

# Commands

Command information, along with sample outputs and optional arguments, is available in [the usage section](https://github.com/wealdtech/ethdo/blob/master/docs/usage.md).

# HOWTO

There is a [HOWTO](https://github.com/wealdtech/ethdo/blob/master/docs/howto.md) that covers details about how to carry out various common tasks.

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/ethdo/issues).

## License

[Apache-2.0](LICENSE) © 2019, 2020 Weald Technology Trading Ltd

