#  RPC Transaction Sender

This is the README for the original `nethermind tests`. For `stress` and `continuous` tests, see
`continuous/README.md` and `stress/README.md`.

This application is designed to send transactions to chiado, gnosis or both at configurable intervals.

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/shutter-network/nethermind-tests.git
    cd nethermind-tests
    ```

2. Copy the environment variables from `template.env` and adjust them as needed:

```env
PRIVATE_KEY="YOUR_PRIVATE_KEY"
MODE="chiado,gnosis,send-wait"

#CHIADO TEST
CHIADO_URL="https://erpc.chiado.staging.shutter.network"
CHIADO_SEND_INTERVAL="60"

#GNOSIS TEST
GNOSIS_URL="https://erpc.gnosis.shutter.network"
GNOSIS_SEND_INTERVAL="600"

#SEND AND WAIT TEST
NODE_URL="https://erpc.chiado.staging.shutter.network"
WAIT_TX_TIMEOUT=10
TEST_DURATION=1
```

- `MODE=chiado`: Sends transactions at intervals defined by `CHIADO_SEND_INTERVAL` to the Chiado URL.
- `MODE=gnosis`: Sends transactions at intervals defined by `GNOSIS_SEND_INTERVAL` to the Gnosis URL.
- `MODE=send-wait`: 
  - Sends a transaction to the network defined in `NODE_URL` at nonce `n`
  - waits for a timeout defined by `WAIT_TX_TIMEOUT`
  - then sends the next one at nonce `n+ 1`
  - test is run for the duration defined in `TEST_DURATION`

- Multiple tests can be run at the same time by separating the different modes with a comma, i.e. `MODE="chiado,gnosis"`.

3. Build and run the application:
    ```sh
   docker-compose up --build -d
    ```
