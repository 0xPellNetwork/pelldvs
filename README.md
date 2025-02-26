# PellDVS

PellDVS is a security engine designed to provide economic security for DVS applications based on the Pell Chain. Developers can use PellDVS to quickly build DVS applications, leveraging its network and security capabilities as the underlying layer for these applications.

PellDVS is structured around three major technical components:

- **Network Layer**: PellDVS offers a standard RPC interface and P2P communication interface. It provides a unified user interface for upper-layer applications and facilitates the interconnection of different DVS networks.
- **Security Layer**: PellDVS includes a unified signature service and aggregation service, ensuring secure operations for upper-layer applications.
- **Application Layer (AVSI Interface)**: PellDVS exposes an Abstract Validated Service Interface (AVSI), allowing upper-layer applications to implement the interface specification in any language and use the underlying capabilities of PellDVS.

## Quick Start

Follow the steps below to set up PellDVS locally and run an end-to-end (e2e) test.

### 1. Clone the Repository

First, clone this repository and navigate to the e2e test directory:

```
git clone https://github.com/0xPellNetwork/pelldvs && cd pelldvs/test/e2e
```

### 2. Build the Docker Image

Build the necessary Docker images:

```
GITHUB_TOKEN=<YOUR_GITHUB_TOKEN> make docker-build-all
```

### 3. Start All Dependent Nodes

Start all required nodes in detached mode, ensuring that unused containers are removed:

```
docker compose up operator -d --remove-orphans
```

### 4. Run the Local e2e Test

Run the end-to-end test:

```
make docker-test-pelle2e
```

### 5. Clean Up the Node Environment

Once the e2e test completes, clean up the node environment:

```
make docker-down
```

## Documentation

For detailed documentation, visit the official [PellDVS Documentation](https://docs.pell.network/dvs-developer-guides/introduction/).

## Releases

To ensure stability, avoid using the `main` branch for production. Instead, rely on the latest [releases](https://github.com/0xPellNetwork/pelldvs/releases).

For production environments or if you need assistance, please feel free to reach out to us via one of the following methods, listed in order of preference:

- [Create a new discussion on GitHub](https://github.com/0xPellNetwork/pelldvs/discussions)
- Contact us on [Telegram](https://t.me/Pell_Network)
- Join the [Pell Network Discord](https://discord.gg/interchain) and participate in the `#developers` channel.

For more details on how releases are managed, check [here](./RELEASES.md).

## Security

To report security vulnerabilities, please refer to our [bug bounty program](https://docs.pell.network/security/bug-bounty-program).

## Contributing

We encourage contributions to PellDVS! Please review our [Code of Conduct](CODE_OF_CONDUCT.md) and familiarize yourself with the [contributing guidelines](CONTRIBUTING.md) and [style guide](STYLE_GUIDE.md) before getting involved. Additionally, you may find it useful to read the [specifications](./spec/README.md) to better understand the project.

### Additional Notes:

- **Issue Tracker**: For any issues or bugs, please check the [Issues page](https://github.com/0xPellNetwork/pelldvs/issues) and report new issues if necessary.
- **Testing**: We rely on automated tests to maintain code quality. Please ensure that any changes are thoroughly tested.

Thank you for contributing to PellDVS!
