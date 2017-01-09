# Mainflux Core

[![License](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)
[![Build Status](https://travis-ci.org/mainflux/mainflux-core.svg?branch=master)](https://travis-ci.org/mainflux/mainflux-core)
[![Go Report Card](https://goreportcard.com/badge/github.com/mainflux/mainflux-core)](https://goreportcard.com/report/github.com/mainflux/mainflux-core)
[![Join the chat at https://gitter.im/Mainflux/mainflux](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/Mainflux/mainflux?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Mainflux Core is core service of [Mainflux IoT Cloud Platform](https://github.com/mainflux/mainflux).

It exposes an API for system provisioning and management.

### Security
For professional deployments Mainflux is usually combined with [Mainflux Authentication and Authorization Server](https://github.com/mainflux/mainflux-auth-server) which adds fine-grained security based on customizable API keys.

Mainflux Auth Server also provides user accounts and device and application access control with simple customizable scheme based on scoped JWTs.

### Install/Deploy
Mainflux Core uses [MongoDB](https://www.mongodb.com/), so insure that it is installed on your system. You will also need to run [NATS](https://github.com/nats-io/gnatsd). Finally, [Mainflux MQTT broker](https://github.com/mainflux/mainflux-mqtt) needs to be running also.

Installing Mainflux is trivial [`go get`](https://golang.org/cmd/go/):
```bash
go get github.com/mainflux/mainflux-core
$GOBIN/mainflux-core
```

If you are new to Go, more information about setting-up environment and fetching Mainflux code can be found [here](https://github.com/mainflux/mainflux-core-doc/blob/master/goenv.md).

### Documentation
Development documentation can be found [here](http://mainflux.io).

### Community
#### Mailing list
[mainflux](https://groups.google.com/forum/#!forum/mainflux) Google group

For quick questions and suggestions you can also use GitHub Issues.

#### IRC
[Mainflux Gitter](https://gitter.im/Mainflux/mainflux?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

#### Twitter
[@mainflux](https://twitter.com/mainflux)

### Authors
Main architect and BDFL of Mainflux project is [@drasko](https://github.com/drasko). Additionaly, initial version of Mainflux was architectured and crafted by [@janko-isidorovic](https://github.com/janko-isidorovic), [@nmarcetic](https://github.com/nmarcetic) and [@mijicd](https://github.com/mijicd).

Maintainers are listed in [MAINTAINERS](MAINTAINERS) file.

Contributors are listed in [CONTRIBUTORS](CONTRIBUTORS) file.

### License
[Apache License, version 2.0](LICENSE)
