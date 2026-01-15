#!/bin/bash
# Tortoise Benchmark Script

set -e

echo "======================================"
echo "Tortoise Framework Benchmark Suite"
echo "======================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Functions
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check prerequisites
check_prereqs() {
    log_info "Checking prerequisites..."
    
    local missing=()
    
    which curl > /dev/null || missing+=("curl")
    which jq > /dev/null || missing+=("jq")
    which ab > /dev/null 2>/dev/null || log_warn "apache-bench not found (optional)"
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing tools: ${missing[*]}"
        exit 1
    fi
    
    log_info "Prerequisites OK"
}

# Start gateway
start_gateway() {
    log_info "Starting Tortoise Gateway..."
    
    # Kill any existing gateway
    pkill -f tortoise-gateway 2>/dev/null || true
    
    # Start gateway in background
    cd tortoise-core
    cargo build --release 2>/dev/null
    ./target/release/tortoise-gateway &
    cd ..
    
    GATEWAY_PID=$!
    
    # Wait for gateway to be ready
    log_info "Waiting for gateway to be ready..."
    for i in {1..30}; do
        if curl -s http://127.0.0.1:8080/health > /dev/null; then
            log_info "Gateway ready!"
            return 0
        fi
        sleep 1
    done
    
    log_error "Gateway failed to start"
    return 1
}

# Stop gateway
stop_gateway() {
    log_info "Stopping gateway..."
    pkill -f tortoise-gateway 2>/dev/null || true
}

# Benchmark 1: Cold Start
benchmark_cold_start() {
    log_info "Benchmark 1: Cold Start"
    
    start_time=$(date +%s%N)
    
    # Create a new agent
    RESPONSE=$(curl -s -X POST http://127.0.0.1:8080/api/v1/agents \
        -H "Content-Type: application/json" \
        -d '{"name":"bench-agent","model_provider":"openai","model":"gpt-4","skills":[]}')
    
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    echo "  Response: $RESPONSE"
    echo "  Duration: ${duration}ms"
}

# Benchmark 2: Agent Creation (throughput)
benchmark_agent_creation() {
    log_info "Benchmark 2: Agent Creation Throughput"
    
    local count=100
    start_time=$(date +%s%N)
    
    for i in $(seq 1 $count); do
        curl -s -X POST http://127.0.0.1:8080/api/v1/agents \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"agent-$i\",\"model_provider\":\"openai\",\"model\":\"gpt-4\",\"skills\":[]}" > /dev/null
    done
    
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    throughput=$(echo "scale=2; $count * 1000 / $duration" | bc)
    
    echo "  Created: $count agents"
    echo "  Total time: ${duration}ms"
    echo "  Throughput: ${throughput} agents/sec"
}

# Benchmark 3: Memory Operations
benchmark_memory() {
    log_info "Benchmark 3: Memory Operations"
    
    # Store
    start_time=$(date +%s%N)
    for i in $(seq 1 1000); do
        curl -s -X POST http://127.0.0.1:8080/api/v1/memory \
            -H "Content-Type: application/json" \
            -d "{\"key\":\"bench-$i\",\"value\":\"test-value-$i\"}" > /dev/null
    done
    end_time=$(date +%s%N)
    store_duration=$(( (end_time - start_time) / 1000000 ))
    store_throughput=$(echo "scale=2; 1000 * 1000 / $store_duration" | bc)
    
    # Recall
    start_time=$(date +%s%N)
    for i in $(seq 1 1000); do
        curl -s "http://127.0.0.1:8080/api/v1/memory/bench-$i" > /dev/null
    done
    end_time=$(date +%s%N)
    recall_duration=$(( (end_time - start_time) / 1000000 ))
    recall_throughput=$(echo "scale=2; 1000 * 1000 / $recall_duration" | bc)
    
    echo "  Store: 1000 ops in ${store_duration}ms (${store_throughput} ops/sec)"
    echo "  Recall: 1000 ops in ${recall_duration}ms (${recall_throughput} ops/sec)"
}

# Benchmark 4: Search
benchmark_search() {
    log_info "Benchmark 4: Memory Search"
    
    # Store many entries first
    for i in $(seq 1 500); do
        curl -s -X POST http://127.0.0.1:8080/api/v1/memory \
            -H "Content-Type: application/json" \
            -d "{\"key\":\"search-$i\",\"value\":\"test value $i\"}" > /dev/null
    done
    
    # Search
    start_time=$(date +%s%N)
    for i in $(seq 1 100); do
        curl -s -X POST http://127.0.0.1:8080/api/v1/memory/search \
            -H "Content-Type: application/json" \
            -d '{"query":"test","limit":10}' > /dev/null
    done
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    throughput=$(echo "scale=2; 100 * 1000 / $duration" | bc)
    
    echo "  Search: 100 queries in ${duration}ms (${throughput} queries/sec)"
}

# Benchmark 5: MCP Tools
benchmark_mcp_tools() {
    log_info "Benchmark 5: MCP Tools"
    
    # List tools
    start_time=$(date +%s%N)
    for i in $(seq 1 500); do
        curl -s http://127.0.0.1:8080/api/v1/mcp/tools > /dev/null
    done
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    throughput=$(echo "scale=2; 500 * 1000 / $duration" | bc)
    
    echo "  List Tools: 500 calls in ${duration}ms (${throughput} calls/sec)"
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    stop_gateway
}

# Main
main() {
    check_prereqs
    
    trap cleanup EXIT
    
    start_gateway
    
    echo ""
    echo "======================================"
    echo "Running Benchmarks"
    echo "======================================"
    
    benchmark_cold_start
    echo ""
    benchmark_agent_creation
    echo ""
    benchmark_memory
    echo ""
    benchmark_search
    echo ""
    benchmark_mcp_tools
    
    echo ""
    echo "======================================"
    echo "Benchmark Complete"
    echo "======================================"
}

main "$@"
