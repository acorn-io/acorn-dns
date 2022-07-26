# Acorn-DNS

FQDNs on demand. Powering on-acorn.io

Will create A, AAAA, CNAME, and TXT records in Route53.

Backed by a SQL database. Supports sqlite for development and Maria/MySQL for production.


## CLI

```help
./bin/acorn-dns api-server -h
NAME:
   acorn-dns api-server - acorn api server

USAGE:
   acorn-dns api-server [command options] [arguments...]

OPTIONS:
   --port value                        HTTP Server Port (default: 4315) [$ACORN_DNS_PORT]
   --route53-zone-id value             AWS Route53 Zone ID where records will be created [$ACORN_ROUTE53_ZONE_ID]
   --route53-record-ttl-seconds value  AWS Route53 record TTL (default: 300) [$ACORN_ROUTE53_RECORD_TTL_SECONDS]
   --purge-interval-seconds value      How often to run the domain and record purge daemon. Default 86,400 (1 day) (default: 86400) [$ACORN_PURGE_INTERVAL_SECONDS]
   --domain-max-age-seconds value      Max age a domain can be without being renewed before it's deleted. Default 2,592,000 (30 days) (default: 2592000) [$ACORN_DOMAIN_MAX_AGE_SECONDS]
   --record-max-age-seconds value      Max age a domain can be without being renewed before it's deleted. Default 172,800 (2 days) (default: 172800) [$ACORN_RECORD_MAX_AGE_SECONDS]
   --db-engine value                   The type of DB to connect to, sqlite or mariadb (default: "sqlite") [$ACORN_DB_ENGINE]
   --db-sqlite-dsn value               The DSN to use to connect to a sqlite db (default: "file:acorn.sqlite?_pragma=foreign_keys(1)") [$ACORN_DB_SQLITE_DSN]
   --db-user value                     Database user [$ACORN_DB_USER]
   --db-password value                 Database password [$ACORN_DB_PASSWORD]
   --db-name value                     Name of the database [$ACORN_DB_NAME]
   --db-host value                     Database host [$ACORN_DB_HOST]
   --db-port value                     Database port [$ACORN_DB_PORT]
   --log-level value, -l value         Log Level (default: "info") [$LOGLEVEL]
   --log-caller                        log the caller (aka line number and file) (default: false)
   --log-disable-color                 disable log coloring (default: false)
   --log-full-timestamp                force log output to always show full timestamp (default: false)
   --help, -h                          show help (default: false)
```
