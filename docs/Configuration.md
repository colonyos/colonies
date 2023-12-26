# Configuration 
To make Kubernetes deployments easier, Colonies does not have configuration files, but is rather configured using environmental variables. 

## Environmental variables 
### Database 
Colonies requires the following variables to be able to successfully connect to a PostgreSQL server.

```console
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8
export LC_CTYPE=UTF-8
export TZ=Europe/Stockholm
export COLONIES_DB_HOST="localhost"
export COLONIES_DB_USER="postgres"
export COLONIES_DB_PORT="50070"
export COLONIES_DB_PASSWORD="rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
```

### CLI 
The following variables are utilized by the CLI tool to minimize the number of flags required when executing commands.

```console
export COLONIES_SERVER_TLS="false"
export COLONIES_SERVER_HOST="localhost"
export COLONIES_SERVER_PORT="50080"
export COLONIES_SERVER_ID="039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103"
export COLONIES_SERVER_PRVKEY="fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d"
export COLONIES_COLONY_ID="4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
export COLONIES_COLONY_PRVKEY="ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514"
export COLONIES_EXECUTOR_ID="3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac"
export COLONIES_EXECUTOR_PRVKEY="ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"
export COLONIES_EXECUTOR_TYPE="cli"
```

### Prometheus monitoring 
The Colonies server has built-in support for Prometheus instrumentation. The variables below controls which port the monitoring server should run at and how it metrics should be collected. 

```console
export COLONIES_MONITOR_PORT="21120"
export COLONIES_MONITOR_INTERVAL="1"
```

### Cron and generator frequency 
The environmental variable below defines how often the Colonies server should run the cron and generator workers. By setting the variables to 1000, the checker workers will run every second.

```console
export COLONIES_CRON_CHECKER_PERIOD="1000"
export COLONIES_GENERATOR_CHECKER_PERIOD="1000"
```

### Exclusive assign 
When exclusive assignment is enabled, all assign requests are handled by the elected leader server in the Colonies cluster. This guarantees that a process is assigned to exactly one executor. However, this approach may result in reduced performance as assign requests cannot be evenly distributed across all Colonies server replicas for load balancing purposes."

```console
export COLONIES_EXCLUSIVE_ASSIGN="true"
```

### Re-registration 
Configure the variable below to enable executors to re-registration without prior unregistration. This functionality can for example be useful in a Kubernetes environments where ungraceful termination of a Pod may hinder the executors ability to unregister gracefully."

```console
export COLONIES_ALLOW_EXECUTOR_REREGISTER="false"
```

### Retention 
The variables below to automatically purge successful processes older than 604800 seconds (1 week).

```console
export COLONIES_RETENTION="false"
export COLONIES_RETENTION_POLICY="604800"
```

### Profiling
It is possible to use the Golang pprof tool to profile the Colonies code.

```console
export COLONIES_SERVER_PROFILER="false"
export COLONIES_SERVER_PROFILER_PORT="6060"
```

Set the variable above and generate a memory usage PDF report using the command below.

```console
go tool pprof --pdf  http://rocinante:6060/debug/pprof/allocs
```
