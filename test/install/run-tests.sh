#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
CONTAINER_CLI=${CONTAINER_CLI:-docker}
IMAGE_NAME=${IMAGE_NAME:-install-test}
DOCKERFILE_PATH=${DOCKERFILE_PATH:-"$REPO_ROOT/test/install/Dockerfile"}
BUILD_CONTEXT=${BUILD_CONTEXT:-"$REPO_ROOT"}
SKIP_BUILD=${SKIP_BUILD:-0}

SCENARIOS=(
  no-root
  missing-docker
  target-user-root
  invalid-username
  nonexistent-user
  fresh-install
  upgrade-env-merge
  checksum-mismatch
  no-letsencrypt
  with-letsencrypt
  service-running
)

usage() {
  cat <<EOF
Usage: $(basename "$0") [--skip-build] [--build-only] [--list] [scenario ...]

Runs the install.sh Docker integration scenarios using $CONTAINER_CLI.

Environment overrides:
  CONTAINER_CLI   Container CLI to use (default: docker)
  IMAGE_NAME      Test image tag (default: install-test)
  DOCKERFILE_PATH Dockerfile path (default: test/install/Dockerfile)
  BUILD_CONTEXT   Build context (default: repo root)
EOF
}

list_scenarios() {
  printf '%s\n' "${SCENARIOS[@]}"
}

ensure_container_cli() {
  if ! command -v "$CONTAINER_CLI" >/dev/null 2>&1; then
    echo "Error: container CLI '$CONTAINER_CLI' not found." >&2
    exit 1
  fi
}

build_image() {
  echo "==> Building test image '$IMAGE_NAME' with $CONTAINER_CLI"
  "$CONTAINER_CLI" build -f "$DOCKERFILE_PATH" -t "$IMAGE_NAME" "$BUILD_CONTEXT"
}

run_container() {
  local command=$1
  shift

  local -a args
  args=("$CONTAINER_CLI" run --rm)

  while [ "$#" -gt 0 ]; do
    args+=(-e "$1")
    shift
  done

  args+=("$IMAGE_NAME" bash -lc "$command")

  "${args[@]}"
}

assert_case() {
  local name=$1
  local command=$2
  local expected_exit=$3
  shift 3

  local -a env_vars=()
  local -a patterns=()
  local mode=patterns

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --env)
        mode=env
        ;;
      --patterns)
        mode=patterns
        ;;
      *)
        if [ "$mode" = env ]; then
          env_vars+=("$1")
        else
          patterns+=("$1")
        fi
        ;;
    esac
    shift
  done

  echo "==> [$name]"

  local output
  local status

  set +e
  if [ "${#env_vars[@]}" -gt 0 ]; then
    output=$(run_container "$command" "${env_vars[@]}" 2>&1)
  else
    output=$(run_container "$command" 2>&1)
  fi
  status=$?
  set -e

  if [ "$status" -ne "$expected_exit" ]; then
    echo "FAIL [$name] expected exit $expected_exit, got $status" >&2
    printf '%s\n' "$output" >&2
    return 1
  fi

  local pattern
  for pattern in "${patterns[@]}"; do
    if ! grep -Fq "$pattern" <<<"$output"; then
      echo "FAIL [$name] missing output: $pattern" >&2
      printf '%s\n' "$output" >&2
      return 1
    fi
  done

  echo "PASS [$name]"
}

scenario_no_root() {
  assert_case \
    "no-root" \
    'sudo -u testuser bash /tmp/install.sh' \
    1 \
    --patterns \
    'This script must be run as root (use sudo)'
}

scenario_missing_docker() {
  assert_case \
    "missing-docker" \
    'rm -f /usr/bin/docker && bash /tmp/install.sh' \
    1 \
    --patterns \
    'Docker is not installed'
}

scenario_target_user_root() {
  assert_case \
    "target-user-root" \
    'bash /tmp/install.sh' \
    1 \
    --env \
    'TARGET_USER=root' \
    --patterns \
    'Could not determine a non-root user'
}

scenario_invalid_username() {
  assert_case \
    "invalid-username" \
    'bash /tmp/install.sh' \
    1 \
    --env \
    'TARGET_USER=bad!!' \
    --patterns \
    "Invalid username 'bad!!'"
}

scenario_nonexistent_user() {
  assert_case \
    "nonexistent-user" \
    'bash /tmp/install.sh' \
    1 \
    --env \
    'TARGET_USER=nobody99' \
    --patterns \
    "Unable to determine installation user 'nobody99'"
}

scenario_fresh_install() {
  assert_case \
    "fresh-install" \
    'rm -rf /etc/letsencrypt /home/testuser/.bot_agent /etc/systemd/system/bot_agent.service /usr/local/bin/bot_agent && bash /tmp/install.sh latest && test -x /usr/local/bin/bot_agent && test -f /home/testuser/.bot_agent/env.conf && test -f /etc/systemd/system/bot_agent.service && grep -Fxq "SERVICE_PORT=:8081" /home/testuser/.bot_agent/env.conf && grep -Fxq "User=testuser" /etc/systemd/system/bot_agent.service && ! grep -Fxq "ReadOnlyPaths=/etc/letsencrypt" /etc/systemd/system/bot_agent.service' \
    0 \
    --patterns \
    'Bot Agent Installation Complete'
}

