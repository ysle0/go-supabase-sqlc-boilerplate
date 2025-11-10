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

# Function to kill any existing processes on port 8080
kill_existing_server() {
  echo "🔍 Checking for existing processes on port 8080..."

  # Find processes using port 8080
  local pids=$(lsof -ti:8080 2>/dev/null || true)

  if [ -n "$pids" ]; then
    echo "⚠️  Found existing processes on port 8080: $pids"
    echo "🛑 Killing existing processes..."

    # Kill the processes
    echo "$pids" | xargs kill -9 2>/dev/null || true

    # Wait a moment for processes to terminate
    sleep 2

    # Verify processes are gone
    local remaining=$(lsof -ti:8080 2>/dev/null || true)
    if [ -n "$remaining" ]; then
      echo "❌ Warning: Some processes may still be running on port 8080: $remaining"
    else
      echo "✓ Successfully killed all processes on port 8080"
    fi
  else
    echo "✓ No existing processes found on port 8080"
  fi

  echo ""
}

# Function to wait for server to be ready
wait_for_server() {
  local max_attempts=10
  local attempt=1

  echo "Waiting for question server to be ready..."

  while [ $attempt -le $max_attempts ]; do
    if curl -s -f "http://localhost:8080/health" >/dev/null 2>&1; then
      echo "✓ Question server is ready"
      return 0
    fi

    echo "  Attempt $attempt/$max_attempts - Server not ready yet..."
    sleep 2
    attempt=$((attempt + 1))
  done

  echo "Error: Question server failed to start after $max_attempts attempts"
  return 1
}

# Function to execute parse requests
execute_parse_requests() {
  echo ""
  echo "📊 Executing question parse requests..."
  echo ""

  # Execute classic mode parse request
  echo "🎮 Parsing questions for classic mode..."
  if curl -s -X GET "http://localhost:8080/v1/parse/classic-mode" \
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
  if curl -s -X GET "http://localhost:8080/v1/parse/challenge-mode?game_type=all" \
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
  echo "=================================================="
  echo "🔄 Supabase Database + Questions Reset Script"
  echo "=================================================="
  echo ""
  echo "This script performs a complete reset of the local Supabase database"
  echo "and refreshes question data from Google Sheets."
  echo ""
  echo "What this script does:"
  echo "  1. ✅ Validates required dependencies (supabase, curl)"
  echo "  2. 🗄️ Resets local Supabase database (drops data, applies migrations)"
  echo "  3. 🚀 Starts the question server in the background"
  echo "  4. ⏳ Waits for the server to be ready"
  echo "  5. 📊 Parses questions for classic mode"
  echo "  6. 📊 Parses questions for challenge-mode mode"
  echo "  7. 🛑 Stops the question server"
  echo ""
  echo "⚠️  WARNING: This will permanently delete all data in your local database!"
  echo ""

  # Check dependencies
  check_dependencies

  # Reset Supabase database
  reset_supabase_database

  echo ""

  # Kill any existing server on port 8080
  kill_existing_server

  # Build and start question server in background
  echo "🚀 Building question server..."
  cd ../servers/cmd/question/
  if ! go build -o question .; then
    echo "❌ Failed to build question server"
    exit 1
  fi
  echo "✓ Question server built successfully"

  echo "🚀 Starting question server..."
  ./question &
  SERVER_PID=$!

  sleep 5

  # Wait for server to start and be ready
  # if ! wait_for_server; then
  #     echo "🛑 Stopping question server due to startup failure..."
  #     kill -9 $SERVER_PID 2>/dev/null || true
  #     exit 1
  # fi

  # Execute parse requests
  if execute_parse_requests; then
    echo ""
    echo "🎉 Database reset and question parsing completed successfully!"
  else
    echo ""
    echo "❌ Question parsing failed"
    echo "🛑 Stopping question server..."
    kill -9 $SERVER_PID 2>/dev/null || true
    exit 1
  fi

  # Kill the server
  echo ""
  echo "🛑 Stopping question server..."
  kill -9 $SERVER_PID 2>/dev/null || true

  echo ""
  echo "✅ Script completed successfully!"
  echo ""
  echo "Results:"
  echo "  🗄️ Supabase database: Completely reset with fresh schema and seed data"
  echo "  📊 Classic mode questions: Updated from Google Sheets"
  echo "  📊 challenge-mode mode questions: Updated from Google Sheets"
  echo ""
  echo "Next steps:"
  echo "  - Your local database is now in a clean state"
  echo "  - Questions are ready for game sessions"
  echo "  - Start the ingame server to begin accepting game connections"
  echo "  - Check logs for any migration or parsing errors"
}

# Run main function
main "$@"
