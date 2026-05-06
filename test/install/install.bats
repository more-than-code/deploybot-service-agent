#!/usr/bin/env bats

setup_file() {
  export REPO_ROOT
  REPO_ROOT=$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)
  export CONTAINER_CLI="${CONTAINER_CLI:-docker}"
  export IMAGE_NAME="${IMAGE_NAME:-install-test}"

  bash "$REPO_ROOT/test/install/run-tests.sh" --build-only
}

@test "lists available install scenarios" {
  run bash "$REPO_ROOT/test/install/run-tests.sh" --list

  [ "$status" -eq 0 ]
  [[ "$output" == *"fresh-install"* ]]
  [[ "$output" == *"checksum-mismatch"* ]]
}

@test "install.sh docker test matrix passes" {
  run env SKIP_BUILD=1 CONTAINER_CLI="$CONTAINER_CLI" IMAGE_NAME="$IMAGE_NAME" \
    bash "$REPO_ROOT/test/install/run-tests.sh"

  [ "$status" -eq 0 ]
  [[ "$output" == *"All requested install.sh scenarios passed."* ]]
}
