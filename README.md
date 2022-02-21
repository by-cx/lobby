# Lobby - simple server/service discovery service

TLDR: This a labeling tool for your servers. Like AWS resource tags but available everywhere.

In one of ours projects we needed service discovery that doesn't need complicated setup just to share
a simple information about running services and checking if they are still alive. So we came up with
this small service we called Lobby. It's like a lobby for users in games but in this case there are
servers instead. Each server runs one or more instances of lobby daemon and it regularly sends info
about its hostname and configured labels.

Labels are similar what you could know from AWS. It's basically alternative to resources' tags feature
with the only different that you can use this anywhere including AWS. Every server sends something called
"discovery packet" which is basically a json that looks like this:

```json
{
    "hostname": "smtp.example.com",
    "labels": [
        "service:smtp",
        "public_ip4:1.2.3.4",
        "public_ip6:2a03::1"
    ]
}
```

The packet contains information what's the server's hostname and then list of labels describing
what's running on it, what are the IP addresses, services, directories to backup or server's location for example.
What's in the labels is completely up to you but in some use-cases (Node Exporter API endpoint) it
expects "NAME:VALUE" format.

The labels can be configured via environment variables but also as files located in
*/etc/lobby/labels* (configurable path) so it can dynamically change. Another way is to use
*lobbyctl* which can add new labels at runtime.

When everything is running just call your favorite http client against "http://localhost:1313/"
on any of the running instances and lobby returns a list of all available servers and
their labels. You can hook it to Prometheus, deployment scripts, CI/CD automations or
your internal system that sends emails and it needs to know where is the SMTP server for
example.

