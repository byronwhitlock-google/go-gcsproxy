#!/bin/bash
#

if [[ -z "$CA_BUNDLE" ]]; then
  echo "Error: CA_BUNDLE environment variable is not set. eg: /Users/<USERNAME>/certs/mitmproxy-ca.pem" >&2
  exit 1
fi

if [[ -z "$BUCKET" ]]; then
  echo "Error: BUCKET environment variable is not set." >&2
  exit 1
fi

if ! command -v bats &> /dev/null; then
  echo "BATS is not installed. To install run:"

# Install BATS based on your system's package manager
  echo "Debian/Ubuntu"
  echo "  sudo apt update"
  echo "  sudo apt install -y bats"
  echo "macOS with Homebrew"
  echo "  brew install bats-core"
  echo "  brew install bats-support"
  exit 1
fi

# global_setup
export HTTPS_PROXY=http://localhost:9080
export REQUESTS_CA_BUNDLE=$CA_BUNDLE
gcloud config set core/custom_ca_certs_file $REQUESTS_CA_BUNDLE

# Trap EXIT signal to call global_teardown()
global_teardown() { 
  unset HTTPS_PROXY
  unset REQUESTS_CA_BUNDLE
  gcloud config unset core/custom_ca_certs_file 
}
# Trap EXIT signal to call global_teardown()
trap global_teardown EXIT

# Check if a parameter was provided
if [ -n "$1" ]; then
  # Parameter provided, run the specific test file
  test_file="tests/$1"
  if [ -f "$test_file" ]; then
    bats "$test_file"
  else
    echo "Error: Test file '$test_file' not found."
    exit 1
  fi
else
  # No parameter, run all tests in the 'tests' directory
  
  bats ./tests/*.bats
fi