scenario_upgrade_env_merge() {
  assert_case \
    "upgrade-env-merge" \
    'rm -rf /home/testuser/.bot_agent /etc/systemd/system/bot_agent.service /usr/local/bin/bot_agent && mkdir -p /home/testuser/.bot_agent && cat <<'"'"'EOF'"'"' > /home/testuser/.bot_agent/env.conf
SERVICE_PORT=:9999
API_KEY=custom_key
EOF
chown -R testuser:testuser /home/testuser/.bot_agent
bash /tmp/install.sh latest
grep -Fxq "SERVICE_PORT=:9999" /home/testuser/.bot_agent/env.conf
grep -Fxq "API_KEY=custom_key" /home/testuser/.bot_agent/env.conf
grep -Fxq "DOCKER_HOST=unix:///var/run/docker.sock" /home/testuser/.bot_agent/env.conf' \
    0 \
    --patterns \
    'Environment file updated with new configuration options'
}

scenario_checksum_mismatch() {
  assert_case \
    "checksum-mismatch" \
    'printf "old-binary\n" > /usr/local/bin/bot_agent && chmod 0755 /usr/local/bin/bot_agent
set +e
bash /tmp/install.sh latest
status=$?
set -e
[ "$status" -eq 1 ]
grep -Fxq "old-binary" /usr/local/bin/bot_agent' \
    0 \
    --env \
    'INSTALL_TEST_BAD_CHECKSUM=1' \
    --patterns \
    'Binary integrity verification failed!'
}

scenario_no_letsencrypt() {
  assert_case \
    "no-letsencrypt" \
    'rm -rf /etc/letsencrypt /etc/systemd/system/bot_agent.service /usr/local/bin/bot_agent /home/testuser/.bot_agent && bash /tmp/install.sh latest && ! grep -Fxq "ReadOnlyPaths=/etc/letsencrypt" /etc/systemd/system/bot_agent.service && ! grep -Fq "ssl-cert" /etc/systemd/system/bot_agent.service' \
    0 \
    --patterns \
    '/etc/letsencrypt not found - certbot may not be installed'
}

scenario_with_letsencrypt() {
  assert_case \
    "with-letsencrypt" \
    'rm -rf /etc/letsencrypt /etc/systemd/system/bot_agent.service /usr/local/bin/bot_agent /home/testuser/.bot_agent
mkdir -p /etc/letsencrypt/live/example.test
printf "cert\n" > /etc/letsencrypt/live/example.test/fullchain.pem
printf "key\n" > /etc/letsencrypt/live/example.test/privkey.pem
bash /tmp/install.sh latest
grep -Fxq "ReadOnlyPaths=/etc/letsencrypt" /etc/systemd/system/bot_agent.service
grep -Fq "SupplementaryGroups=docker ssl-cert" /etc/systemd/system/bot_agent.service
getent group ssl-cert | grep -Fq testuser' \
    0 \
    --patterns \
    'Certificate access configured: group=ssl-cert'
}

scenario_service_running() {
  assert_case \
    "service-running" \
    'rm -rf /etc/systemd/system/bot_agent.service /usr/local/bin/bot_agent /home/testuser/.bot_agent && bash /tmp/install.sh latest && test -f /var/run/install-test-systemctl-bot_agent-active' \
    0 \
    --env \
    'INSTALL_TEST_SERVICE_ACTIVE=1' \
    --patterns \
    'Restarting bot_agent service...' \
    'Service restarted successfully'
}

run_named_scenario() {
  case "$1" in
    no-root) scenario_no_root ;;
    missing-docker) scenario_missing_docker ;;
    target-user-root) scenario_target_user_root ;;
    invalid-username) scenario_invalid_username ;;
    nonexistent-user) scenario_nonexistent_user ;;
    fresh-install) scenario_fresh_install ;;
    upgrade-env-merge) scenario_upgrade_env_merge ;;
    checksum-mismatch) scenario_checksum_mismatch ;;
    no-letsencrypt) scenario_no_letsencrypt ;;
    with-letsencrypt) scenario_with_letsencrypt ;;
    service-running) scenario_service_running ;;
    *)
      echo "Error: unknown scenario '$1'" >&2
      exit 1
      ;;
  esac
}

main() {
  local build_only=0
  local -a requested_scenarios=()

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --skip-build)
        SKIP_BUILD=1
        ;;
      --build-only)
        build_only=1
        ;;
      --list)
        list_scenarios
        exit 0
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        requested_scenarios+=("$1")
        ;;
    esac
    shift
  done

  ensure_container_cli

  if [ "$SKIP_BUILD" != "1" ]; then
    build_image
  fi

  if [ "$build_only" = "1" ]; then
    exit 0
  fi

  if [ "${#requested_scenarios[@]}" -eq 0 ]; then
    requested_scenarios=("${SCENARIOS[@]}")
  fi

  local scenario
  for scenario in "${requested_scenarios[@]}"; do
    run_named_scenario "$scenario"
  done

  echo "All requested install.sh scenarios passed."
}

main "$@"
