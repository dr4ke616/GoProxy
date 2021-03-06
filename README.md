
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
After installation is complete there will be a config file in `/etc/goproxy/config.json`.

#### Default
The default config file will as follows:

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://stackoverflow.com",
    "SSL": {
        "active": false,
        "key_file": "<PATH_TO_KEY>",
        "cert_file": "<PATH_TO_CERT>",
        "listening_port": "8099"
    },
    "routing_options": []
}
```
Set the `target_url` to any destination host you'd like. Then browse to `http://localhost:8088` to see if it works.

#### Static Routing Options
GoProxy also supports some custom routing options. You can alter the method types by adding a json object to the `routing_options`array and setting the values:

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://stackoverflow.com",
    "SSL": {
        "active": false,
        "key_file": "<PATH_TO_KEY>",
        "cert_file": "<PATH_TO_CERT>",
        "listening_port": "8099"
    },
    "routing_options": [
        {
            "uri": "/some-endpoint/",
            "to_method": "POST",
            "copy_paramaters": false,
            "custom_headers": []
        }
    ]
}
```

Go Proxy will also try to edit the headers in the request and the response to the proxied host. Specify the rules by adding a json object to the `custom_headers`array and setting the values:

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://stackoverflow.com",
    "SSL": {
        "active": false,
        "key_file": "<PATH_TO_KEY>",
        "cert_file": "<PATH_TO_CERT>",
        "listening_port": "8099"
    },
    "routing_options": [
        {
            "uri": "/some-endpoint/",
            "to_method": "POST",
            "copy_paramaters": true,
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

For example, say the target host has a header of `Content-Type: text/plain`. When `replace` is set to `false`, GoProxy will try represent the headers as `Content-Type: text/plain, application/json, application/xml`. If `replace` is set to `true`the headers will appear as `Content-Type: application/json, application/xml`, (removing the `text/plain` value).

If `copy_paramaters` was set to true GoProxy will try create a request body from any data encoded onto the url based on the `Content-Type`. At the moment only `application/json` and `application/x-www-form-urlencoded` is supported. For example:

If we have the URL `http://host/query?foo=bar&num=1&is_true=true&copy=str1&copy=str2`

If `application/json` was set as the `Content-Type`, this will build the following json:
```json
{
    "foo": "bar",
    "num": 1,
    "is_true": true,
    "copy": ["str1", "str2"],
}
```

If `application/x-www-form-urlencoded` was set as the `Content-Type`, the body will be set as the following:
```
foo=bar&num=1&is_true=true&copy=str1&copy=str2
```

#### SSL
GoProxy supports incomeing requests over HTTPS, just specify it in the config.

```json
{
    "log_file": "/var/log/goproxy.log",
    "listening_port": "8088",
    "target_url": "http://stackoverflow.com",
    "SSL": {
        "active": true,
        "key_file": "server.key",
        "cert_file": "server.crt",
        "listening_port": "8099"
    },
    "routing_options": []
}
```

Just set the active flag to true and specify a path to the key and cert file.

## Logs
If you wish to view GoProxy's log, the logs can be found at `/var/log/goproxy.log`. The logs location can be changed in the config file by setting the `log_file` value.

## Tests
If you wish to run the tests use the makefile provided by running `make test`

## TODO
- Copy Paramaters to support application/xml
- Dynamic routing options over URL
