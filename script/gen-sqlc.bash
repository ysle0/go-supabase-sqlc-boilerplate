#!/bin/bash

set -e

# Function to check if sqlc is installed
check_sqlc() {
  if ! command -v sqlc &>/dev/null; then
    echo "Error: sqlc is not installed or not in PATH"
    echo "Please install sqlc: https://docs.sqlc.dev/en/latest/overview/install.html"
    exit 1
  fi
}

# Function to remove old generated files
remove_old_generated_files() {
  local config_file="$1"
  local config_name="$2"
  local config_dir=$(dirname "$config_file")

  echo "Removing old generated files for $config_name..."

  if [ ! -f "$config_file" ]; then
    echo "Error: Config file not found: $config_file"
    exit 1
  fi

  cd "$config_dir"

  # Parse config file to find output directories and remove generated files
  if grep -q "go:" "$(basename "$config_file")"; then
    # Go generation - remove .go files in output directory
    local go_out_dir=$(grep -A5 "go:" "$(basename "$config_file")" | grep "out:" | sed 's/.*out: *"\?\([^"]*\)"\?.*/\1/' | head -1)
    if [ -n "$go_out_dir" ] && [ -d "$go_out_dir" ]; then
      echo "  Removing Go files from $go_out_dir"
      find "$go_out_dir" -name "*.query.sql.go" -delete 2>/dev/null || true
      find "$go_out_dir" -name "models.go" -delete 2>/dev/null || true
      find "$go_out_dir" -name "db.go" -delete 2>/dev/null || true
      find "$go_out_dir" -name "copyfrom.go" -delete 2>/dev/null || true
    fi
  fi

  if grep -q "typescript:" "$(basename "$config_file")"; then
    # TypeScript generation - remove .ts files in output directory
    local ts_out_dir=$(grep -A5 "typescript:" "$(basename "$config_file")" | grep "out:" | sed 's/.*out: *"\?\([^"]*\)"\?.*/\1/' | head -1)
    if [ -n "$ts_out_dir" ] && [ -d "$ts_out_dir" ]; then
      echo "  Removing TypeScript files from $ts_out_dir"
      find "$ts_out_dir" -name "*_query.sql.ts" -delete 2>/dev/null || true
    fi
  fi

  cd - >/dev/null
}

# Function to run sqlc generate with specific config file
run_sqlc_generate() {
  local config_file="$1"
  local config_name="$2"
  local config_dir=$(dirname "$config_file")

  echo "Generating SQL code from $config_name..."

  if [ ! -f "$config_file" ]; then
    echo "Error: Config file not found: $config_file"
    exit 1
  fi

  # Remove old generated files first
  remove_old_generated_files "$config_file" "$config_name"

  cd "$config_dir"
  sqlc generate -f "$(basename "$config_file")"
  echo " Generated SQL code from $config_name"
  cd - >/dev/null
}

