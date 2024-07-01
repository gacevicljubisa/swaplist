# SwapList

SwapList is a tool to track transactions of a swarm node. It provides functionalities to retrieve transaction sender addresses with timestamps from the Gnosis Chain using either the Chainstack API and Gnosis RPC or the Gnosis Scan API.

## Features

- Retrieve transaction sender addresses with timestamps.
- Save transaction data to a file.

## Requirements

- Go 1.22 or later

## Installation

```sh
git clone https://github.com/gacevicljubisa/swaplist.git
cd swaplist
make binary
```

### Usage

#### Limit

Command that retrieves a limited number of transaction sender addresses with timestamps from the Gnosis Chain. 

```sh
./dist/swaplist limit
```

It has the following flags:

```console
  -a, --address string   Contract address on Gnosis Chain (default "0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420")
  -k, --apikey string    API key for Gnosis Scan (default "DEN397GUGXKJN6T14HU2W8MTZVVMXZ57AU")
  -n, --number uint32    Number of addresses to retrieve (0-10000) (default 1000)
  -o, --order string     Order to retrieve addresses (asc/desc) (default "asc")
```

#### Full

Long running command that retrieves all transaction sender addresses with timestamps from the Gnosis Chain from logs.

```sh
./dist/swaplist full
```

It has the following flags:

```console
  -a, --address string   Contract address on Gnosis Chain (default "0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420")
```