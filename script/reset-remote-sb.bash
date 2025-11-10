#!/bin/bash

set -e

# Function to check if required tools are installed
check_dependencies() {
  local missing_tools=()

  if ! command -v npx supabase &>/dev/null; then
    missing_tools+=("supabase")
  fi

  if ! command -v curl &>/dev/null; then
    missing_tools+=("curl")
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
      "curl")
        echo "  - curl: Usually pre-installed or available via package manager"
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

# Function to get remote database URL and credentials
get_remote_connection_info() {
  echo "🔍 Getting remote database connection info..."

  # Get the connection info from Supabase
  local db_url=$(npx supabase status --output=json 2>/dev/null | grep -o '"db_url":"[^"]*"' | cut -d'"' -f4 || true)

  if [ -z "$db_url" ]; then
    echo "❌ Could not get remote database URL"
    echo "Make sure your project is properly linked and you have the necessary permissions"
    exit 1
  fi

  echo "✓ Remote database connection verified"
  echo "$db_url"
}

# Function to wait for remote server to be ready
wait_for_remote_server() {
  local max_attempts=15
  local attempt=1
  local server_url="$1"

  echo "⏳ Waiting for remote question server to be ready..."

  while [ $attempt -le $max_attempts ]; do
    if curl -s -f "${server_url}/health" >/dev/null 2>&1; then
      echo "✓ Remote question server is ready"
      return 0
    fi

    echo "  Attempt $attempt/$max_attempts - Server not ready yet..."
    sleep 3
    attempt=$((attempt + 1))
  done

  echo "❌ Remote question server failed to respond after $max_attempts attempts"
  echo "Please check if your question server is deployed and accessible"
  return 1
}

# Function to execute parse requests against remote server
execute_remote_parse_requests() {
  echo ""
  echo "📊 Executing question parse requests against remote server..."
  echo "  Server URL: https://trivia-question.fly.dev/"
  echo ""

  # Execute classic mode parse request
  echo "🎮 Parsing questions for classic mode..."
  if curl -s -X GET "https://trivia-question.fly.dev/v1/parse/classic-mode" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json"; then
    echo ""
    echo "✓ Classic mode questions parsed successfully"
  else
    echo "❌ Failed to parse classic mode questions"
    return 1
  fi

  echo ""

  # Execute challenge-mode mode parse request
  echo "🎯 Parsing questions for challenge-mode mode..."
  if curl -s -X GET "https://trivia-question.fly.dev/v1/parse/challenge-mode?game_type=all" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json"; then
    echo ""
    echo "✓ challenge-mode mode questions parsed successfully"
  else
    echo "❌ Failed to parse challenge-mode mode questions"
    return 1
  fi
}

# Main execution
main() {
  echo "=========================================================="
  echo "🔄 Remote Supabase Database + Questions Reset Script"
  echo "=========================================================="
  echo ""
  echo "This script performs a complete reset of the REMOTE Supabase database"
  echo "and refreshes question data from Google Sheets via deployed server."
  echo ""
  echo "What this script does:"
  echo "  1. ✅ Validates required dependencies (supabase, curl)"
  echo "  2. 🔐 Verifies Supabase CLI authentication"
  echo "  3. 📋 Gets linked project information"
  echo "  4. 🗄️ Resets remote Supabase database (drops data, applies migrations)"
  echo "  5. 🌐 Gets production question server URL from test.http"
  echo "  6. ⏳ Waits for the remote server to be ready"
  echo "  7. 📊 Parses questions for classic mode"
  echo "  8. 📊 Parses questions for challenge-mode mode"
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

  # Get production server URL for question parsing
  local server_url
  # Wait for remote server to be ready
  sleep 3

  # Execute parse requests
  if execute_remote_parse_requests; then
    echo ""
    echo "🎉 Remote database reset and question parsing completed successfully!"
  else
    echo ""
    echo "❌ Question parsing failed"
    echo "The database was reset successfully, but question parsing encountered errors."
    echo "You may need to manually trigger question parsing or check your server logs."
    exit 1
  fi

  echo ""
  echo "✅ Script completed successfully!"
  echo ""
  echo "Results:"
  echo "  🗄️ Remote Supabase database: Completely reset with fresh schema and seed data"
  if [ -n "$server_url" ]; then
    echo "  📊 Classic mode questions: Updated from Google Sheets"
    echo "  📊 challenge-mode mode questions: Updated from Google Sheets"
  else
    echo "  📊 Question parsing: Skipped (could not get production server URL)"
  fi
  echo ""
  echo "Next steps:"
  echo "  - Your remote database is now in a clean state"
  if [ -n "$server_url" ]; then
    echo "  - Questions are ready for game sessions"
  else
    echo "  - Manually trigger question parsing if needed"
  fi
  echo "  - Start or restart your ingame server to begin accepting game connections"
  echo "  - Check your deployment logs for any migration or parsing errors"
  echo "  - Verify your application is working with the reset database"
}

# Run main function
main "$@"