# Function to collect and display generated files
collect_generated_files() {
  local generated_files=()
  local file_count=0

  # Collect PostgreSQL Go files
  if [ -d "$PROJECT_ROOT/servers/internal/sql" ]; then
    while IFS= read -r -d '' file; do
      generated_files+=("$file")
      ((file_count++))
    done < <(find "$PROJECT_ROOT/servers/internal/sql" -name "*.go" -print0 2>/dev/null)
  fi

  # Collect TypeScript files
  if [ -d "$PROJECT_ROOT/supabase/functions/_shared/queries" ]; then
    while IFS= read -r -d '' file; do
      generated_files+=("$file")
      ((file_count++))
    done < <(find "$PROJECT_ROOT/supabase/functions/_shared/queries" -name "*.ts" -print0 2>/dev/null)
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
      echo "  📄 $relative_path"
    done
  else
    echo "⚠️  No generated files found"
  fi
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Main execution
main() {
  echo "======================================"
  echo "🔨 SQLC Code Generation Script"
  echo "======================================"
  echo ""
  echo "This script generates type-safe Go and TypeScript code from SQL queries"
  echo "using SQLC. It processes PostgreSQL and SQLite Go code generation, plus"
  echo "Supabase Edge Functions TypeScript code."
  echo ""
  echo "What this script does:"
  echo "  1. ✅ Validates SQLC installation"
  echo "  2. 🧹 Removes old generated files to prevent conflicts"
  echo "  3. 🔨 Generates PostgreSQL Go code for servers (servers/scripts/)"
  echo "  4. 🔨 Generates TypeScript code for Supabase functions (supabase/scripts/)"
  echo "  5. 🔧 Fixes TypeScript imports for Deno compatibility"
  echo ""

  # Check if sqlc is installed
  check_sqlc

  # Get the script directory and project root
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

  # Define paths and config files
  SERVER_SCRIPTS_DIR="$PROJECT_ROOT/servers/scripts"
  SUPABASE_SQLC_DIR="$PROJECT_ROOT/supabase/scripts"
  POSTGRESQL_CONFIG="$SERVER_SCRIPTS_DIR/sqlc-postgresql.yaml"

  # Generate PostgreSQL Go code for servers
  if [ -f "$POSTGRESQL_CONFIG" ]; then
    run_sqlc_generate "$POSTGRESQL_CONFIG" "servers (PostgreSQL)"
  else
    echo "Warning: PostgreSQL config not found: $POSTGRESQL_CONFIG"
  fi

  # Generate TypeScript code for Supabase functions
  if [ -d "$SUPABASE_SQLC_DIR" ]; then
    SUPABASE_CONFIG="$SUPABASE_SQLC_DIR/sqlc.yaml"
    if [ -f "$SUPABASE_CONFIG" ]; then
      run_sqlc_generate "$SUPABASE_CONFIG" "supabase (TypeScript)"
  
      # Fix TypeScript imports for Deno and insert type alias
      echo "Fixing TypeScript imports for Deno and inserting 'type Sql = postgres.Sql;' at line 3..."
  
      # Replace sqlc's import of Sql with Deno-compatible import
      find "$PROJECT_ROOT/supabase/functions/_shared/queries" -name "*.ts" -exec sed -i '' 's/import { Sql } from "postgres";/import postgres from "https:\/\/deno.land\/x\/postgresjs\@v3.4.7\/mod.js";/g' {} \;
      find "$PROJECT_ROOT/supabase/functions/_shared/queries" -name "*.ts" -exec sed -i '' 's/import { Sql } from "npm:postgres";/import postgres from "https:\/\/deno.land\/x\/postgresjs\@v3.4.7\/mod.js";/g' {} \;
  
      # Insert the type alias at line 3 if not already present
      find "$PROJECT_ROOT/supabase/functions/_shared/queries" -name "*.ts" -exec sh -c '
f="$1"
if ! grep -q "type Sql = postgres.Sql;" "$f"; then
  lines=$(wc -l < "$f")
  if [ "$lines" -lt 2 ]; then
    # ensure at least two lines so insertion at line 3 places alias as 3rd line
    echo "" >> "$f"
  fi
  awk '\''NR==3{print "type Sql = postgres.Sql;"} {print}'\'' "$f" > "$f.tmp" && mv "$f.tmp" "$f"
fi' _ {} \;
  
      echo "✓ Fixed TypeScript imports and inserted type alias"
    else
      echo "Warning: Supabase config not found: $SUPABASE_CONFIG"
    fi
  else
    echo "Warning: Supabase SQLC directory not found: $SUPABASE_SQLC_DIR"
  fi

  echo ""
  echo "🎉 All SQLC code generation completed successfully!"
  
  # Collect and display generated files
  collect_generated_files
  
  echo ""
  echo "Generated files locations:"
  echo "  📁 PostgreSQL Go code: servers/internal/sql/"
  echo "  📁 TypeScript code: supabase/functions/_shared/queries/"
  echo ""
  echo "Next steps:"
  echo "  - Review generated code for any compilation errors"
  echo "  - Import and use generated types in your application code"
  echo "  - Run tests to ensure database operations work correctly"
}

# Run main function
main "$@"

