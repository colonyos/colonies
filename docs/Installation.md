## Installation
### Start a TimescaleDB server
```console
docker run -d --name timescaledb -p 5432:5432 -v /storage/fast/lib/timescaledb/data:/var/lib/postgresql/data -e POSTGRES_PASSWORD=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7 --restart unless-stopped timescale/timescaledb:latest-pg12
```

### Setup a database
```console
./bin/colonies database create --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

### In case, you would like to clear the database
```console
./bin/colonies database drop --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```

### Start a Colonies server 
```console
./bin/colonies server start --rootpassword=secret --port=8080 --tlscert=./cert/cert.pem --tlskey=./cert/key.pem --dbhost localhost --dbport 5432 --dbuser postgres --dbpassword=rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7
```
