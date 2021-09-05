# Lobby - simple server/service discovery service

In one of ours projects we needed service discovery that doesn't need complicated setup just to share
a simple information about running services and checking if they are still alive. So we came up with
this small service we call Lobby. It's like a lobby in games but in this case there are servers. Each
server runs one or more instances of lobby daemon and it regularly sends how it's configured.

We call the information about the server and services running on it *labels*. Every server shares
"discovery packet" which is basically a json that looks like this:

```json
{
    "hostname": "smtp.example.com",
    "labels": [
        "service:smtp",
        "public_ip4:1.2.3.4",
        "public_ip6:2a03::1"
    ],
    "last_check": 1630612478
}
```

The packet contains information what's the server hostname and then list of labels describing
what's running on it and what are the IP addresses. What's in the labels is completely up to you
but in some use-cases (Node Exporter API endpoint) it expects "NAME:VALUE" format.

The labels can be configured via environment variables but also as files located in
*/etc/lobby/labels* (configurable path) so it can dynamically change.

When everything is running just call your favorite http client against "http://localhost:1313/"
on any of the running instances and lobby returns you list of all available servers and
their labels. You can hook it to Prometheus, deployment scripts, CI/CD automations or
your internal system that sends emails and it needs to know where is the SMTP server for
example.

Lobby doesn't care if you have a one or thousand instances of it running. Each instance
is connected to a common point which is a [NATS server](https://nats.io/) in this case. NATS is super fast and reliable
messaging system which handles the communication part but also the high availability part.
NATS is easy to run and it offloads a huge part of the problem from lobby itself.

The code is open to support multiple backends and it's not that hard to add a new one.

## Quickstart guide

The quickest way how to run lobbyd on your server is this:

```shell
wget -O /usr/local/bin/lobbyd https://....
chmod +x /usr/local/bin/lobbyd
wget -O /usr/local/bin/lobbyctl https://....
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
service file. It doesn't need to access almost anything in your system.

To test if local instance is running call this:

    lobbyctl discovery

## Daemon

There are other config directives you can use to fine-tune lobbyd to exactly what you need.

| Environment variable    | Type   | Default           | Required | Note                                                                                                                                                    |
| ----------------------- | ------ | ----------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| TOKEN                   | string |                   | no       | Authentication token for API, if empty auth is disabled                                                                                                 |
| HOST                    | string | 127.0.0.1         | no       | IP address used for the REST server to listen                                                                                                           |
| PORT                    | int    | 1313              | no       | Port related to the address above                                                                                                                       |
| NATS_URL                | string |                   | yes      | NATS URL used to connect to the NATS server                                                                                                             |
| NATS_DISCOVERY_CHANNEL  | string | lobby.discovery   | no       | Channel where the keep-alive packets are sent                                                                                                           |
| LABELS                  | string |                   | no       | List of labels, labels should be separated by comma                                                                                                     |
| LABELS_PATH             | string | /etc/lobby/labels | no       | Path where filesystem based labels are located, one label per line, filename is not important for lobby                                                 |
| RUNTIME_LABELS_FILENAME | string | _runtime          | no       | Filename for file created in LabelsPath where runtime labels will be added                                                                              |
| HOSTNAME                | string |                   | no       | Override local machine's hostname                                                                                                                       |
| CLEAN_EVERY             | int    | 15                | no       | How often to clean the list of discovered servers to get rid of the not alive ones [secs]                                                               |
| KEEP_ALIVE              | int    | 5                 | no       | how often to send the keep-alive discovery message with all available information [secs]                                                                |
| TTL                     | int    | 30                | no       | After how many secs is discovery record considered as invalid                                                                                           |
| NODE_EXPORTER_PORT      | int    | 9100              | no       | Default port where node_exporter listens on all registered servers, this is used when the special prometheus labels doesn't contain port                |
| REGISTER                | bool   | true              | no       | If true (default) then local instance is registered with other instance (discovery packet is sent regularly), if false the daemon runs only as a client |


### Service discovery for Prometheus

Lobbyd has an API endpoint that returns list of targets for [Prometheus's HTTP SD config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#http_sd_config). That
allows you to use lobbyd to configure Prometheus dynamically based on running servers. There are special kind of labels that are used to set the output for Prometheus properly.
Let's check this:

    prometheus:nodeexporter:host:192.168.1.1
    prometheus:nodeexporter:port:9100
    prometheus:nodeexporter:location:prague

If you set port to *-* lobby daemon omits port entirely from the output.

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
GET /                                # Same as /v1/discoveries
GET /v1/discovery                    # Returns current local discovery packet
GET /v1/discoveries                  # Returns list of all discovered servers and their labels.
GET /v1/discoveries?labels=LABELS    # output will be filtered based on one or multiple labels separated by comma
GET /v1/prometheus/:name             # Generates output for Prometheus's SD config, name is group of the monitoring services described above.
POST /v1/labels                      # Add runtime labels that will persist over daemon restarts. Labels should be in the body of the request, one line per one label.
DELETE /v1/labels                    # Delete runtime labels. One label per line. Can't affect the labels from environment variables or labels added from the LabelPath.
```

If there is an error the error message is returned as plain text.

## TODO

* [X] Tests
* [ ] Command hooks - script or list of scripts that are triggered when discovery status has changed
* [ ] Support for multiple active backend drivers
* [ ] SNS driver
* [X] API to allow add labels at runtime



