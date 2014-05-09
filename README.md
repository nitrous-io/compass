## Nitrous.IO Authoritative DNS Server

This accepts DNS requests to our boxhost domain and correctly routes them to
the correct server. Acts as the replacement for Route53 that doesnt suck, and
also doesnt requires us to use Route53 API

For now it will read PostgreSQL database to fetch data.
