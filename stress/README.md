# (Stress) Integration tests for shutter on gnosis chain

This contains `go test`-runnable testcases to stress- and edge case-test a shutter on gnosis chain live system, in an end-to-end fashion. It does however not use the `encrypting-rpc-server`, but instead submits transactions directly to the sequencer contract.

As a pre-requisite, an account with sufficient `"ETH"` (or other gas tokens corresponding to the chain the system is deployed on) is required. Its private key needs to be available hex encoded in an environment variable `STRESS_TEST_PK=caffee…`.

The most notable test cases are
- `TestStressManyNoWait`, which sends a large number of shutterized transfers at once, in order to test the limits of decryption key generation and propagation on the `keyper` side, as well as decryption performance on the validator side.
- `TestInception`, which sends shutterized transactions inside shutterized transactions.
- `TestStressExceedEncryptedGasLimit`, which tests that there is a limit to the gas that can be used in encrypted transactions per block.

Most other test cases are very simple, and could be considered basic functionality "smoke tests".

**Note**: Due to the live nature of the test environment, not all `PASS`ing tests can show the absence of errors. For example:

`TestStressExceedEncryptedGasLimit`
1) The test relies on a hardcoded value for the limit. If the parameters of the live system change, that value may be outdated.
2) The test shall `PASS` if not all transactions end in the same shutterized block. It can however happen, that an external entity also submitted transactions and therefore the assumed correct behavior happened only by chance.

Also: Since we have no insight to the validator, we rely on timeouts for waiting for transactions to be included on-chain. Additionally, we can not reliably check for failing transactions.

In short: this PR is to some extent ab-using the `go test` environment to allow us excecuting these tests.

There are also two special "Test…" functions, that do not test anything, but can be used for house keeping:
- `TestEmptyAccounts`, which allows to drain all funds from extra accounts that got generated during previous test runs.
- `TestFixNonce`, which can help to clear out nonce-gaps from the public mem-pool for the main test account.

## Setup

To run the tests, edit the values in `envrc_sample` and make sure the environment variables are exported.
This works, e.g. by sourcing the edited file `source envrc_sample` or, if you are using `direnv`, by copying the
file to `.envrc` and running `direnv allow`.

## Running single tests

Navigate to this directory and run `go test -run $NAME_FRAGMENT`, where `$NAME_FRAGMENT` will be matched from the existing
test names. E.g. `go test -run Single` will evaluate to run `TestStressSingle`.

## Reclaiming funds

Most tests will fund some accounts from the primary test key account (as defined in `STRESS_TEST_PK`). In order to allow for 
recovery of the used funds, all created accounts will be stored in a file called `pk.hex`.

You can periodically run `go test -run Account` to reclaim all funds to the primary test account. If this went according to your expectations,
you can remove the backup file afterwards (`rm pk.hex`). 