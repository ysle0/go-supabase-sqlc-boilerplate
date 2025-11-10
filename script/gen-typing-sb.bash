#!/bin/bash

set -e

# Function to check if Supabase CLI is installed
check_supabase() {
    if ! command -v npx supabase &> /dev/null; then
        echo "Error: Supabase CLI is not installed or not in PATH"
        echo "Please install Supabase CLI: https://supabase.com/docs/guides/cli"
        exit 1
    fi
}

# Function to check if we're in the correct directory
check_directory() {
    if [ ! -d "../supabase" ]; then
        echo "Error: supabase directory not found. Make sure you're running this script from the script/ directory"
        exit 1
    fi

    if [ ! -d "../supabase/functions/_shared" ]; then
        echo "Creating _shared directory in supabase/functions/"
        mkdir -p "../supabase/functions/_shared"
    fi
}

# Function to generate TypeScript types
generate_types() {
    echo "Generating Supabase TypeScript types..."

    # Change to supabase directory to run the command
    cd ../supabase

    # Generate types and save to _shared/schema.ts
    npx supabase gen types typescript --local > functions/_shared/schema.ts

    # Check if the file was created successfully
    if [ -f "functions/_shared/schema.ts" ]; then
        echo "✓ TypeScript types generated successfully: supabase/functions/_shared/schema.ts"

        # Show file size for verification
        local file_size=$(wc -c < "functions/_shared/schema.ts")
        echo "  File size: $file_size bytes"

        # Show first few lines to verify content
        echo "  Preview (first 10 lines):"
        head -n 10 "functions/_shared/schema.ts" | sed 's/^/    /'
    else
        echo "Error: Failed to generate schema.ts file"
        exit 1
    fi

    cd - > /dev/null
}

# Main execution
main() {
    echo "====================================="
    echo "📝 Supabase TypeScript Types Generator"
    echo "====================================="

    # Check if Supabase CLI is installed
    check_supabase

    # Check directory structure
    check_directory

    # Generate TypeScript types
    generate_types

    echo ""
    echo "🎉 Supabase TypeScript types generation completed successfully!"
    echo ""
    echo "Next steps:"
    echo "  - Import types in your Edge Functions: import { Database } from './_shared/schema.ts'"
    echo "  - Use types for type-safe database operations"
}

# Run main function
main "$@"
