#!/bin/bash

# Build script for Veranda
# Usage: ./build.sh [dev|release|icons|clean]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${GREEN}Veranda Build Script${NC}"
echo "===================="

case "${1:-dev}" in
  dev)
    echo -e "${YELLOW}Running development mode...${NC}"
    echo "This requires two terminals. Running in single-terminal mode..."
    echo ""
    cd "$PROJECT_ROOT/src-tauri"
    cargo tauri dev
    ;;
    
  release)
    echo -e "${YELLOW}Building for release...${NC}"
    
    # Build frontend
    echo "Building frontend..."
    cd "$PROJECT_ROOT"
    npm run build
    
    # Build Tauri app
    echo "Building Tauri app..."
    cd "$PROJECT_ROOT/src-tauri"
    cargo tauri build
    
    echo -e "${GREEN}Build complete!${NC}"
    echo "App bundle location:"
    ls -la "$PROJECT_ROOT/src-tauri/target/release/bundle/"
    ;;
    
  icons)
    echo -e "${YELLOW}Generating icons...${NC}"
    cd "$PROJECT_ROOT"
    node scripts/generate-icons.cjs
    echo -e "${GREEN}Icons generated!${NC}"
    ;;
    
  check)
    echo -e "${YELLOW}Checking code...${NC}"
    cd "$PROJECT_ROOT/src-tauri"
    cargo check
    echo -e "${GREEN}Rust code checks passed!${NC}"
    ;;
    
  clean)
    echo -e "${YELLOW}Cleaning build artifacts...${NC}"
    cd "$PROJECT_ROOT"
    rm -rf dist
    rm -rf node_modules/.cache
    cd "$PROJECT_ROOT/src-tauri"
    cargo clean
    echo -e "${GREEN}Clean complete!${NC}"
    ;;
    
  *)
    echo "Usage: $0 [dev|release|icons|check|clean]"
    echo ""
    echo "Commands:"
    echo "  dev     - Run in development mode"
    echo "  release - Build for production"
    echo "  icons   - Generate app icons"
    echo "  check   - Check Rust code"
    echo "  clean   - Clean build artifacts"
    exit 1
    ;;
esac