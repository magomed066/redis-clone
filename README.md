# Redis Clone in Go

A small Redis-inspired, in-memory key-value server written in Go. I built this project for learning purposes: to practice TCP networking, parsing the Redis Serialization Protocol (RESP), working with goroutines and mutexes, implementing key expiration, and saving data to disk.

This is an educational project, not a production-ready database or a complete Redis implementation.

## Features

- String keys and values stored in memory.
- TCP server with a separate goroutine for each client connection and mutex-protected store operations.
- A subset of Redis commands using RESP arrays of bulk strings.
- Key expiration, checked on access and cleaned up by a background worker every second.
- JSON snapshots saved on request or graceful shutdown and restored at startup.
- Go standard library only; no third-party Go dependencies.

## Setup

### Requirements

- Go 1.27.2 or newer, as specified in `go.mod`.
- Git to clone the repository.
- `redis-cli` to try the examples below. Only the client is needed; you do not need to start a Redis server.

### Run the server

```sh
git clone https://github.com/magomed066/redis-clone.git
cd redis-clone
go run ./cmd
```

The server listens on **127.0.0.1:6380**. The address and port are currently hardcoded. Make sure port 6380 is available.

Alternatively, build and run an executable from the repository root:

```sh
go build -o redis-clone ./cmd
./redis-clone
```

Run the server from the same directory each time if you want to reuse its saved data: `snapshot.json` is read and written relative to the current working directory. That directory must be writable to save snapshots.

### Try it

In another terminal:

```sh
redis-cli -h 127.0.0.1 -p 6380 PING
# PONG

redis-cli -h 127.0.0.1 -p 6380 SET name "Go learner"
# OK

redis-cli -h 127.0.0.1 -p 6380 GET name
# "Go learner"

redis-cli -h 127.0.0.1 -p 6380 SET counter 0
redis-cli -h 127.0.0.1 -p 6380 INCR counter
# (integer) 1

redis-cli -h 127.0.0.1 -p 6380 SET temporary hello PX 5000
# OK — expires after 5 seconds

redis-cli -h 127.0.0.1 -p 6380 SAVE
# OK
```

Press **Ctrl+C** in the server terminal to stop it gracefully and save a final snapshot.

## Supported commands

Command names and the `PX` option are case-insensitive. Keys and values are case-sensitive.

| Command | Behavior |
| --- | --- |
| `PING` | Returns `PONG`. A message argument is not supported. |
| `ECHO message` | Returns the supplied message as a bulk string. |
| `SET key value` | Creates or overwrites a string value and removes any existing expiration. Returns `OK`. |
| `SET key value PX milliseconds` | Sets a value with an expiration in milliseconds. The duration must be a positive integer that fits in Go's `time.Duration`. Returns `OK`. |
| `GET key` | Returns the value, or a null bulk string (`nil` in `redis-cli`) if the key is missing or expired. |
| `INCR key` | Increments a signed 64-bit integer value and returns the result. A missing key starts at `1`. Preserves an existing expiration; invalid integers and overflow return errors. |
| `DEL key [key ...]` | Deletes one or more keys and returns the number actually deleted. Missing, expired, and repeated keys do not increase the count. |
| `EXPIRE key seconds` | Sets or replaces an expiration in seconds. Returns `1` for an existing key, otherwise `0`. Zero or negative seconds delete the key immediately. |
| `TTL key` | Returns remaining lifetime in whole seconds, rounded down; `-1` means no expiration and `-2` means the key is missing or expired. |
| `SAVE` | Writes a snapshot to `snapshot.json` and returns `OK` on success. |

Unknown commands and unsupported arguments return errors.

## Persistence

At startup, the server restores `snapshot.json` if it exists. A missing file starts an empty store; an unreadable or invalid snapshot stops startup with an error.

Snapshots contain values and absolute expiration timestamps. Expired entries are skipped when saving and loading, so time spent offline still counts toward expiration. Saving writes a temporary file, syncs and closes it, then renames it to replace the snapshot.

Snapshots are saved by `SAVE` and during graceful shutdown with Ctrl+C. There is no periodic autosave or append-only log. A crash, forced termination, or other shutdown that bypasses the interrupt handler can lose changes made since the last successful snapshot. SIGTERM is not handled as a graceful shutdown signal.

## Project structure

```text
cmd/
  main.go                 TCP server, connections, and shutdown
  resp.go                 RESP request parsing and response encoding
  commands.go             Command dispatch and argument validation
internal/store/
  store.go                In-memory storage, expiration, and synchronization
  snapshot.go             JSON snapshot saving and loading
  store_test.go           Basic storage and increment tests
go.mod                    Module path and Go version
```

## Tests

Run the existing tests from the repository root:

```sh
go test ./...
```

To run them with Go's race detector on a supported platform:

```sh
go test -race ./...
```

The current tests cover basic set/get/delete behavior and integer increments, including invalid values and overflow. They do not yet cover the network server, RESP parser, expiration, or snapshots.

## Limitations

- **Partial Redis compatibility:** only the commands and forms listed above are implemented. There are no lists, hashes, sets, transactions, pub/sub, replication, clustering, or multiple databases. `SET` options such as `EX`, `NX`, and `XX` are unsupported.
- **Limited protocol support:** requests must be RESP arrays of bulk strings. Inline commands and RESP3 are unsupported. Clients that require commands such as `HELLO`, `AUTH`, `SELECT`, or `CLIENT` may not work.
- **Request size limits:** at most 128 array elements (including the command name), 1 MiB per bulk string, and 4 MiB of combined bulk-string payload per command. Malformed or oversized requests close the connection.
- **Local development only:** the server binds to loopback and has no authentication, TLS, or configuration flags.
- **Memory and connection limits:** all data lives in memory, with no configured memory cap, eviction policy, connection cap, or client read/write timeout. Store operations share a single mutex.
- **Simple persistence:** snapshots are custom JSON, not Redis RDB or AOF files. Arbitrary binary strings are not guaranteed to round-trip through JSON because invalid UTF-8 is replaced during encoding. Snapshot saving also copies the live data into memory.
- **Learning-stage validation:** test coverage is limited, and the project has not been validated for production workloads or full Redis compatibility.