Lobby doesn't care if you have a one or thousand instances of it running. Each instance
is connected to a common point which is a [NATS server](https://nats.io/) or Redis. NATS is super fast and reliable
messaging system which handles the communication part but also the high availability part.
NATS is easy to run and it offloads a huge part of the problem from lobby itself. But Redis
is not a bad choice either in some cases.

The code is open to support multiple backends and it's not that hard to add a new one.
Support for NATS is only less than 150 lines.

## Quickstart guide

The quickest way how to run lobby on your server is this:

```shell
wget -O /usr/local/bin/lobbyd https://github.com/by-cx/lobby/releases/download/v1.2/lobbyd-1.2-linux-amd64
chmod +x /usr/local/bin/lobbyd
wget -O /usr/local/bin/lobbyctl https://github.com/by-cx/lobby/releases/download/v1.2/lobbyctl-1.2-linux-amd64
chmod +x /usr/local/bin/lobbyctl

# Update NATS_URL and LABELS here
cat << EOF > /etc/systemd/system/lobbyd.service
[Unit]
Description=Server Lobby service
After=network.target

[Service]
Environment="NATS_URL=tls://nats.example.com:4222"
Environment="LABELS=service:ns,ns:primary,public_ip4:1,2,3,4,public_ip6:2a03::1,location:prague"
ExecStart=/usr/local/bin/lobbyd
PrivateTmp=false

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl start lobbyd
systemctl enable lobbyd
```

If you run lobbyd in production, consider to create its own system user and group and add both into this
service file. It doesn't need to access everything in your system.

To test if local instance is running call this:

    lobbyctl discovery

## Daemon

There are other config directives you can use to fine-tune lobbyd to exactly what you need.

| Environment variable     | Type   | Default           | Required          | Note                                                                                                                                                    |
| ------------------------ | ------ | ----------------- | ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| TOKEN                    | string |                   | no                | Authentication token for API, if empty auth is disabled                                                                                                 |
| HOST                     | string | 127.0.0.1         | no                | IP address used for the REST server to listen                                                                                                           |
| PORT                     | int    | 1313              | no                | Port related to the address above                                                                                                                       |
| DISABLE_API              | bool   | false             | no                | If true API interface won't start                                                                                                                       |
| DRIVER                   | string | NATS              | yes               | Selects which driver is used to exchange the discovery packets.                                                                                         |
| NATS_URL                 | string |                   | yes (NATS driver) | NATS URL used to connect to the NATS server                                                                                                             |
| NATS_DISCOVERY_CHANNEL   | string | lobby.discovery   | no                | Channel where the keep-alive packets are sent                                                                                                           |
| REDIS_HOST               | string | 127.0.0.1"        | no                | Redis host                                                                                                                                              |
| REDIS_PORT               | uint16 | 6379              | no                | Redis port                                                                                                                                              |
| REDIS_DB                 | string | 0                 | no                | Redis DB                                                                                                                                                |
| REDIS_CHANNEL            | string | lobby:discovery   | no                | Redis channel                                                                                                                                           |
| REDIS_PASSWORD           | string |                   | no                | Redis password                                                                                                                                          |
| LABELS                   | string |                   | no                | List of labels, labels should be separated by comma                                                                                                     |
| LABELS_PATH              | string | /etc/lobby/labels | no                | Path where filesystem based labels are located, one label per line, filename is not important for lobby                                                 |
| RUNTIME_LABELS_FILENAME  | string | _runtime          | no                | Filename for file created in LabelsPath where runtime labels will be added                                                                              |
| HOSTNAME                 | string |                   | no                | Override local machine's hostname                                                                                                                       |
| CLEAN_EVERY              | int    | 15                | no                | How often to clean the list of discovered servers to get rid of the not alive ones [secs]                                                               |
| KEEP_ALIVE               | int    | 5                 | no                | how often to send the keep-alive discovery message with all available information [secs]                                                                |
| TTL                      | int    | 30                | no                | After how many secs is discovery record considered as invalid                                                                                           |
| NODE_EXPORTER_PORT       | int    | 9100              | no                | Default port where node_exporter listens on all registered servers, this is used when the special prometheus labels doesn't contain port                |
| REGISTER                 | bool   | true              | no                | If true (default) then local instance is registered with other instance (discovery packet is sent regularly), if false the daemon runs only as a client |
| CALLBACK                 | string |                   | no                | Path to a script that runs when the the discovery packet records are changed. Not running for first                                                     |
| CALLBACK_COOLDOWN        | int    | 15                | no                | Cooldown prevents the call back script to run sooner than configured amount of seconds after last run is finished.                                      |
| CALLBACK_FIRST_RUN_DELAY | int    | 30                | no                | Wait for this amount of seconds before callback is run for first time after fresh start of the daemon                                                   |


### Callback script

When your application cannot support Lobbyd's API it can be configured via callback script that runs everytime something has changed in the network. Callback script is run every 15 seconds (configured by CALLBACK_COOLDOWN) but only when something has changed.

The script runs under the same user as lobbyd. When lobbyd starts first thirty seconds (CALLBACK_FIRST_RUN_DELAY) is ignored and then the script is run for first time. After these thirty seconds everything runs in loop based on the changes in the network.

All current discovery packets are passed to the callback script via standard input. It's basically the same input you get if you run `lobbyctl discoveries`.

### Service discovery for Prometheus

Lobbyd has an API endpoint that returns list of targets for [Prometheus's HTTP SD config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config). That
allows you to use lobbyd to configure Prometheus dynamically based on running servers. There are special kind of labels that are used to set the output for Prometheus properly.
Let's check this:

    prometheus:nodeexporter:host:192.168.1.1
    prometheus:nodeexporter:port:9100
    prometheus:nodeexporter:location:prague

If you set port to *-* lobby daemon omits port entirely from the output.

There can be multiple `host` labels but only one `port` label and all prometheus labels (last line) will be common for all hosts labels. If port is omitted then default 9100 is used or port can also be part of the host label.

When you open URL http://localhost:1313/v1/prometheus/nodeexporter it returns this:

```json
[
    {
        "Labels": {
            "location": "prague"
        },
        "Targets": [
            "192.168.1.1:9100"
        ]
    }
]
```

"nodeexporter" can be anything you want. It determines name of the monitored service, the service that provides the */metrics* endpoint.

There is also a minimal way how to add server to the prometheus output. Simply set label *prometheus:nodeexporter* and it
will use default port from the environment variable above and hostname of the server

```json
[
    {
        "Labels": {},
        "Targets": [
            "192.168.1.1:9100"
        ]
    }
]
```

At least one prometheus label has to be set to export the monitoring service in the prometheus output.

## Command line tool

To access your servers from command line or shell scripts you can use *lobbyctl*.

```
Usage of lobbyctl:
  -host string
    	Hostname or IP address of lobby daemon
  -port uint
    	Port of lobby daemon
  -proto string
    	Select HTTP or HTTPS protocol
  -token string
    	Token needed to communicate lobby daemon, if empty auth is disabled

Commands:
  discovery                      returns discovery packet of the server where the client is connected to
  discoveries                    returns list of all registered discovery packets
  labels add LABEL [LABEL] ...   adds new runtime labels
  labels del LABEL [LABEL] ...   deletes runtime labels
```

It uses Go client library also located in this repository.


## REST API

So far the REST API is super simple and it has only two endpoints:

```
GET /                                                  # Same as /v1/discoveries
GET /v1/discovery                                      # Returns current local discovery packet
GET /v1/discoveries                                    # Returns list of all discovered servers and their labels.
GET /v1/discoveries?labels=LABELS&prefixes=PREFIXES    # output will be filtered based on one or multiple labels separated by comma or it can search for given prefixes, only one of those will be used
GET /v1/prometheus/:name                               # Generates output for Prometheus's SD config, name is group of the monitoring services described above.
POST /v1/labels                                        # Add runtime labels that will persist over daemon restarts. Labels should be in the body of the request, one line per one label.
DELETE /v1/labels                                      # Delete runtime labels. One label per line. Can't affect the labels from environment variables or labels added from the LabelPath.
```

If there is an error the error message is returned as plain text.

## API clients

* Golang client is part of this repository.
* There is also [Python client](https://github.com/by-cx/lobby-python) available.

## Notes

I wanted to use SQS or SNS as backend but when I checked the services I found out
it wouldn't work. SNS would require open HTTP server whicn is hard to do in our
infrastructure and I couldn't find a way how SQS could deliver every message
to all instances of lobbyd.

Instead I decided to implement Redis because it's much easier to use for
development and testing. But there is no reason why it couldn't work in production
too.


## TODO

* [X] Tests
* [X] Command hooks - script or list of scripts that are triggered when discovery status has changed
* [ ] Support for multiple active backend drivers
* [X] Redis driver
* [X] Remove the 5 secs waiting when daemon is stopped
* [X] API to allow add labels at runtime
* [ ] Check what happens when driver is disconnected


