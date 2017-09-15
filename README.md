# docker-ee-chargeback

This repo contains a utility tool to generate hourly reports of usage
against a Docker EE deployment, which can then be used to bill internal
departments for usage.

This tool exposes the `Collection` for the given resources, which can then
be mapped back to internal teams or billing departments.   We recommend
different billable departments not have their resources placed in the
same collections for this reason.

The key metrics exposed in this tool are:

* `CPU` - Cumulative CPU second used over the duration of the report (typically 1 hour)
* `Memory` - Min, max, and average MB of memory used by the container during the interval
* `Network` - RX/TX MB for each container
* `Container Storage` - Min, Max, and average MB of ephemeral usage within the container (excluding the image layers)
* `Volume` - Min, Max, and average MB used by storage volumes on the system



## Using published images from Docker Hub

Note: The metrics stored within EE vary across releases.  You should
use a version tag of this tool that matches the `major.minor` version
of your UCP.

1. Download an admin bundle from UCP on your EE cluster
2. Source the env script so your docker commands are sent through UCP

Run manually to verify you're able to get the CSV results.

```
MANAGER=$(docker node ls --filter role=manager --format '{{.Hostname}}' | head -1)
docker run --rm -e constraint:node==${MANAGER} -v ucp-node-certs:/certs \
    dhiltgen/chargeback:2.2 --certs /certs \
    --ucp $(echo $DOCKER_HOST | cut -f3 -d/ | cut -f1 -d:)
```

Once you've confirmed this looks good, you can set up a cron-job to run
this command hourly against your cluster.  You can append the results to
a single file and omit the header row with `--omit-header` added at the
end of the command.  Run `docker run --rm dhiltgen/chargeback:2.2 --help`
for more detailed usage information.


## Developing

**For Developers wishing to modify this tool**

If you're building against a UCP cluster, since the image will ultimately
need to run on a manager, you can use scheduler constraints to build on
a manager (not recommended for a production cluster!)

Once you've downloaded a bundle (preferably admin, since you'll need an
admin bundle to run the final image anyway) you can build it with:


```
MANAGER=$(docker node ls --filter role=manager --format '{{.Hostname}}' | head -1)
```


```
docker build --build-arg constraint:node==${MANAGER} -t chargeback:2.2 .
```

Then to run it:

```
docker run --rm -e constraint:node==${MANAGER} -v ucp-node-certs:/certs \
    chargeback:2.2 --certs /certs \
    --ucp $(echo $DOCKER_HOST | cut -f3 -d/ | cut -f1 -d:)

```
