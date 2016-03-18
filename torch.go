package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

var (
	flConcurrency   = flag.String("concurrency", "1", "max number of concurrent operations to Docker")
	flDockerVersion = flag.String("dockerversion", "latest", "specify the version of Docker client to use")
	flContainers    = flag.String("containers", "1000", "number of containers to run")
	flKill          = flag.String("kill", "10s", "time to kill a container after it is executed")
	flTime          = flag.String("time", "30", "amount of time to collect data")
	flHost          = flag.String("host", "", "host:port to use to connect to docker")
)

func main() {
	flag.Parse()
	if len(*flHost) == 0 {
		fmt.Fprintln(os.Stderr, "Missing flag `--host` with host:port to connect to")
		os.Exit(1)
	}

	dockerPath, err := exec.LookPath("docker")

	if err != nil {
		resp, err := http.Get("https://get.docker.com/builds/Linux/x86_64/docker-" + *flDockerVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error downloading docker version: %v\n", err)
			os.Exit(1)
		}

		dockerPath = "/usr/bin/docker"
		dockerF, err := os.Create(dockerPath)
		if err := dockerF.Chmod(777); err != nil {
			fmt.Fprintf(os.Stderr, "error downloading docker version: %v\n", err)
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error downloading docker version: %v\n", err)
			os.Exit(1)
		}
		if _, err := io.Copy(dockerF, resp.Body); err != nil {
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "error downloading docker version: %v\n", err)
			os.Exit(1)
		}
		resp.Body.Close()
		dockerF.Close()
	}

	stressCfg, err := ioutil.TempFile("", "stress.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating stress config: %v\n", err)
		os.Exit(1)
	}

	_, err = io.Copy(stressCfg, os.Stdin)
	stressCfg.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating stress config: %v\n", err)
		os.Exit(1)
	}

	torch, err := exec.LookPath("go-torch")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find go-torch bin: %v\n", err)
		os.Exit(1)
	}

	torchCmd := exec.Command(torch, "-t", *flTime, "--url", "http://"+*flHost, "--print")
	torchCmd.Stdout = os.Stdout
	torchCmd.Stderr = bytes.NewBuffer(nil)

	stress, err := exec.LookPath("docker-stress")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find docker-stress bin: %v\n", err)
		os.Exit(1)
	}
	stressCmd := exec.Command(stress, "-c", *flConcurrency, "--kill", *flKill, "--config", stressCfg.Name(), "--binary", dockerPath)
	stressCmd.Stderr = bytes.NewBuffer(nil)
	stressCmd.Stdout = stressCmd.Stderr
	stressCmd.Env = append(stressCmd.Env, "DOCKER_HOST=tcp://"+*flHost)

	stressWait := make(chan error, 1)
	go func() {
		stressWait <- stressCmd.Run()
	}()

	torchWait := make(chan error, 1)
	go func() {
		torchWait <- torchCmd.Run()
	}()

	exit := 0
	select {
	case err := <-torchWait:
		stressCmd.Process.Kill()
		if err != nil {
			exit = 1
		}
		<-stressWait
	case err := <-stressWait:
		if err != nil {
			torchCmd.Process.Kill()
			exit = 1
		}
		<-torchWait
	}

	io.Copy(os.Stderr, stressCmd.Stderr.(*bytes.Buffer))
	io.Copy(os.Stderr, torchCmd.Stderr.(*bytes.Buffer))
	os.Exit(exit)
}
