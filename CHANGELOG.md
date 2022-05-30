dev:
  - do not fetch sync committee information for epoch summaries prior to Altair
  - ensure that "attester inclusion" without validator returns appropriate error
  - provide more information in "epoch summary" with verbose flag

1.22.0:
  - add "ropsten" to the list of supported networks

1.21.0:
  - add "validator credentials get"

1.20.0:
  - add "chain queues"

1.19.1:
  - add the ability to import keystores to ethdo wallets
  - use defaults to connect to beacon nodes if no explicit connection defined

1.19.0:
  - add "epoch summary"

1.18.2:
  - tidy up output of "block info"

1.18.1:
  - do not show execution payload if empty

1.18.0:
  - add "-ssz" option to "block info"
  - add "block analyze" command
  - support bellatrix

1.17.0:
  - add sync committee information to "chain time"
  - add details of vote success to "attester inclusion --verbose"
  - add "synccommittee inclusion"

1.15.1:
  - provide sync committee slots in "chain status"
  - clarify that --connection should be a URL

1.15.0:
  - add --period to "synccommittee members", can be "current", "next"
  - add "validator expectation"

1.14.0:
  - add "chain verify signedcontributionandproof"
  - show both block and body root in "block info"
  - add exit / withdrawable epoch to "validator info"

1.13.0:
  - rework and provide additional information to "chain status" output

1.12.0:
  - add "synccommittee members"

1.11.0
  - add Altair information to "block info"
  - add more information to "chain info"

1.10.2
  - use local shamir code (copied from github.com/hashicorp/vault)

1.10.0
  - add "wallet sharedexport" and "wallet sharedimport"

1.9.1
  - Avoid crash when required interfaces for chain status command are not supported
  - Avoid crash with latest version of herumi/go-bls

1.9.0
  - allow use of Ethereum 1 address as withdrawal credentials

1.8.1
  - fix issue where 'attester duties' and 'attester inclusion' could crash

1.8.0
  - add "chain time"
  - add "validator keycheck"

1.7.5:
  - add "slot time"
  - add "attester duties"
  - add "node events"
  - add activation epoch to "validator info"

1.7.3:
  - fix issue where base directory was ignored for wallet creation
  - new "validator duties" command to display known duties for a given validator
  - update go-eth2-client to display correct validator status from prysm

1.7.2:
  - new "account derive" command to derive keys directly from a mnemonic and derivation path
  - add more output to "deposit verify" to explain operation

1.7.1:
  - fix "store not set" issue

1.7.0:
  - "validator depositdata" now defaults to mainnet, does not silently fetch fork version from chain
  - update deposit data output to version 3, to allow for better deposit checking
  - use go-eth2-client for beacon node communications
  - deprecated "--basedir" in favor of "--base-dir"
  - deprecated "--storepassphrase" in favor of "--store-passphrase"
  - deprecated "--walletpassphrsae" in favor of "--wallet-passphrsae"
  - renamed "--exportpassphrase" and "--importpassphrase" flags to "--passphrase"
  - reworked internal structure of account-related commands
  - reject weak passphrases by default

1.6.1:
  - "attester inclusion" defaults to previous epoch
  - output array for launchpad deposit data JSON in all situations

1.6.0:
  - update BLS HKDF function to match spec 04
  - add --launchpad option to "validator depositdata" to output data in launchpad format

1.5.9:
  - fix issue where wallet mnemonics were not normalised to NFKD
  - "block info" supports fetching the gensis block (--slot=0)
  - "attester inclusion" command finds the inclusion slot for a validator's attestation
  - "account info" with verbose option now displays participants for distributed accounts
  - fix issue where distributed account generation without a passphrase was not allowed

1.5.8:
  - allow raw deposit transactions to be supplied to "deposit verify"
  - move functionality of "account withdrawalcredentials" to be part of "account info"
  - add genesis validators root to "chain info"
