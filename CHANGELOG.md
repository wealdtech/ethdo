1.35.2:
  - update dependencies

1.35.1:
  - fix output for various commands that may encounter an empty slot

1.35.0:
  - support Deneb
  - add start and end dates for eth1votes period

1.34.1:
  - fix period parsing for "synccommittee members" command

1.34.0:
  - update dependencies
  - use Capella fork for all exits
  - support Deneb beta 5

1.33.2:
  - fix windows build

1.33.1:
  - add "slot" to "proposer duties" command
  - add activation epoch and time to "validator info" command where applicable
  - add "holesky" to the list of supported networks
  - avoid crash when requesting validators from beacon node without debug enabled

1.33.0:
  - show all slots with 'synccommittee inclusion'
  - add "wallet batch" command

1.32.0:
  - fix incorrect error when "deposit verify" is not given a withdrawal address
  - allow truncated mnemonics (first four characters of each word)
  - add deneb information to "block info"
  - add epoch parameter to "validator yield"
  - add proposer index to "block info"
  - "block info" honours "--quiet" flag
  - "block info" accepts "--block-time" option
  - increase default operation timeout from 10s to 30s
  - "epoch summary" JSON lists number of blobs

1.31.0:
  - initial support for deneb
  - add "--generate-keystore" option for "account derive"
  - update "validator exit" command to be able to generate multiple exits
  - support for 12-word and 18-word mnemonics with single-word (no whitespace) passphrases
  - add JSON output for "validator expectation"

1.30.0:
  - add "chain spec" command
  - add "validator withdrawal" command

1.29.2:
  - fix regression where validator index could not be used as an account specifier

1.29.0:
  - allow use of keystores with validator credentials set
  - tidy up various command options to provide more standard usage
  - add mainnet fallback beacon node

1.28.4:
  - allow validator exit to use a keystore as its validator parameter

1.28.2:
  - fix bix stopping validator exit creation by direct validator specification

1.28.1:
  - generate error message if "validator credentials set" process fails to generate any credentials
  - allow import of accounts with null name field in their keystore
  - show text of execution payload extra data if available

1.28.0:
  - support additional mnemonic word list languages
  - increase minimum timeout for commands that fetch all validators to 2 minutes
  - provide better error messages when offline preparation file cannot be read
  - allow creation of all credential change operations related to a private key (thanks to @joaocenoura)

1.27.1:
  - fix issue with voluntary exits using incorrect domain (thanks to @0xTylerHolmes)

1.27.0:
  - use new build system
  - support S3 credentials
  - update operation of validator exit to match validator credentials set

1.26.5:
  - provide validator information in "chain status" verbose output

1.26.4:
  - provide details of BLS to execution change operations with verbose block output

1.26.3:
  - provide support for additional S3 store options
  - show error when attempting to delete non-filesystem wallets
  - provide additional support for Capella

1.26.2
  - remove check that requires capella prior to generating validator credentials change operations

1.26.1
  - add ability to generate validator credentials change operations prior to the fork in which they become usable

1.26.0
  - add commands and documentation to set user validator credentials (not usable until capella)

1.25.3
  - add more information to "epoch summary"
  - add "validator summary"

1.25.2:
  - no longer require connection parameter
  - support "block analyze" on bellatrix (thanks @tcrossland)
  - check deposit message root match for verifying deposits (thanks @aaron-alderman)

1.25.0:
  - add "proposer duties"
  - add deposit signature verification to "deposit verify"

1.24.1:
  - fix potential crash when new validators are activated
  - add "sepolia" to the list of supported networks

1.24.0:
  - add "validator yield"

1.23.1:
  - do not fetch future state for chain eth1votes

1.23.0:
  - do not fetch sync committee information for epoch summaries prior to Altair
  - ensure that "attester inclusion" without validator returns appropriate error
  - provide more information in "epoch summary" with verbose flag
  - add "chain eth1votes"

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
