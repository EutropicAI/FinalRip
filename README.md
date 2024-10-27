# FinalRip

a distributed video processing tool, based on FFmpeg and VapourSynth

[![CI-test](https://github.com/TensoRaws/FinalRip/actions/workflows/CI-test.yml/badge.svg)](https://github.com/TensoRaws/FinalRip/actions/workflows/CI-test.yml)
[![golangci-lint](https://github.com/TensoRaws/FinalRip/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/TensoRaws/FinalRip/actions/workflows/golangci-lint.yml)
[![Docker Build CI](https://github.com/TensoRaws/FinalRip/actions/workflows/CI-docker.yml/badge.svg)](https://github.com/TensoRaws/FinalRip/actions/workflows/CI-docker.yml)
[![Release](https://github.com/TensoRaws/FinalRip/actions/workflows/Release.yml/badge.svg)](https://github.com/TensoRaws/FinalRip/actions/workflows/Release.yml)
[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/RJWBNXdmdaDACvcacXFQ3e/Ge3dVaX4GmktGiL9Jb1ADB/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/circleci/RJWBNXdmdaDACvcacXFQ3e/Ge3dVaX4GmktGiL9Jb1ADB/tree/main)

### Architecture

![FinalRip](https://raw.githubusercontent.com/TensoRaws/.github/refs/heads/main/finalrip.png)

_We cut the original video into multiple clips, and then process each clip in parallel in queue order. After all clips are processed, we merge them into the final video._

### Preparations

- docker and docker-compose
- Nvidia GPU / AMD GPU with ROCm support [(7000 series on WSL2)](https://github.com/TensoRaws/vs-playground/blob/main/docker-compose-rocm.yml)
- GPU container runtime (optional)
- make (optional)

### Quick Deployment

edit [Line 19](https://github.com/TensoRaws/FinalRip/blob/main/deploy/docker-compose/lite/docker-compose.yml#L19) to your LAN IP address

```bash
docker-compose -f deploy/docker-compose/lite/docker-compose.yml up -d
```

It will run all containers in a single host, then open `http://localhost:8989` in your browser to access the dashboard, open `http://localhost:8080` to access the Asynq monitor.

### Start

![Dashboard](https://raw.githubusercontent.com/TensoRaws/.github/refs/heads/main/finalrip.gif)

We use [this container](https://github.com/TensoRaws/vs-playground) as the base image, which contains FFmpeg, VapourSynth, PyTorch...

So in the dashboard, we can select a compatible script to process the video!

### Distributed Deployment

Deploy the system in a distributed way, you can refer to the [docker-compose](./deploy/docker-compose) directory for more details.

first, run docker-compose-base.yml to start the basic services, and open Consul dashboard, add a K/V pair with key `finalrip.yml` and value is the content of [finalrip.yml](./conf/finalrip.yml) -- or your own configuration file

then, run docker-compose-server.yml to start the dashboard, server, cut worker, and merge worker services

finally, run docker-compose-encode.yml to start the encode worker services, we can deploy multiple encode workers in different hosts to speed up the encoding process

_Note: we suggest that deploy oss service, cut & merge worker in the same host_

### Configuration

Override the default configuration by setting:

#### Environment variables >> Config File / Remote Config File (Consul)

Special Env Variables:

- `FINALRIP_REMOTE_CONFIG_HOST` Consul host, default is None, that means load config from local file. When set, it will load the config from the Consul K/V store. When set to `EASYTIER` / `easytier`, will try load config from `FINALRIP_EASYTIER_HOST:8500` (`10.126.126.251:8500` by default).
- `FINALRIP_REMOTE_CONFIG_KEY` Consul key, default is `finalrip.yml`

### A new script?

In [vs-playground](https://github.com/TensoRaws/vs-playground), we provide the same environment as the encode worker, so you can develop and test your script in the playground.

### For Advanced User

- API document: [here](https://apifox.com/apidoc/shared-0b6425d8-0140-4822-9f59-f1d6d7784b03)
- Build your own `encode` image: refer to the [vs-playground](https://github.com/TensoRaws/vs-playground), and set the `Template Repo` name in dashboard if you wanna select a script from the repo's `templates` folder.

### Build

```bash
make all
make pt
```

`make pt-rocm` for AMD GPU

### Reference

- [asynq](https://github.com/hibiken/asynq)
- [gin](https://github.com/gin-gonic/gin)
- [FFmpeg](https://github.com/FFmpeg/FFmpeg)
- [VapourSynth](https://github.com/vapoursynth/vapoursynth)

### License

This project is licensed under the GPL-3.0 license - see the [LICENSE file](https://github.com/TensoRaws/FinalRip/blob/main/LICENSE) for details.
