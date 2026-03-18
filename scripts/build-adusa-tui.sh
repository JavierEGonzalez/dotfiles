#!/bin/bash
# Build and link adusa-tui binary to ~/bin
set -e

exec "$(dirname "${BASH_SOURCE[0]}")/install-adusa-tui.sh"
