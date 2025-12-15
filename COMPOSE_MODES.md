## Continuous modes via Docker Compose

Only one of the two services is meant to run at a time:
- `continuous`: end-to-end continuous testing
- `continuous-graffiti`: graffiti-only mode

Defaults (Gnosis Chain):
- `continuous`: `ENV_FILE=.env`
- `continuous-graffiti`: `ENV_FILE=.mainnet-graffiti.env`
- `DOCKER_NETWORK=keyper-metrics_default`
- `BLAME_DIR=./data/blame`

Chiado runs use `ENV_FILE=chiado.env`, `DOCKER_NETWORK=chiado-observer_default`, and a Chiado-specific `BLAME_DIR=$(pwd)/data/chiado-blame`.

### Run commands
- Gnosis / continuous: `docker compose up continuous`
- Gnosis / continuous-graffiti: `docker compose up continuous-graffiti`
- Chiado / continuous: `ENV_FILE=chiado.env DOCKER_NETWORK=chiado-observer_default BLAME_DIR="$(pwd)/data/chiado-blame" docker compose up continuous`
- Chiado / continuous-graffiti: `ENV_FILE=chiado-graffiti.env DOCKER_NETWORK=chiado-observer_default BLAME_DIR="$(pwd)/data/chiado-graffiti-blame" docker compose up continuous-graffiti`
- Custom blame dir (example mainnet graffiti folder): `BLAME_DIR="$(pwd)/data/graffiti-blame" docker compose up continuous-graffiti`

### Run the same test on chiado and gnosis at the same time (using `-p`)
- Continuous (gnosis): `docker compose -p continuous-gnosis up -d continuous`
- Continuous (chiado): `ENV_FILE=chiado.env DOCKER_NETWORK=chiado-observer_default BLAME_DIR="$(pwd)/data/chiado-blame" docker compose -p continuous-chiado up -d continuous`
- Graffiti (gnosis): `docker compose -p graffiti-gnosis up -d continuous-graffiti`
- Graffiti (chiado): `ENV_FILE=chiado-graffiti.env DOCKER_NETWORK=chiado-observer_default BLAME_DIR="$(pwd)/data/chiado-graffiti-blame" docker compose -p graffiti-chiado up -d continuous-graffiti`
- Stop both continuous stacks: `docker compose -p continuous-gnosis down && docker compose -p continuous-chiado down`
- Stop both graffiti stacks: `docker compose -p graffiti-gnosis down && docker compose -p graffiti-chiado down`

### Notes
- `ENV_FILE` defaults to `.env` for `continuous` and `.mainnet-graffiti.env` for `continuous-graffiti`; override as needed.
- `DOCKER_NETWORK` defaults to `keyper-metrics_default`; set it when you need `chiado-observer_default` (or any other external network that exists locally).
- `BLAME_DIR` defaults to `./data/blame` for `continuous` and `./data/graffiti-mainnet-blame` for `continuous-graffiti`; set it per-chain (e.g., `$(pwd)/data/graffiti-chiado-blame`) to keep outputs separated.
