# pgweb (Maintained by Flow.BI) 

Simple web-based and cross platform PostgreSQL database explorer. Flow.BI is embedding it in the Flow.BI user interface as a easy option to access the PostgreSQL based SQL API via the web-based user interface.

[![Release](https://img.shields.io/github/release/flowbi/pgweb.svg?label=Release)](https://github.com/flowbi/pgweb/releases)
[![Linux Build](https://github.com/flowbi/pgweb/actions/workflows/checks.yml/badge.svg)](https://github.com/flowbi/pgweb/actions?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/flowbi/pgweb)](https://goreportcard.com/report/github.com/flowbi/pgweb)
[![GoDoc](https://godoc.org/github.com/flowbi/pgweb?status.svg)](https://godoc.org/github.com/flowbi/pgweb)
[![Docker Pulls](https://img.shields.io/docker/pulls/flowbi/pgweb.svg)](https://hub.docker.com/r/flowbi/pgweb/)

## Overview

Pgweb is a web-based database explorer for PostgreSQL, written in Go, and works
on Mac, Linux and Windows machines. Distributed as a simple binary with zero dependencies.
Very easy to use and packs just the right amount of features.

Flow.BI is a generative AI service to integrate independent enterprise datasets. pgweb is used to provide a convenient graphical user interface to the SQL API inside the web application.

To implement missing features, latest changes by the community, and to customize the application for better embedding in applications, it was decided to maintain a fork from the original author. Our development effort will focus on features required for the Flow.BI application. However, we are open to pull requests from the community and will add them to this fork.

Our goal is to keep the original design principles by the original author, especially simplicity. In addition, we are going to support and maintain the original features, including the standalone availability of pgweb (the use of Flow.BI is not required) to maintain the highest value for the user community.

Our use case is to embed pgweb inside our Flow.BI user interface:

<img width="1678" height="1058" alt="image" src="https://github.com/user-attachments/assets/26ed09f0-0229-429e-bc81-b67eef1e0475" />

Additional screenshots exist that show pgweb in standalone action:

[See original application screenshots](SCREENS.md)

## Features

- Cross-platform: Mac/Linux/Windows (64bit).
- Simple installation (distributed as a single binary).
- Zero dependencies.
- Works with PostgreSQL 9.1+.
- Supports native SSH tunnels.
- Multiple database sessions.
- Execute and analyze custom SQL queries.
- Table and query data export to CSV/JSON/XML.
- Query history.
- Server bookmarks.

Visit [WIKI](https://github.com/flowbi/pgweb/wiki) for more details.

## Demo

Visit https://pgweb-demo.fly.dev/ to see the original Pgweb in action. Flow.BI is not maintaining a demo at this time.

## Installation

- [Precompiled binaries](https://github.com/flowbi/pgweb/releases) for supported operating systems are available.
- [More installation options](https://github.com/flowbi/pgweb/wiki/Installation)

## Usage

Start server:

```
pgweb
```

You can also provide connection flags:

```
pgweb --host localhost --user myuser --db mydb
```

Connection URL scheme is also supported:

```
pgweb --url postgres://user:password@host:port/database?sslmode=[mode]
pgweb --url "postgres:///database?host=/absolute/path/to/unix/socket/dir"
```

### Multiple database sessions

To enable multiple database sessions in pgweb, start the server with:

```
pgweb --sessions
```

Or set environment variable:

```
PGWEB_SESSIONS=1 pgweb
```

## Testing

Before running tests, make sure you have PostgreSQL server running on `localhost:5432`
interface. Also, you must have `postgres` user that could create new databases
in your local environment. Pgweb server should not be running at the same time.

Execute test suite:

```
make test
```

If you're using Docker locally, you might also run pgweb test suite against
all supported PostgreSQL version with a single command:

```
make test-all
```

## Contribute

- Fork this repository
- Create a new feature branch for a new functionality or bugfix
- Commit your changes
- Execute test suite
- Push your code and open a new pull request
- Use [issues](https://github.com/flowbi/pgweb/issues) for any questions
- Check [wiki](https://github.com/flowbi/pgweb/wiki) for extra documentation

## License

The MIT License (MIT). See [LICENSE](LICENSE) file for more details.
