### api service 

Testing with 2 terminal sessions. You may want to proxy a k8s cluster
for the k8s test to use localhost:8888, or modify the yaml/json in
tests configuration.


If using the support scripts for kubernetes private test cluster you
may want to enable that scripting with the following options.

```
cp /path/to/k8s-cfg/k8s-cfg  .
ln -s /path/to/k8s-cfg/.k8s-cfg .
```


```
kubectl proxy --port=8888 --api-prefix=/
```

*Added consul api protocol to put, then get of the same key value using local instance consul*

Test protocol: consul send arg and check receive value.

```
apiVersion: v0.1
kind: test
spec:
- name: consul-check
  protocol: consul
  send:
    spec:
      arg:
        - '{ "name": "value", "type": "value" }'
      env:
        name: value
      uri: consul://localhost:8500
  recv:
    ack: 
      recv: '{ "name": "value", "type": "value" }'
    nak: {}
```

*api test demo application*

*Possible future state* 

* bin/api-service: provide the api endpoint and serve provisioning and status monitoring
* bin/api-deploy: client provision / deployment, more or less curl + some options

For demo purposes this is prebuilt with bin/* 

To test run

```
bin/api-service --debug --verbose
```

Switch terminal windows and run

```
make test-api-service
```

Will execute a group of json/yaml requests to do some.


The driver just performs direct executions of yaml formatted objects.

```
bin/driver --file cmd-docker.yaml
```

Whose content is a command to execute a docker busybox container built
to perform a null validation with an echo.

```
apiVersion: v0.1
kind: test
spec:
- name: docker-json
  protocol: cmd
  send:
    spec:
      arg:
      - run
      - --rm
      - busybox
      - /bin/echo
      - '{ "reply": "ok" }'
      env:
        name: value
      uri: file:///usr/bin/docker
  recv:
    ack:
      stdout: |
        { "reply": "ok" }
    nak: {}
```

The response is:

```
INFO: 2015/10/13 17:04:42 manage.go:89: ApiVersion [        ] \
    test [docker-json                     ] type [file] \
    status [    ok] rc [0] uri [file:///usr/bin/docker] \
    args ["run" "--rm" "busybox" "/bin/echo" "{ \"reply\": \"ok\" }"] expect [{ "reply": "ok" }]

```

From a yaml formatted file, converted to json and sent as the data
using curl to post it to the path /api/v1/validate


```

curl --request POST --data "$(yaml2json < cmd-docker.yaml)" http://localhost:9999/api/v1/validate

{
  "filename": "API Call from /api/v1/validate",
  "apiVersion": "v0.1",
  "kind": "response",
  "TimeStamp": "2015.10.13.17.14.19.-0500",
  "spec": {
    "status": [
      {
        "timestamp": "2015.10.13.17.14.20.-0500",
        "error": "<nil>",
        "exit": 0,
        "name": "docker-json",
        "text": "{ \"reply\": \"ok\" }\n"
      }
    ]
  }
}

```


Calling the api server responds with:

```
{"filename":"API Call from /api/v1/validate","apiVersion":"v0.1","kind":"response","TimeStamp":"2015.10.13.17.14.44.-0500","spec":{"status":[{"timestamp":"2015.10.13.17.14.44.-0500","error":"\u003cnil\u003e","exit":0,"name":"docker-json","text":"{ \"reply\": \"ok\" }\n"}]}}

```

```
{
  "filename": "API Call from /api/v1/validate",
  "apiVersion": "v0.1",
  "kind": "response",
  "TimeStamp": "2015.10.13.17.14.44.-0500",
  "spec": {
    "status": [
      {
        "timestamp": "2015.10.13.17.14.44.-0500",
        "error": "<nil>",
        "exit": 0,
        "name": "docker-json",
        "text": "{ \"reply\": \"ok\" }\n"
      }
    ]
  }
}
```

Build out api-service.go to drive calls via http://api:port/api/v1/validate/ POST.

Utility functions for json stream editing added with other utilities.

*https://stedolan.github.io/jq/download/*

Utility functions for json stream editing added with other utilities

        bin/jq

Add bin/{json2yaml,yaml2json,jsonutil,jq} tools for testing and general
utility use.

*https://stedolan.github.io/jq/download/*

*Utility driver sample bin/driver implementation to perform validation checks*

*Anticipated extensions*

```
Possible use cases:

Api service front end for sequential execution

Unified application validation system centralizing execution with
uniform api [ json call formats for suitably configured tooling. ]

```

*Perhaps in the future Api service to enable call of tests similar to via name /api/v1/validate/{{test-name}}*

```
  * http://localhost/api/v1/validate

```

```
make test-driver has some example execution of tests.
```

```
bin/driver --file check-echo.yaml

Info: 2015/10/09 15:11:44 driver.go:336: ApiVersion [    v0.1] test [tcp-test-string-nak             ] type [ tcp] status [    ok] rc [1] uri [tcp://localhost:7] args ["test string"] expect [test string nak]
Info: 2015/10/09 15:11:44 driver.go:336: ApiVersion [    v0.1] test [tcp-test-string-ack             ] type [ tcp] status [    ok] rc [0] uri [tcp://localhost:7] args ["test string ack"] expect [test string ack]
```

```
Info: 2015/10/09 15:11:44 driver.go:336: ApiVersion [    v0.1]\
    test [tcp-test-string-nak             ] type [ tcp] status [    ok] rc [1]\
        uri [tcp://localhost:7] args ["test string"] expect [test string nak]
Info: 2015/10/09 15:11:44 driver.go:336: ApiVersion [    v0.1]\
    test [tcp-test-string-ack             ] type [ tcp] status [    ok] rc [0] \
        uri [tcp://localhost:7] args ["test string ack"] expect [test string ack]

```


*Samples for validation in json and yaml syntax are included.*

*The default makefile is configured to use godeps for vendoring.*

```

make -k init get save build test-driver

GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep get "gopkg.in/yaml.v2" 
GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep save
GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep go build -a -ldflags '-s'

```

### consul test setup

* https://hub.docker.com/r/progrium/consul/

```
docker run -it -p 8400:8400 -p 8500:8500 -p 8600:53/udp oba11/consul -server -bootstrap
```
