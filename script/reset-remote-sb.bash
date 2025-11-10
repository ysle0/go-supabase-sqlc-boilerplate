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

# Function to check if user is logged into Supabase
check_supabase_auth() {
  echo "🔐 Checking Supabase authentication..."

  if ! npx supabase projects list &>/dev/null; then
    echo "❌ You are not logged into Supabase CLI"
    echo ""
    echo "Please log in first:"
    echo "  supabase login"
    echo ""
    echo "Then link your project:"
    echo "  supabase link --project-ref YOUR_PROJECT_REF"
    exit 1
  fi

  echo "✓ Supabase authentication verified"
}

# Function to get project reference
get_project_ref() {
  echo "📋 Getting project reference..."

  # Try to get project ref from linked project
  local project_ref=$(npx supabase status --output=json 2>/dev/null | grep -o '"project_id":"[^"]*"' | cut -d'"' -f4 || true)

  if [ -z "$project_ref" ]; then
    echo "❌ Could not determine project reference. Make sure you have linked a project:"
    echo "  supabase link --project-ref YOUR_PROJECT_REF"
    exit 1
  fi

  echo "✓ Project reference: $project_ref"
  echo "$project_ref"
}

# Function to reset remote Supabase database
reset_supabase_remote() {
  local project_ref="$1"

  echo "🗄️ Resetting remote Supabase database..."
  echo "  Project: $project_ref"
  echo "  This will:"
  echo "    - Drop all existing data in the remote database"
  echo "    - Apply all migrations from scratch"
  echo "    - Restore seed data"
  echo ""
  echo "⚠️  WARNING: This will permanently delete all data in your REMOTE production database!"
  echo ""

  read -p "Are you absolutely sure you want to continue? (type 'YES' to confirm): " confirm
  if [ "$confirm" != "YES" ]; then
    echo "Operation cancelled."
    exit 1
  fi

  echo ""
  echo "🔄 Proceeding with remote database reset..."

  if npx supabase db reset --linked; then
    echo "✓ Remote Supabase database reset completed successfully"
  else
    echo "❌ Failed to reset remote Supabase database"
    exit 1
  fi
}


# Main execution
main() {
  echo "=========================================================="
  echo "🔄 Remote Supabase Database Reset Script"
  echo "=========================================================="
  echo ""
  echo "This script performs a complete reset of the REMOTE Supabase database."
  echo ""
  echo "What this script does:"
  echo "  1. ✅ Validates required dependencies (supabase)"
  echo "  2. 🔐 Verifies Supabase CLI authentication"
  echo "  3. 📋 Gets linked project information"
  echo "  4. 🗄️ Resets remote Supabase database (drops data, applies migrations)"
  echo ""
  echo "⚠️  WARNING: This will permanently delete all data in your REMOTE production database!"
  echo ""

  # Check dependencies
  check_dependencies

  # Check Supabase authentication
  check_supabase_auth

  # Get project reference
  local project_ref=$(get_project_ref)

  echo ""

  # Reset remote Supabase database
  reset_supabase_remote "$project_ref"

  echo ""
  echo "✅ Script completed successfully!"
  echo ""
  echo "Results:"
  echo "  🗄️ Remote Supabase database: Completely reset with fresh schema and seed data"
  echo ""
  echo "Next steps:"
  echo "  - Your remote database is now in a clean state"
  echo "  - Start or restart your services to begin accepting connections"
  echo "  - Check your deployment logs for any migration errors"
  echo "  - Verify your application is working with the reset database"
}

# Run main function
main "$@"
