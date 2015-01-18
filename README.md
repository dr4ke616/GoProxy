
# GoProxy

[![Build Status](https://travis-ci.org/dr4ke616/GoProxy.svg?branch=master)](https://travis-ci.org/dr4ke616/GoProxy)

GoProxy is a simple HTTP web proxy.

## Installation
GoProxy can be installed by compiling from source. In order to be able to install GoProxy you must have golang installed with your enviroment setup and GOPATH configured correctly. You can download the latest version of Go [here](https://golang.org/doc/install).

#### Install with wget
```
wget --no-check-certificate https://raw.github.com/dr4ke616/GoProxy/master/scripts/install -O - | bash
```

#### With curl
```
curl -L https://raw.github.com/dr4ke616/GoProxy/master/scripts/install | bash
```

## Start / Stop / Restart
- To start GoProxy run `sudo /etc/init.d/goproxy start`
- To stop GoProxy run `sudo /etc/init.d/goproxy stop`
- To restart GoProxy run `sudo /etc/init.d/goproxy restart`
- To check the status of GoProxy run `sudo /etc/init.d/goproxy status`

## Config
After instalation is complete there will be a config file in `/etc/goproxy/config.json`.

#### Default
The default config file will as follows:

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://localhost:8080",
    "routing_options": []
}
```
Set the `target_url` to any destination host you'd like. Then browse to `http://localhost:8088` to see if it works.

#### Routing Options
GoProxy also supports some custom routing options. You can alter the method types by adding a json onbject to the `routing_options`array and setting the values:

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://localhost:8080",
    "routing_options": [
        {
            "uri": "/some-endpoint/",
            "from_method": "GET",
            "to_method": "POST",
            "custom_headers": []
        }
    ]
}
```

Go Proxy will also try to edit the headers in the request and the response to the proxied host. Specify the rules by adding a json onbject to the `custom_headers`array and setting the values:

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://localhost:8080",
    "routing_options": [
        {
            "uri": "/some-endpoint/",
            "from_method": "GET",
            "to_method": "POST",
            "custom_headers": [
                {
                    "replace": false,
                    "header_key": "Content-Type",
                    "header_values": ["application/json", "application/xml"]
                }
            ]
        },
    ]
}
```

The `replace` boolean when set to `false` will append on the values specified in the `header_values` array to the header key specified in `header_key` value. When the `replace` boolean is set to `true` it will overwrite any headers that may be set for the `header_key` value.

For example, say the target host has a header of `Content-Type: text/plain`. When `replace` is set to `false`, given the headers we specified in the above config example, GoProxy will try represent the headers as `Content-Type: text/plain, application/json, application/xml`. If `replace` is set to `true`the headers will appear as `Content-Type: application/json, application/xml`, (removing the `text/plain` value).

## Config
If you wish to view GoProxy's log, the logs can be found at `/var/log/goproxy.log`. The can be changed in the config file by setting the `log_file` value.
