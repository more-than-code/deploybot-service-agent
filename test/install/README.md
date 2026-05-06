# install.sh Test Plan

## Static Analysis

```bash
bash -n install.sh          # syntax check
shellcheck install.sh       # lint (brew install shellcheck)
```

## Docker Integration Tests

Build a container that mimics a target host, then exercise each scenario without touching real infrastructure.

### Setup

```bash
docker build -f test/install/Dockerfile -t install-test .
```

The test image already stubs the GitHub release download path, so runs stay local after the image build. If you use Podman on macOS behind a Docker-compatible wrapper, the same commands work. To call Podman explicitly:

```bash
CONTAINER_CLI=podman bash test/install/run-tests.sh
```

### Run the full matrix

```bash
bash test/install/run-tests.sh
```

### Run one scenario

```bash
bash test/install/run-tests.sh fresh-install
```

### Test Matrix

| # | Scenario | Command | Expected |
|---|----------|---------|----------|
| 1 | No root | `bash test/install/run-tests.sh no-root` | Exit 0 from harness; scenario asserts root check |
| 2 | Missing Docker | `bash test/install/run-tests.sh missing-docker` | Exit 0 from harness; scenario asserts Docker preflight failure |
| 3 | TARGET_USER=root | `bash test/install/run-tests.sh target-user-root` | Exit 0 from harness; scenario asserts non-root user guard |
| 4 | Invalid username | `bash test/install/run-tests.sh invalid-username` | Exit 0 from harness; scenario asserts username validation |
| 5 | Non-existent user | `bash test/install/run-tests.sh nonexistent-user` | Exit 0 from harness; scenario asserts getent failure |
| 6 | Fresh install | `bash test/install/run-tests.sh fresh-install` | Binary, env file, and service file are created |
| 7 | Upgrade (env exists) | `bash test/install/run-tests.sh upgrade-env-merge` | Merges defaults and preserves existing values |
| 8 | Checksum mismatch | `bash test/install/run-tests.sh checksum-mismatch` | Detects mismatch and restores previous binary |
| 9 | No /etc/letsencrypt | `bash test/install/run-tests.sh no-letsencrypt` | No `ReadOnlyPaths`, no `ssl-cert` in unit |
| 10 | With /etc/letsencrypt | `bash test/install/run-tests.sh with-letsencrypt` | `ssl-cert` group exists and `ReadOnlyPaths` is present |
| 11 | Service running | `bash test/install/run-tests.sh service-running` | Restart path runs when service is already active |

## BATS Automated Tests

Install: `brew install bats-core` (macOS) or `apt-get install bats` (Linux).

```bash
bats test/install/install.bats
```

The BATS suite builds the image once in `setup_file`, then reuses `test/install/run-tests.sh` for the checks.
