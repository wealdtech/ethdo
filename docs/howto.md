# How to achieve common tasks with ethdo

## Create a new wallet

New wallets can be created with the `ethdo wallet create` command.  Each wallet has to have a unique name, for example:

```sh
ethdo wallet create --wallet="My wallet"
```

Additional options are available to decide the type of wallet and encryption.

## Back up a wallet

A wallet can be backed up with the `ethdo wallet export` command.  This creates an encrypted backup of the wallet, for example:

```sh
ethdo wallet export --wallet="My wallet" --exportpassphrase="export secret" >export.dat
```

Note that by default the wallet backup is printed to the console, hence the `>export.dat` to redirect it to a file.

## Restore a wallet

A backed up wallet can be restored with the `ethdo wallet import` command, for example:

```sh
ethdo wallet import --importdata=export.dat --importpassphrase="export secret"
```

In this example the wallet to be imported is being read from the `export.dat` file.

Note that if a wallet with the same name already exists it cannot be imported.

## Where is my wallet?

Details of the location of a wallet can be found with the `ethdo wallet info` command, for example:

```sh
ethdo wallet info --verbose --wallet="My wallet"
```

This will provide, amongst other information, a `Location` line giving the directory where the wallet information resides.
