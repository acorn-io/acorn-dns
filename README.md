# Acorn-DNS

FQDNs on demand. Powering on-acorn.io

Will create A, AAAA, CNAME, and TXT records in Route53.

Backed by a SQL database. Supports sqlite for development and Maria/MySQL for production.


## CLI

```help
NAME:
   acorn-dns - Let's do DNS

USAGE:
   acorn-dns [global options] command [command options] [arguments...]

VERSION:
   0.1.0-main

AUTHOR:
   The Acorn Labs Dev Team <engineering@acorn.io>

COMMANDS:
   api-server  acorn api server
   version     print version
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```
