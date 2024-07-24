# (Stress) Integration tests for shutter on gnosis chain

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