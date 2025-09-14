#!/bin/bash

# Test script to demonstrate Local Layers functionality
echo "🧪 Testing Local Layers functionality..."

# Create test layer directory if it doesn't exist
mkdir -p ./layers/test-layer
echo "test_config: local_value" > ./layers/test-layer/test.yaml
echo "#!/bin/bash\necho 'Hello from local layer!'" > ./layers/test-layer/hello.sh
chmod +x ./layers/test-layer/hello.sh

# Create test Otterfile
cat > ./test-Otterfile << 'EOF'
# Test local layers
VAR PROJECT_NAME=local-test

# Local layer with relative path
LAYER ./layers/test-layer TARGET test-output

# Local layer with template variables
LAYER ./layers/app-config TARGET app TEMPLATE project=${PROJECT_NAME} env=development
EOF

echo "📁 Created test layer at ./layers/test-layer"
echo "📄 Created test Otterfile"

# Initialize otter if needed
if [ ! -d ".otter" ]; then
    echo "🔧 Initializing otter..."
    otter init
fi

# Test the local layers
echo "🚀 Testing local layers with otter build..."
otter build -f ./test-Otterfile

echo ""
echo "✅ Test completed! Check the following:"
echo "   - test-output/ directory should contain files from ./layers/test-layer"
echo "   - app/ directory should contain config.json with substituted variables"

# Clean up
echo ""
echo "🧹 Cleaning up test files..."
rm -f ./test-Otterfile
rm -rf ./test-output
rm -rf ./app
rm -rf ./layers/test-layer

echo "✨ Local layers test complete!"
