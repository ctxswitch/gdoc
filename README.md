# Godoc Service For Private Github Repositories.

The Godoc web service provides a way to publish godocs for private Github repositories.  It uses Github topics as a way to identify which repositories under a user or organization to publish locally.  By default it uses the `godoc` topic but is configurable to any topic value.

Several configuration parameters are available for controlling the behavior of the service.  They are defined through environment variables and include:

* `GITHUB_TOKEN`: A personal access token with permissions to access and list the repositories.  **Required**
* `GITHUB_USER`: The Github user or organization that will be scraped.  Only single values are currently supported. **Required**
* `GITHUB_TOKEN_USER`: If the user that owns the personal access token is different than the owner or the repositories are part of an organization, specify the token user.  Defaults to the `GITHUB_USER`.
* `GITHUB_POLL_INTERVAL`: The interval to check for changes on Github.  Takes a duration string for the value.  The string is an unsigned decimal number(s), with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".  Default is `5m`.
* `GITHUB_TOPIC`: The topic that will be used as a filter to identify repositories that will be synchronized.  Default is `godoc`
* `GODOC_PORT`: The port that godoc will run on. Default is `6060`.
* `GODOC_ROOT`: The workspace root that will be passed to godoc.  This is also the root of where your repositories will be cloned and updated.  Default is `/usr/local/go`.
* `GODOC_INDEX_INTERVAL`: The indexing interval for godoc.  0 for the godoc default (5m), negative to only index once at startup.  Default for this service is `1m`
* `LOG_LEVEL`: Changes the verbosity of the logging service.  Default is `INFO`.

This is a basic service that does not provide any coordination in terms of repository synchronization.  As such, scaling this out for availability reasons could be impactful on your API limits.  In the future, the possibility of shared object storage and leader elections could solve this, but these features have not yet been planned.

## Install and Run

The recommended way to run is using docker.  You can build and run the container directly or run the prebuilt container.

```
docker run --rm -it -e GITHUB_TOKEN=$GITHUB_TOKEN -e GITHUB_USER=$GITHUB_USER -p 6060:6060 ctxsh/gdoc
```

Browse to your Github account and add a topic tag of `godoc` to the repositories that you would like the service to discover.  Once the topic tag has been added, the service will pick up the new repository on it's next Github poll.  The new information will be available after the next index cycle has completed and the broswer page has been refreshed.


Direct your browser to http://localhost:6060 and browse your go documentation.
