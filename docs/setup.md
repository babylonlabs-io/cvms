# Service Guide Document

## Medium

- [[EN] Efficiently Monitor Cosmos-SDK based Blockchain Nodes with Cosmos Validator Monitoring Service (CVMS) 🌌 🔭](https://medium.com/cosmostation/efficiently-monitor-cosmos-sdk-nodes-with-cosmos-validator-monitoring-service-cvms-e24a63c73462?source=collection_category---4------0-----------------------)

* [[KR] Cosmos Validator Monitoring Service (CVMS) 🌌 🔭 를 통한 Cosmos-SDK 기반 블록체인 노드를 효과적으로 모니터링하는 방법](https://medium.com/cosmostation/cosmos-validator-monitoring-service-cvms-%EB%A5%BC-%ED%86%B5%ED%95%9C-cosmos-sdk-%EA%B8%B0%EB%B0%98-%EB%B8%94%EB%A1%9D%EC%B2%B4%EC%9D%B8-%EB%85%B8%EB%93%9C%EB%A5%BC-%ED%9A%A8%EA%B3%BC%EC%A0%81%EC%9C%BC%EB%A1%9C-%EB%AA%A8%EB%8B%88%ED%84%B0%EB%A7%81%ED%95%98%EB%8A%94-%EB%B0%A9%EB%B2%95-8b5541bcb834)

## Example: Setup

Copy the `.env` file if you need to customize service ports, log-level, prometheus.yaml or other configurations.

### Setup

```bash
# clone
git clone https://github.com/cosmostation/cvms.git && cd cvms

# Create a config file from example config
cp .resource/example-validator-config.yaml config.yaml

# Skip copy .env file
# cp .resource/.env.example .env

# Modify the config file for your validator
vi config.yaml

# Run cvms
docker compose up --build -d
```

**Example config.yaml for validator mode**

```yaml
# NOTE: Customize this variables by your needs
# 1. network mode:
#   ex) monikers: ['all']
#   des) This will enable network mode to monitor all validators status in the blockchain network
#
# 2. validator mode:
#   ex) monikers: ['Cosmostation1', 'Cosmostation2']
#   des) This will enable validator mode for whitelisted specific validators
monikers: ['Cosmostation']

# if user is one of validators, they want to operator for whole chains which already operating as validator.
chains:
  # NOTE: display name will be used only this config to indicate followed arguments to communicate internal team members
  - display_name: 'cosmos'
    # NOTE: chain_id is a key for support_chains list. YOU SHOULD match correct CHAIN ID
    chain_id: cosmoshub-4
    # NOTE: these addresses will be used for balance usage tracking such as validator, broadcaster or something.
    tracking_addresses:
      - 'cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep4tgu9q'
    nodes:
      # NOTE: currently grpc endpoint doesn't support ssl
      - rpc: 'https://rpc-cosmos.endpoint.xyz'
        api: 'https://lcd-cosmos.endpoint.xyz'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'

  - display_name: 'injective'
    chain_id: injective-1
    tracking_addresses:
      - 'inj1rvqzf9u2uxttmshn302anlknfgsatrh5mcu6la'
      - 'inj1mtxhcchfyvvs6u4nmnylgkxvkrax7c2la69l8w' # eventnonce orchestrator address or something
    nodes:
      - rpc: 'https://rpc-injective.endpoint.xyz'
        api: 'https://lcd-injective.endpoint.xyz'
        grpc: 'grpc-injective.endpoint.xyz:9090'

  - display_name: 'sei'
    chain_id: pacific-1
    tracking_addresses: []
    nodes:
      - rpc: 'https://rpc-sei.endpoint.xyz'
        api: 'https://lcd-sei.endpoint.xyz'
        grpc: 'grpc-sei.endpoint.xyz:9090'

  - display_name: dydx
    chain_id: dydx-mainnet-1
    tracking_addresses: []
    nodes:
      - rpc: 'https://rpc-dydx.endpoint.xyz'
        api: 'https://lcd-dydx.endpoint.xyz'
        grpc: 'grpc-dydx.endpoint.xyz:9090'

  - display_name: neutron
    chain_id: neutron-1
    tracking_addresses: []
    nodes:
      - rpc: 'https://rpc-neutron.endpoint.xyz'
        api: 'https://lcd-neutron.endpoint.xyz'
        grpc: 'grpc-neutron.endpoint.xyz:9090'
    provider_nodes:
      - rpc: 'https://rpc-cosmos.endpoint.xyz'
        api: 'https://lcd-cosmos.endpoint.xyz'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'

  - display_name: axelar
    chain_id: axelar-dojo-1
    tracking_addresses:
      - axelar146kdz9stlycvacm03hg0t5fxq6jszlc4gtxgpr
    nodes:
      - rpc: 'https://rpc-axelar.endpoint.xyz'
        api: 'https://lcd-axelar.endpoint.xyz'
        grpc: 'grpc-axelar.endpoint.xyz:9090'
```

**Example config.yaml for network mode**

```yaml
# NOTE: Customize this variables by your needs
# 1. network mode:
#   ex) monikers: ['all']
#   des) This will enable network mode to monitor all validators status in the blockchain network
#
# 2. validator mode:
#   ex) monikers: ['Cosmostation1', 'Cosmostation2']
#   des) This will enable validator mode for whitelisted specific validators
monikers: ['all']

# if user is one of validators, they want to operator for whole chains which already operating as validator.
chains:
  # NOTE: display name will be used only this config to indicate followed arguments to communicate internal team members
  - display_name: 'cosmos'
    # NOTE: chain_id is a key for support_chains list. YOU SHOULD match correct CHAIN ID
    chain_id: cosmoshub-4
    # NOTE: these addresses will be used for balance usage tracking such as validator, broadcaster or something.
    tracking_addresses:
      - 'cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep4tgu9q'
    nodes:
      # NOTE: currently grpc endpoint doesn't support ssl
      - rpc: 'https://rpc-cosmos.endpoint.xyz'
        api: 'https://lcd-cosmos.endpoint.xyz'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'

  - display_name: 'injective'
    chain_id: injective-1
    tracking_addresses:
      - 'inj1rvqzf9u2uxttmshn302anlknfgsatrh5mcu6la'
      - 'inj1mtxhcchfyvvs6u4nmnylgkxvkrax7c2la69l8w' # eventnonce orchestrator address or something
    nodes:
      - rpc: 'https://rpc-injective.endpoint.xyz'
        api: 'https://lcd-injective.endpoint.xyz'
        grpc: 'grpc-injective.endpoint.xyz:9090'

  - display_name: 'sei'
    chain_id: pacific-1
    tracking_addresses: []
    nodes:
      - rpc: 'https://rpc-sei.endpoint.xyz'
        api: 'https://lcd-sei.endpoint.xyz'
        grpc: 'grpc-sei.endpoint.xyz:9090'

  - display_name: dydx
    chain_id: dydx-mainnet-1
    tracking_addresses: []
    nodes:
      - rpc: 'https://rpc-dydx.endpoint.xyz'
        api: 'https://lcd-dydx.endpoint.xyz'
        grpc: 'grpc-dydx.endpoint.xyz:9090'

  - display_name: neutron
    chain_id: neutron-1
    tracking_addresses: []
    nodes:
      - rpc: 'https://rpc-neutron.endpoint.xyz'
        api: 'https://lcd-neutron.endpoint.xyz'
        grpc: 'grpc-neutron.endpoint.xyz:9090'
    provider_nodes:
      - rpc: 'https://rpc-cosmos.endpoint.xyz'
        api: 'https://lcd-cosmos.endpoint.xyz'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'

  - display_name: axelar
    chain_id: axelar-dojo-1
    tracking_addresses:
      - axelar146kdz9stlycvacm03hg0t5fxq6jszlc4gtxgpr
    nodes:
      - rpc: 'https://rpc-axelar.endpoint.xyz'
        api: 'https://lcd-axelar.endpoint.xyz'
        grpc: 'grpc-axelar.endpoint.xyz:9090'
```

## Example Using Environment Variables

Environment variables can be used to inject values into the configuration file.
For this a simple template variable marked by `${}` (e.g. `${SUPER_SECRET_ENV_VAR}`) can be added to the config file.

```yaml
monikers: ['all']

chains:
  - display_name: 'cosmos'
    chain_id: cosmoshub-4
    tracking_addresses:
      - 'cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep4tgu9q'
      - '${THIS_IS_ALSO_AN_ENV_VAR}'
    nodes:
      - rpc: 'https://rpc-cosmos.endpoint.xyz/${SUPER_SECRET_TOKEN}'
        api: 'https://lcd-cosmos.endpoint.xyz/${ANOTHER_ENV_VAR}'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'
```

## Example: Supporting Custom Chain

For devents, testnets, localnet even if unsupported mainnets, Use `custom_chains.yaml` for CVMS

```bash
# Copy custom_chains.yaml from example
# You should locate it into docker/cvms/custom_chains.yaml
cp docker/cvms/custom_chains.yaml.example docker/cvms/custom_chains.yaml

# And then Mount your custom_chains.yaml into docker compose
# Just uncomment the CUSTOM_CHAINS_FILE
CUSTOM_CHAINS_FILE=custom_chains.yaml

# After init CVMS, you can check current support_chains with custom_chains by using exporter app
curl -X GET http://localhost:9200/support_chains
```

**Example custom_chains.yaml**

```yaml
---
mintstation-1:
  protocol_type: cosmos
  support_asset:
    denom: umint
    decimal: 6
  packages:
    - block
    - upgrade
    - uptime
    - voteindexer #for cometbft consensus vote
    - veindexer # for vote-extension
```

## Example: Setup for Persistent Mode

> **Note:** Persistent mode means CVMS uses an external PostgreSQL service to store index data permanently.  
> If you want to retain historical records, please follow the steps in this guide.

You can copy and customize the `.env` file to configure service ports, log levels, `prometheus.yaml`, or other settings as needed.

---

### Setup

```bash
# Clone the repository
git clone https://github.com/cosmostation/cvms.git && cd cvms

# Create a configuration file based on the example
cp .resource/example-validator-config.yaml config.yaml

# Copy the .env file for external DB settings
cp .resource/.env.example .env

# Modify environment variables for persistent mode
# (See example .env configuration below)

# Edit the config file to match your validator setup
vi config.yaml

# Override the default Docker Compose file to disable the internal Postgres service
cp .resource/docker-compose.override.yaml.persistent_mode docker-compose.override.yaml

# Start CVMS
docker compose up --build -d
```

**Example .env**

```ini
###### CVMS Services #######
#EXPORTER_PORT=9200
#INDEXER_PORT=9300
#LOG_COLOR_DISABLE=false
#LOG_LEVEL=4
#CONFIG_PATH=./config.yaml
#CUSTOM_CHAINS_FILE=custom_chains.yaml
# If you don't want to delete old records, use "persistence" instead of specific period
DB_RETENTION_PERIOD=persistence
# If you're operating docker service in not default directory, please enable this env for cadvisor and promtail
# DOCKER_ROOT=/data/docker

####### Prometheus Service #######
# PROM_SERVER_PORT=9090
# PROM_CONFIG_FILE=custom-prometheus.yml
# PROM_RETENTION_TIME=1m

####### Postgres Service #######
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cvms
DB_USER=cvms
DB_PASSWORD=mysecretpassword


####### Grafana Service #######
#ADMIN_USER=admin
#ADMIN_PASSWORD=admin
#GRAFANA_HOME_DASHBOARD=/etc/grafana/provisioning/dashboards/network/network-status-dashboard.json
#GRAFANA_SERVER_PORT=3000

####### Alertmanager Service #######
#ALERTMANAGER_SERVER_PORT=9093
```

**Example config.yaml**

```yaml
# NOTE: Customize this variables by your needs
# 1. network mode:
#   ex) monikers: ['all']
#   des) This will enable network mode to monitor all validators status in the blockchain network
#
# 2. validator mode:
#   ex) monikers: ['Cosmostation1', 'Cosmostation2']
#   des) This will enable validator mode for whitelisted specific validators
monikers: ['all']

# if user is one of validators, they want to operator for whole chains which already operating as validator.
chains:
  - display_name: 'cosmos'
    # NOTE: chain_id is a key for support_chains list. YOU SHOULD match correct CHAIN ID
    chain_id: cosmoshub-4
    # NOTE: these addresses will be used for balance usage tracking such as validator, broadcaster or something.
    tracking_addresses:
      - 'cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep4tgu9q'
    nodes:
      # NOTE: currently grpc endpoint doesn't support ssl
      - rpc: 'https://rpc-cosmos.endpoint.xyz'
        api: 'https://lcd-cosmos.endpoint.xyz'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'
```
