DockerTorch
===========

Docker Torch builds flame graphs from pprof output from a docker daemon.
It uses github.com/uber/go-torch to read pprof data into a flame graph and
github.com/crosbymichael/docker-stress to provide workload against the Docker
daemon to measure.


### Usage

```
$ cat stress-config.json | docker run -i --rm cpuguy83/docker-torch -H tcp://<host>:<port> --concurrency 8 --containers 2000 -t 60 > torch.svg
```

That will create 2000 containers with a concurrency of 8 while reading the
pprof output for 60 seconds output an SVG to stdout. If `-t` is reached it will
stop creating containers.


```
Usage of /opt/flame-on/flame-on:
  -concurrency string
    	max number of concurrent operations to Docker (default "1")
  -containers string
    	number of containers to run (default "1000")
  -dockerversion string
    	specify the version of Docker client to use (default "latest")
  -host string
    	host:port to use to connect to docker
  -kill string
    	time to kill a container after it is executed (default "10s")
  -time string
    	amount of time to collect data (default "30")
```


Docker Torch does not spin up a Docker daemon for you, you must do this and make
sure to listen on a TCP port without TLS.
