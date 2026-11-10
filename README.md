# nodecheck

![ci](https://github.com/notdinim/nodecheck/actions/workflows/ci.yml/badge.svg)

A small command line tool that gives you a quick health report for a Linux machine: CPU load, memory, disk space and NVIDIA GPUs.

```
$ nodecheck
Host:    DESKTOP-CN15JPM
Load:    0.12 (1 min) on 24 CPUs
Memory:  14331 MB available of 15913 MB
Disk /:  1018.8 GB free of 1081.1 GB
GPU 0:   NVIDIA GeForce RTX 4070 Ti, 12282 MiB
Status:  OK
```

I built this to learn how Linux infrastructure software actually gets shipped. The tool itself is simple on purpose. The interesting part is everything around it: Debian and RPM packaging, a signed apt repository, CI/CD with GitHub Actions, containers, and running it on every node of a Kubernetes cluster.

## Install

The latest release is **v0.1.0**.

### With apt (Ubuntu and Debian)

```bash
curl -fsSL https://notdinim.github.io/apt/KEY.gpg | gpg --dearmor | sudo tee /usr/share/keyrings/notdinim.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/notdinim.gpg] https://notdinim.github.io/apt ./" | sudo tee /etc/apt/sources.list.d/notdinim.list
sudo apt update && sudo apt install nodecheck
```

Updates come through `sudo apt upgrade` like any other package.

### Other options

Grab a `.deb`, `.rpm` or `.tar.gz` for amd64 or arm64 from the [Releases](https://github.com/notdinim/nodecheck/releases) page.

## Usage

| Command               | What it does                           |
| --------------------- | -------------------------------------- |
| `nodecheck`           | Prints the health report               |
| `nodecheck --json`    | Same report as JSON, handy for scripts |
| `nodecheck --version` | Shows the installed version            |

The exit code is `0` when everything looks fine and `1` when there is at least one warning, so you can use it in scripts and cron jobs. It warns when:

- load is higher than the number of CPUs
- less than 10% of memory is available
- less than 10% of the root disk is free

## How it works

nodecheck has no dependencies. It reads straight from Linux:

- **Load** from `/proc/loadavg`
- **Memory** from `/proc/meminfo` (it uses `MemAvailable`, which accounts for file cache)
- **Disk** from the kernel's filesystem stats for `/`
- **GPUs** from `nvidia-smi`, if it's installed

## How it's built and shipped

```
  git push   -->  GitHub Actions (ci)       -->  go vet + go test

  git tag v* -->  GitHub Actions (release)  -->  GoReleaser
                                                     |
                               +---------------------+---------------------+
                               |                     |                     |
                              .deb                  .rpm               .tar.gz
                          (amd64, arm64)        (amd64, arm64)      + checksums
                               |
                               v
                apt repo on GitHub Pages, signed with my GPG key
                               |
                               v
                     sudo apt install nodecheck


  Dockerfile  -->  container image  -->  docker run  or  Kubernetes DaemonSet
                                                        (one pod on every node)
```

- **GoReleaser** builds static binaries (`CGO_ENABLED=0`) for both architectures and packages them with nFPM.
- **The apt repo** lives in [notdinim/apt](https://github.com/notdinim/apt). Its index is signed with GPG, so apt can verify every package came from me and wasn't changed.

#
