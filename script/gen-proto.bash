#!/bin/bash

set -e

# Function to check if protoc is installed
check_protoc() {
  if ! command -v protoc &>/dev/null; then
    echo "Error: protoc is not installed or not in PATH"
    echo "Please install Protocol Buffers compiler:"
    echo "  macOS: brew install protobuf"
    echo "  Linux: apt-get install protobuf-compiler"
    echo "  Or download from: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
  fi

  echo "✓ Found protoc version: $(protoc --version)"
}

# Function to check if Go protoc plugins are installed
check_protoc_plugins() {
  local missing_plugins=()

  if ! command -v protoc-gen-go &>/dev/null; then
    missing_plugins+=("protoc-gen-go")
  fi

  if ! command -v protoc-gen-go-grpc &>/dev/null; then
    missing_plugins+=("protoc-gen-go-grpc")
  fi

  if [ ${#missing_plugins[@]} -ne 0 ]; then
    echo "Error: Missing required protoc plugins:"
    for plugin in "${missing_plugins[@]}"; do
      echo "  - $plugin"
    done
    echo ""
    echo "Install missing plugins with:"
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    echo ""
    echo "Make sure \$GOPATH/bin is in your PATH"
    exit 1
  fi

  echo "✓ Found protoc-gen-go"
  echo "✓ Found protoc-gen-go-grpc"
}

# Function to remove old generated files
remove_old_generated_files() {
  local proto_dir="$1"

  echo "Removing old generated files..."

  if [ ! -d "$proto_dir" ]; then
    echo "Warning: Proto directory not found: $proto_dir"
    return
  fi

  # Remove all .pb.go and _grpc.pb.go files
  find "$proto_dir" -name "*.pb.go" -delete 2>/dev/null || true
  find "$proto_dir" -name "*_grpc.pb.go" -delete 2>/dev/null || true

  echo "✓ Cleaned old generated files"
}

# Function to find and compile proto files
compile_proto_files() {
  local proto_base_dir="$1"
  local proto_files=()
  local file_count=0

  echo "Searching for .proto files in $proto_base_dir..."

  # Find all .proto files
  while IFS= read -r -d '' file; do
    proto_files+=("$file")
    ((file_count++))
  done < <(find "$proto_base_dir" -name "*.proto" -print0 2>/dev/null)

  if [ $file_count -eq 0 ]; then
    echo "⚠️  No .proto files found in $proto_base_dir"
    return
  fi

  echo "✓ Found $file_count .proto file(s)"
  echo ""

  # Compile each proto file
  for proto_file in "${proto_files[@]}"; do
    local proto_dir=$(dirname "$proto_file")
    local proto_name=$(basename "$proto_file")
    local relative_path="${proto_file#$PROJECT_ROOT/}"

    echo "Compiling: $relative_path"

    # Change to the directory containing the proto file
    cd "$proto_dir"

    # Compile with protoc
    # --go_out=. : Generate Go code in current directory
    # --go_opt=paths=source_relative : Use source-relative paths
    # --go-grpc_out=. : Generate gRPC Go code in current directory
    # --go-grpc_opt=paths=source_relative : Use source-relative paths
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           "$proto_name"

    echo "✓ Generated: ${proto_name%.proto}.pb.go"
    echo "✓ Generated: ${proto_name%.proto}_grpc.pb.go"
    echo ""

    cd - >/dev/null
  done
}

# Function to collect and display generated files
collect_generated_files() {
  local proto_base_dir="$1"
  local generated_files=()
  local file_count=0

  # Collect all generated .pb.go files
  if [ -d "$proto_base_dir" ]; then
    while IFS= read -r -d '' file; do
      generated_files+=("$file")
      ((file_count++))
    done < <(find "$proto_base_dir" -name "*.pb.go" -print0 2>/dev/null)
  fi

  # Display results
  echo ""
  echo "📊 Generated Files Summary:"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📁 Total files generated: $file_count"
  echo ""

  if [ $file_count -gt 0 ]; then
    echo "📋 File list:"
    for file in "${generated_files[@]}"; do
      # Get relative path from project root
      local relative_path="${file#$PROJECT_ROOT/}"
      local file_size=$(ls -lh "$file" | awk '{print $5}')
      echo "  📄 $relative_path ($file_size)"
    done
  else
    echo "⚠️  No generated files found"
  fi
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Main execution
main() {
  echo "======================================"
  echo "🔨 Protocol Buffers Code Generation Script"
  echo "======================================"
  echo ""
  echo "This script generates Go code from Protocol Buffer (.proto) files"
  echo "including gRPC service definitions."
  echo ""
  echo "What this script does:"
  echo "  1. ✅ Validates protoc and required plugins installation"
  echo "  2. 🧹 Removes old generated files (.pb.go, _grpc.pb.go)"
  echo "  3. 🔍 Searches for all .proto files in servers/internal/shared/protobuf/"
  echo "  4. 🔨 Compiles each .proto file to Go code"
  echo "  5. 📊 Displays summary of generated files"
  echo ""

  # Check if protoc is installed
  check_protoc
  echo ""

  # Check if protoc plugins are installed
  check_protoc_plugins
  echo ""

  # Get the script directory and project root
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

  # Define proto base directory
  PROTO_BASE_DIR="$PROJECT_ROOT/servers/internal/shared/pb"

  if [ ! -d "$PROTO_BASE_DIR" ]; then
    echo "Error: Proto base directory not found: $PROTO_BASE_DIR"
    exit 1
  fi

  # Remove old generated files
  remove_old_generated_files "$PROTO_BASE_DIR"
  echo ""

  # Compile proto files
  compile_proto_files "$PROTO_BASE_DIR"

  echo "🎉 All Protocol Buffer code generation completed successfully!"

  # Collect and display generated files
  collect_generated_files "$PROTO_BASE_DIR"

  echo ""
  echo "Generated files location:"
  echo "  📁 Go code: servers/internal/shared/pb/*/*.pb.go"
  echo ""
  echo "Next steps:"
  echo "  - Review generated code for any compilation errors"
  echo "  - Import and use generated types in your gRPC services"
  echo "  - Update server implementations to use new protobuf definitions"
  echo "  - Run 'go mod tidy' if new dependencies were added"
}

# Run main function
main "$@"
