#  RPC Transaction Sender

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
MODE="chiado,gnosis"
CHIADO_SEND_INTERVAL="60"
GNOSIS_SEND_INTERVAL="600"
```

3. Build and run the application:
    ```sh
   docker-compose up --build -d
    ```