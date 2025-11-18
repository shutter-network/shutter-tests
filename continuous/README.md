# Continuous tests

This contains a test script, that sends special crafted transactions on each shutterized block. The goal is to have identifiable transactions to allow analysis on inclusion performance. In order to run it:

```
# build `main.go`

cd nethermind-tests; mkdir -p bin && go build -o ./bin/main .
```

```
# ensure the following environment variables and proper values:

# Websocket for "ethereum" (gnosis, ...) rpc
export CONTINUOUS_TEST_RPC_URL=wss://... 
# Private key hex (without 0x prefix) that has enough funding to run the tests
export CONTINUOUS_TEST_PK=
# Contract address (with 0x prefix) for shutter key broadcast contract
export CONTINUOUS_KEY_BROADCAST_CONTRACT_ADDRESS=
# Contract address (with 0x prefix) for shutter keyper set manager contract
export CONTINUOUS_KEYPER_SET_CONTRACT_ADDRESS=
# Contract address (with 0x prefix) for shutter sequencer contract
export CONTINUOUS_SEQUENCER_ADDRESS=
# db user for 'observer' db
export CONTINUOUS_DB_USER=postgres
# db password for 'observer' db
export CONTINUOUS_DB_PASS=test
# db address for 'observer' db
export CONTINUOUS_DB_ADDRESS=localhost:5432
# db name for 'observer' db
export CONTINUOUS_DB_NAME=shutter_metrics
# path to private key file (this is where the testing framework will store additional test accounts - you should back up this file regularily)
export CONTINUOUS_PK_FILE=/home/konrad/Projects/nethermind-tests/pk.hex
# where to store analysis files
export CONTINUOUS_BLAME_FOLDER="/tmp/blame"
# (optional) particular validator indices to monitor (comma-separated list)
export CONTINUOUS_VALIDATOR_INDICES=
```

Make sure, there is an [observer](https://github.com/shutter-network/observer) running and its database accessible as defined in the environment above.

Then you can run the test:
```
./bin/main continuous
```


This will regularily write analysis "blamefiles" to the configured location.

If you need to, you can do the analysis retroactively, by defining a block range and running:
```
./bin/main collect $start-block $end-block
```
