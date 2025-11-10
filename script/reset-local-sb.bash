#!/bin/bash

set -e

# Function to check if required tools are installed
check_dependencies() {
  local missing_tools=()

  if ! command -v npx supabase &>/dev/null; then
    missing_tools+=("supabase")
  fi

  if [ ${#missing_tools[@]} -gt 0 ]; then
    echo "Error: Missing required tools: ${missing_tools[*]}"
    echo ""
    echo "Installation instructions:"
    for tool in "${missing_tools[@]}"; do
      case $tool in
      "supabase")
        echo "  - Supabase CLI: https://supabase.com/docs/guides/cli"
        ;;
      esac
    done
    exit 1
  fi
}

# Function to reset Supabase database
reset_supabase_database() {
  echo "🗄️ Resetting local Supabase database..."
  echo "  This will:"
  echo "    - Drop all existing data"
  echo "    - Apply all migrations from scratch"
  echo "    - Restore seed data"
  echo ""

  if npx supabase db reset --local; then
    echo "✓ Supabase database reset completed successfully"
  else
    echo "❌ Failed to reset Supabase database"
    exit 1
  fi
}



# Main execution
main() {
  echo "=================================================="
  echo "🔄 Supabase Database Reset Script"
  echo "=================================================="
  echo ""
  echo "This script performs a complete reset of the local Supabase database."
  echo ""
  echo "What this script does:"
  echo "  1. ✅ Validates required dependencies (supabase)"
  echo "  2. 🗄️ Resets local Supabase database (drops data, applies migrations)"
  echo ""
  echo "⚠️  WARNING: This will permanently delete all data in your local database!"
  echo ""

  # Check dependencies
  check_dependencies

  # Reset Supabase database
  reset_supabase_database

  echo ""
  echo "✅ Script completed successfully!"
  echo ""
  echo "Results:"
  echo "  🗄️ Supabase database: Completely reset with fresh schema and seed data"
  echo ""
  echo "Next steps:"
  echo "  - Your local database is now in a clean state"
  echo "  - Start your services to begin accepting connections"
  echo "  - Check logs for any migration errors"
}

# Run main function
main "$@"
