#!/bin/bash
# Docker AMD64 Container Manager
# Manage the persistent development container

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

CONTAINER_NAME="ravact-amd64-dev"

show_usage() {
    echo "Docker AMD64 Container Manager"
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  start       Start the development container"
    echo "  stop        Stop the container"
    echo "  restart     Restart the container"
    echo "  shell       Open shell in container"
    echo "  status      Show container status"
    echo "  logs        Show container logs"
    echo "  run         Build on Mac and run in container"
    echo "  delete      Delete the container"
    echo "  recreate    Delete and recreate container"
    echo ""
    echo "Examples:"
    echo "  $0 start         # Start container"
    echo "  $0 shell         # Open shell"
    echo "  $0 run           # Build & test"
    echo ""
}

check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}✗ Docker is not running${NC}"
        echo "Please start Docker Desktop and try again"
        exit 1
    fi
}

container_exists() {
    docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
}

container_running() {
    docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
}

cmd_start() {
    check_docker
    
    if ! container_exists; then
        echo -e "${YELLOW}Container doesn't exist. Creating...${NC}"
        ./docker-amd64-dev.sh
        return
    fi
    
    if container_running; then
        echo -e "${GREEN}✓ Container is already running${NC}"
        return
    fi
    
    echo -e "${BLUE}Starting container...${NC}"
    docker start ${CONTAINER_NAME}
    echo -e "${GREEN}✓ Container started${NC}"
}

cmd_stop() {
    check_docker
    
    if ! container_exists; then
        echo -e "${YELLOW}Container doesn't exist${NC}"
        return
    fi
    
    if ! container_running; then
        echo -e "${YELLOW}Container is not running${NC}"
        return
    fi
    
    echo -e "${BLUE}Stopping container...${NC}"
    docker stop ${CONTAINER_NAME}
    echo -e "${GREEN}✓ Container stopped${NC}"
}

cmd_restart() {
    cmd_stop
    cmd_start
}

cmd_shell() {
    check_docker
    
    if ! container_exists; then
        echo -e "${YELLOW}Container doesn't exist. Creating...${NC}"
        ./docker-amd64-dev.sh
        return
    fi
    
    if ! container_running; then
        echo -e "${BLUE}Starting container...${NC}"
        docker start ${CONTAINER_NAME} > /dev/null
    fi
    
    echo -e "${BLUE}Opening shell...${NC}"
    docker exec -it ${CONTAINER_NAME} bash
}

cmd_status() {
    check_docker
    
    echo -e "${BLUE}Container Status${NC}"
    echo ""
    
    if ! container_exists; then
        echo -e "${YELLOW}Container: Not created${NC}"
        echo ""
        echo "Create with: $0 start"
        return
    fi
    
    if container_running; then
        echo -e "${GREEN}Container: Running${NC}"
    else
        echo -e "${YELLOW}Container: Stopped${NC}"
    fi
    
    echo ""
    docker ps -a --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

cmd_logs() {
    check_docker
    
    if ! container_exists; then
        echo -e "${YELLOW}Container doesn't exist${NC}"
        return
    fi
    
    docker logs ${CONTAINER_NAME}
}

cmd_run() {
    ./docker-build-and-test.sh
}

cmd_delete() {
    check_docker
    
    if ! container_exists; then
        echo -e "${YELLOW}Container doesn't exist${NC}"
        return
    fi
    
    echo -e "${RED}This will delete the container${NC}"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo ""
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled"
        return
    fi
    
    if container_running; then
        echo -e "${BLUE}Stopping container...${NC}"
        docker stop ${CONTAINER_NAME} > /dev/null
    fi
    
    echo -e "${BLUE}Deleting container...${NC}"
    docker rm ${CONTAINER_NAME}
    echo -e "${GREEN}✓ Container deleted${NC}"
}

cmd_recreate() {
    cmd_delete
    echo ""
    ./docker-amd64-dev.sh
}

# Main command router
COMMAND="${1:-}"

case "$COMMAND" in
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    shell|sh)
        cmd_shell
        ;;
    status|st)
        cmd_status
        ;;
    logs)
        cmd_logs
        ;;
    run)
        cmd_run
        ;;
    delete|rm)
        cmd_delete
        ;;
    recreate)
        cmd_recreate
        ;;
    help|--help|-h|"")
        show_usage
        ;;
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac
