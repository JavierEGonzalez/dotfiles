#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")/adusa-tui"
BIN_DIR="$HOME/bin"
BINARY_NAME="adusa-tui"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Building ${BINARY_NAME}...${NC}"

if [ ! -d "$PROJECT_ROOT" ]; then
  echo -e "${RED}Error: adusa-tui project not found at $PROJECT_ROOT${NC}"
  exit 1
fi

cd "$PROJECT_ROOT"
go build -o "$BINARY_NAME"

if [ ! -f "$BINARY_NAME" ]; then
  echo -e "${RED}Error: Build failed - binary not created${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Build successful${NC}"

if [ ! -d "$BIN_DIR" ]; then
  echo -e "${YELLOW}Creating $BIN_DIR directory${NC}"
  mkdir -p "$BIN_DIR"
fi

echo -e "${YELLOW}Linking binary to $BIN_DIR/$BINARY_NAME${NC}"

if [ -L "$BIN_DIR/$BINARY_NAME" ]; then
  rm "$BIN_DIR/$BINARY_NAME"
elif [ -f "$BIN_DIR/$BINARY_NAME" ]; then
  echo -e "${RED}Warning: $BIN_DIR/$BINARY_NAME exists and is not a symlink. Backing up...${NC}"
  mv "$BIN_DIR/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME.bak"
fi

ln -s "$PROJECT_ROOT/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME"

echo -e "${GREEN}✓ Symlink created${NC}"

if command -v "$BINARY_NAME" &> /dev/null; then
  echo -e "${GREEN}✓ Binary is accessible in PATH${NC}"
  echo -e "${GREEN}✓ Installation complete!${NC}"
else
  echo -e "${YELLOW}Note: $BINARY_NAME is not in PATH. Make sure $BIN_DIR is in your \$PATH${NC}"
  echo -e "${YELLOW}Add this to your shell config:${NC}"
  echo "  export PATH=\"$BIN_DIR:\$PATH\""
fi
