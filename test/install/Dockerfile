FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive
ENV SUDO_USER=testuser

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        coreutils \
        curl \
        findutils \
        grep \
        passwd \
        sudo \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --shell /bin/bash testuser \
    && echo 'testuser ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/testuser \
    && chmod 0440 /etc/sudoers.d/testuser

RUN mkdir -p /opt/install-test/bin /opt/install-test/fixtures

RUN cat <<'EOF' > /opt/install-test/fixtures/bot_agent
#!/bin/sh
echo "bot_agent test fixture"
EOF
RUN chmod 0755 /opt/install-test/fixtures/bot_agent

RUN cat <<'EOF' > /opt/install-test/bin/curl
#!/bin/bash
set -euo pipefail

REAL_CURL=/usr/bin/curl
FIXTURE=/opt/install-test/fixtures/bot_agent
OUTFILE=
URL=
EXPECT_OUTFILE=0
ORIGINAL_ARGS=("$@")

while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      EXPECT_OUTFILE=1
      ;;
    --output=*)
      OUTFILE="${1#--output=}"
      ;;
    -*)
      ;;
    *)
      if [ "$EXPECT_OUTFILE" = "1" ]; then
        OUTFILE="$1"
        EXPECT_OUTFILE=0
      else
        URL="$1"
      fi
      ;;
  esac
  shift
done

if [ -n "${URL:-}" ] && [ -n "${OUTFILE:-}" ]; then
  case "$URL" in
    https://github.com/more-than-code/deploybot-service-agent/releases/download/*/bot_agent-linux-*.sha256)
      checksum="$(sha256sum "$FIXTURE" | cut -d' ' -f1)"
      if [ "${INSTALL_TEST_BAD_CHECKSUM:-0}" = "1" ]; then
        checksum="0000000000000000000000000000000000000000000000000000000000000000"
      fi
      printf '%s  %s\n' "$checksum" "$(basename "$URL" .sha256)" > "$OUTFILE"
      exit 0
      ;;
    https://github.com/more-than-code/deploybot-service-agent/releases/download/*/bot_agent-linux-*)
      cp "$FIXTURE" "$OUTFILE"
      exit 0
      ;;
  esac
fi

exec "$REAL_CURL" "${ORIGINAL_ARGS[@]}"
EOF
RUN chmod 0755 /opt/install-test/bin/curl

RUN cat <<'EOF' > /opt/install-test/bin/systemctl
#!/bin/sh
set -eu

SERVICE_FILE=/var/run/install-test-systemctl-bot_agent-active

case "${1:-}" in
  --version)
    echo "systemd 252 (install-test shim)"
    exit 0
    ;;
  is-active)
    if [ "${2:-}" = "--quiet" ] && [ "${3:-}" = "bot_agent" ]; then
      if [ "${INSTALL_TEST_SERVICE_ACTIVE:-0}" = "1" ] || [ -f "$SERVICE_FILE" ]; then
        exit 0
      fi
      exit 3
    fi
    ;;
  daemon-reload)
    exit 0
    ;;
  restart)
    if [ "${2:-}" = "bot_agent" ]; then
      if [ "${INSTALL_TEST_RESTART_FAIL:-0}" = "1" ]; then
        exit 1
      fi
      touch "$SERVICE_FILE"
      exit 0
    fi
    ;;
  enable|start|status)
    exit 0
    ;;
esac

echo "install-test systemctl shim: unsupported command: $*" >&2
exit 1
EOF
RUN chmod 0755 /opt/install-test/bin/systemctl

RUN cat <<'EOF' > /usr/bin/docker
#!/bin/sh
echo "install-test docker shim"
exit 0
EOF
RUN chmod 0755 /usr/bin/docker

ENV PATH="/opt/install-test/bin:${PATH}"

COPY install.sh /tmp/install.sh

RUN chmod 0755 /tmp/install.sh

WORKDIR /workspace

CMD ["/bin/bash"]
