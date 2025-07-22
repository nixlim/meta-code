#!/bin/bash
# Install git hooks for the meta-mcp-server project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing git hooks for meta-mcp-server...${NC}"

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if we're in a git repository
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo -e "${RED}Error: Not in a git repository!${NC}"
    echo "Please run this script from the project root."
    exit 1
fi

# Create hooks directory if it doesn't exist
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
mkdir -p "$HOOKS_DIR"

# Install pre-commit hook
echo -e "${YELLOW}Installing pre-commit hook...${NC}"
if [ -f "$HOOKS_DIR/pre-commit" ] && [ ! -L "$HOOKS_DIR/pre-commit" ]; then
    echo -e "${YELLOW}Backing up existing pre-commit hook to pre-commit.bak${NC}"
    mv "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/pre-commit.bak"
fi

# Create the pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
# Pre-commit hook for running golangci-lint
# This hook runs golangci-lint on staged Go files before committing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running golangci-lint pre-commit hook...${NC}"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}golangci-lint is not installed!${NC}"
    echo "Please install it by running: make lint-install"
    exit 1
fi

# Get list of staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    echo -e "${GREEN}No Go files to lint${NC}"
    exit 0
fi

# Create a temporary directory for staged files
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Copy staged files to temporary directory maintaining directory structure
for FILE in $STAGED_GO_FILES; do
    mkdir -p "$TMPDIR/$(dirname $FILE)"
    git show ":$FILE" > "$TMPDIR/$FILE"
done

# Run golangci-lint on the temporary directory
echo -e "${YELLOW}Linting staged Go files...${NC}"

# Change to the temporary directory
cd "$TMPDIR"

# Copy go.mod and go.sum if they exist (needed for some linters)
if [ -f "${OLDPWD}/go.mod" ]; then
    cp "${OLDPWD}/go.mod" .
fi
if [ -f "${OLDPWD}/go.sum" ]; then
    cp "${OLDPWD}/go.sum" .
fi

# Copy .golangci.yml if it exists
if [ -f "${OLDPWD}/.golangci.yml" ]; then
    cp "${OLDPWD}/.golangci.yml" .
fi

# Run golangci-lint
if golangci-lint run --timeout=5m $STAGED_GO_FILES; then
    echo -e "${GREEN}✓ Linting passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Linting failed!${NC}"
    echo -e "${YELLOW}Please fix the issues above before committing.${NC}"
    echo -e "${YELLOW}You can bypass this hook with --no-verify (not recommended)${NC}"
    exit 1
fi
EOF

# Make the hook executable
chmod +x "$HOOKS_DIR/pre-commit"
echo -e "${GREEN}✓ Pre-commit hook installed${NC}"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}⚠ golangci-lint is not installed${NC}"
    echo -e "${BLUE}Installing golangci-lint...${NC}"
    make -C "$PROJECT_ROOT" lint-install
fi

# Verify installation
echo -e "${BLUE}Verifying hook installation...${NC}"
if [ -x "$HOOKS_DIR/pre-commit" ]; then
    echo -e "${GREEN}✓ Pre-commit hook is properly installed and executable${NC}"
else
    echo -e "${RED}✗ Pre-commit hook installation failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Git hooks installation completed successfully!${NC}"
echo -e "${BLUE}The pre-commit hook will run golangci-lint on staged Go files before each commit.${NC}"
echo -e "${YELLOW}To bypass the hook (not recommended), use: git commit --no-verify${NC}"