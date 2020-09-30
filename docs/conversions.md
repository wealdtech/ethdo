# Conversions

Converting from mnemonics to keys can be confusing.  Below are commands that allow you to convert from one thing to another.

## I have a mnemonic

A seed is a 24-word phrase that is used as the start point of a process called hierarchical derivation.  It can be used, in combination with a path, to generate any number of keys.

### I want to be able to create keys from the mnemonic

The first thing you need to do is to create a wallet.  To do this run the command below with the following changes:

  - put your mnemonic in place of the sample `abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon   abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art`
  - pick a suitable passphrase for your wallet in place of `secret`.  You will need to use this in all subsequent commands
  - rename the wallet to something other than `Wallet` if you so desire.  If so, you will need to change it in all subsequent commands

```
$ ethdo wallet create --type=hd --mnemonic='abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art' --wallet=Wallet --walletpassphrase=secret
```

### I want a specific public key.

To create a specific public key you need to have both the mnemonic and the derivation path.  A derivation path looks something like `m/12381/3600/0/0` and is used by `ethdo` to generate a specific private key (from which the public key is in turn derived).

You should first create a wallet as per the previous step.  To then create an account run the command below with the following changes:

  - rename the account to something other than `Account` if you so desire.  If so, you will need to change it in all subsequent commands.
  - put your path in place of `m/12381/3600/0/0`
  - pick a suitable passphrase for your account in place of `secret2`.  you will need to use this in all subsequent commands

```
$ ethdo account create --account=Wallet/Account --path=m/12381/3600/0/0 --walletpassphrase=secret --passphrase=secret2
```

At this point you should be able to view your account info with:

```
$ ethdo account info --account=Wallet/Account
Public key: 0x8f5758ff144be9b09d02858522887fccd8dd2a0404ec43439b0f6978909c2128491951486dcaee0f5794262e46f76738
Path: m/12381/3600/0/0
```

This process can be repated for any number of paths by changing the `path` and providing a different account name each time.

### I want the private key.

To obtain the private key of an account follow the steps above, then run:

```
$ ethdo account key --account=Wallet/Account --passphrase=secret2
0x20e3bf019224fec58b27a9774891704646005d6da7a45d4f35032c1c9b501296
```

### I want the _i_th withdrawal and validator keys

[EIP-2334](https://eips.ethereum.org/EIPS/eip-2334) defines derivation path indices for withdrawal and validator keys.  For a given index _i_ the keys will be at the following paths:

  - withdrawal key: m/12381/3600/_i_/0
  - validator key: m/12381/3600/_i_/0/0

Note that the first index is 0, the second is 1, _etc._

To recreate these as `ethdo` accounts run these commands with the following changes:

  - put your index in place of `_i_`
  - pick suitable passphrases for your account in place of `secret2`

```
$ ethdo account create --account=Wallet/Withdrawal_i_ --path=m/12381/3600/_i_/0 --walletpassphrase=secret --passphrase=secret2
$ ethdo account create --account=Wallet/Validator_i_ --path=m/12381/3600/_i_/0/0 --walletpassphrase=secret --passphrase=secret2
```

### I want to recreate the deposit data for my _i_th validator

To recreate the deposit data for a given validator you can run this command, using whatever changes to the default information you may have carried out previously:

```
$ ethdo validator depositdata --withdrawalaccount=Wallet/Withdrawal_i_ --validatoraccount=Wallet/Validator_i_ --depositvalue=32Ether --passphrase=secret2
```

If you wish to be able to provide this information to the launchpad you can add `--launchpad` to the end of the command.

