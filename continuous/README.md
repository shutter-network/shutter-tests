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

# Continuous Graffiti Mode

The continuous-graffiti mode is a specialized variant of the continuous test mode that targets specific validators based on their graffiti. Instead of sending transactions for every shutterized block, this mode:

1. Queries the observer database to identify the next upcoming slot that will be proposed by a validator whose graffiti matches one of the configured graffiti values
2. Sends encrypted transactions to the sequencer contract, targeting them to be included in that specific validator's slot
3. Monitors transaction inclusion to analyze performance for validators with matching graffiti

This allows for targeted testing and analysis of specific validator sets or operators identified by their graffiti strings.

## Running Continuous Graffiti Mode

To run the continuous-graffiti mode:

```
./bin/main continuous-graffiti
```

## Environment Variables

In addition to all the environment variables required for standard continuous mode (listed above), continuous-graffiti mode requires one additional variable:

```
# Path to JSON file containing list of graffiti strings to match
export GRAFFITI_FILE_PATH=/path/to/graffitis.json
```

### Graffiti File Format

The graffiti file should be a JSON file with the following structure:

```json
{
  "graffitis": [
    "Erigon-Caplin-GSH-C1",
    "gateway.fm",
    "Twinstake",
    "gnos"
  ]
}
```

The file should contain an array of graffiti strings that you want to target. The system will query the observer database to find validators whose graffiti matches any of these strings, and send transactions to be included in their slots.

See `graffitis_example.json` in the project root for a reference example.

