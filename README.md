[![GoDoc](https://godoc.org/github.com/lindell/revaboxy/pkg/revaboxy?status.svg)](https://godoc.org/github.com/lindell/revaboxy/pkg/revaboxy)

Revaboxy
----
Revaboxy is a reverse proxy made solely for A/B testing of front end applications.
It is placed in front of two or more versions of a frontend and does randomize the trafic to the different versions based on probability.
When a users browser makes subsequent requests, revaboxy will automaticly select the same version as before.


Revaboxy is released as [docker images](https://hub.docker.com/r/lindell/revaboxy/tags), [binaries for linux/windows/mac](https://github.com/lindell/revaboxy/releases) and as a [Go library](https://godoc.org/github.com/lindell/revaboxy/pkg/revaboxy).

Environment Variables
----

#### Configuring the versions
When setting up Revaboxy, every version has to be setup with the url and probability that it will be selected.
These environment variables are called `VERSION_NAME_URL` and `VERSION_NAME_PROBABILITY`.

As an example, say we have two version, one called `DEFAULT` and one called `GREEN_BACKGROUND`. The environment variables needed would be:

```bash
VERSION_DEFAULT_URL=http://defaulturl
VERSION_DEFAULT_PROBABILITY=0.6
VERSION_GREEN_BACKGROUND_URL=http://greenbackgroundurl
VERSION_GREEN_BACKGROUND_PROBABILITY=0.4
```

#### Setting to change the behavior of revaboxy
| Name | Default | Description |
| ---- | ------- | ----------- |
| HEADER_NAME | revaboxy-name | The header name sent to the downsteam application |
| COOKIE_NAME | revaboxy-name | The cookie name that is set at the client to keep track of which version was selected |

