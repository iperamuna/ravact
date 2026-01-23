#!/bin/bash
# VM Manager - Manage all ravact VMs
# Quick commands for common VM operations

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ARM64_VM="ravact-dev"
AMD64_VM="ravact-dev-amd64"

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

show_usage() {
    echo "VM Manager - Manage ravact development VMs"
    echo ""
    echo "Usage: $0 <command> [vm]"
    echo ""
    echo "Commands:"
    echo "  list              List all ravact VMs"
    echo "  info [vm]         Show VM information"
    echo "  start [vm]        Start VM(s)"
    echo "  stop [vm]         Stop VM(s)"
    echo "  restart [vm]      Restart VM(s)"
    echo "  shell [vm]        Open shell in VM"
    echo "  ssh [vm]          SSH to VM"
    echo "  run [vm]          Run ravact on VM"
    echo "  sync [vm]         Sync code to VM(s)"
    echo "  clean             Stop and delete all VMs"
    echo "  status            Show status of all VMs"
    echo ""
    echo "VMs:"
    echo "  arm64 | arm       ARM64 VM (native)"
    echo "  amd64 | amd       AMD64 VM (emulated)"
    echo "  all               All VMs (default)"
    echo ""
    echo "Examples:"
    echo "  $0 list"
    echo "  $0 start arm64"
    echo "  $0 shell amd64"
    echo "  $0 sync all"
    echo "  $0 run arm64"
}

get_vm_name() {
    case "$1" in
        arm64|arm)
            echo "$ARM64_VM"
            ;;
        amd64|amd|x86|x86_64)
            echo "$AMD64_VM"
            ;;
        all|"")
            echo "$ARM64_VM $AMD64_VM"
            ;;
        *)
            echo "$1"
            ;;
    esac
}

vm_exists() {
    multipass list 2>/dev/null | grep -q "$1"
}

list_vms() {
    print_header "Ravact VMs"
    
    if vm_exists "$ARM64_VM"; then
        echo -e "${GREEN}✓ $ARM64_VM (ARM64 - Native)${NC}"
        multipass info "$ARM64_VM" | grep -E "State|IPv4|CPUs|Memory|Disk"
        echo ""
    else
        echo -e "${YELLOW}✗ $ARM64_VM (ARM64) - Not created${NC}"
        echo "  Create with: ./scripts/setup-multipass.sh"
        echo ""
    fi
    
    if vm_exists "$AMD64_VM"; then
        echo -e "${GREEN}✓ $AMD64_VM (AMD64 - Emulated)${NC}"
        multipass info "$AMD64_VM" | grep -E "State|IPv4|CPUs|Memory|Disk"
        echo ""
    else
        echo -e "${YELLOW}✗ $AMD64_VM (AMD64) - Not created${NC}"
        echo "  Create with: ./scripts/setup-multipass-amd64.sh"
        echo ""
    fi
}

show_status() {
    print_header "VM Status"
    multipass list | grep -E "Name|ravact" || echo "No ravact VMs found"
}

start_vms() {
    local vms=$(get_vm_name "$1")
    print_header "Starting VMs"
    
    for vm in $vms; do
        if vm_exists "$vm"; then
            echo -e "${BLUE}Starting $vm...${NC}"
            multipass start "$vm"
            echo -e "${GREEN}✓ $vm started${NC}"
            echo ""
        else
            echo -e "${YELLOW}✗ $vm does not exist${NC}"
        fi
    done
}

stop_vms() {
    local vms=$(get_vm_name "$1")
    print_header "Stopping VMs"
    
    for vm in $vms; do
        if vm_exists "$vm"; then
            echo -e "${BLUE}Stopping $vm...${NC}"
            multipass stop "$vm"
            echo -e "${GREEN}✓ $vm stopped${NC}"
            echo ""
        else
            echo -e "${YELLOW}✗ $vm does not exist${NC}"
        fi
    done
}

restart_vms() {
    local vms=$(get_vm_name "$1")
    print_header "Restarting VMs"
    
    for vm in $vms; do
        if vm_exists "$vm"; then
            echo -e "${BLUE}Restarting $vm...${NC}"
            multipass restart "$vm"
            echo -e "${GREEN}✓ $vm restarted${NC}"
            echo ""
        else
            echo -e "${YELLOW}✗ $vm does not exist${NC}"
        fi
    done
}

shell_vm() {
    local vm=$(get_vm_name "$1" | awk '{print $1}')
    
    if vm_exists "$vm"; then
        echo -e "${BLUE}Opening shell in $vm...${NC}"
        multipass shell "$vm"
    else
        echo -e "${RED}✗ $vm does not exist${NC}"
        exit 1
    fi
}

ssh_vm() {
    local vm=$(get_vm_name "$1" | awk '{print $1}')
    
    if vm_exists "$vm"; then
        echo -e "${BLUE}Connecting via SSH to $vm...${NC}"
        ssh "$vm"
    else
        echo -e "${RED}✗ $vm does not exist${NC}"
        exit 1
    fi
}

run_ravact() {
    local vm=$(get_vm_name "$1" | awk '{print $1}')
    
    if vm_exists "$vm"; then
        echo -e "${BLUE}Running ravact on $vm...${NC}"
        echo ""
        multipass exec "$vm" -- sudo /home/ubuntu/ravact-go/ravact
    else
        echo -e "${RED}✗ $vm does not exist${NC}"
        exit 1
    fi
}

sync_vms() {
    local target="$1"
    
    if [[ "$target" == "all" ]] || [[ -z "$target" ]]; then
        echo -e "${BLUE}Syncing to all VMs...${NC}"
        "$(dirname "$0")/sync-all-vms.sh"
    elif [[ "$target" == "arm64" ]] || [[ "$target" == "arm" ]]; then
        echo -e "${BLUE}Syncing to ARM64 VM...${NC}"
        if [[ -f "$HOME/.ravact-multipass-sync.sh" ]]; then
            "$HOME/.ravact-multipass-sync.sh"
        else
            echo -e "${RED}✗ Sync script not found${NC}"
            exit 1
        fi
    elif [[ "$target" == "amd64" ]] || [[ "$target" == "amd" ]]; then
        echo -e "${BLUE}Syncing to AMD64 VM...${NC}"
        if [[ -f "$HOME/.ravact-multipass-amd64-sync.sh" ]]; then
            "$HOME/.ravact-multipass-amd64-sync.sh"
        else
            echo -e "${RED}✗ Sync script not found${NC}"
            exit 1
        fi
    else
        echo -e "${RED}✗ Unknown VM: $target${NC}"
        exit 1
    fi
}

show_info() {
    local vms=$(get_vm_name "$1")
    print_header "VM Information"
    
    for vm in $vms; do
        if vm_exists "$vm"; then
            echo -e "${GREEN}$vm${NC}"
            multipass info "$vm"
            echo ""
        else
            echo -e "${YELLOW}✗ $vm does not exist${NC}"
            echo ""
        fi
    done
}

clean_vms() {
    print_header "Clean All VMs"
    echo -e "${RED}This will DELETE all ravact VMs!${NC}"
    echo ""
    read -p "Are you sure? (type 'yes' to confirm): " confirm
    
    if [[ "$confirm" != "yes" ]]; then
        echo "Cancelled"
        exit 0
    fi
    
    if vm_exists "$ARM64_VM"; then
        echo -e "${BLUE}Deleting $ARM64_VM...${NC}"
        multipass delete "$ARM64_VM"
        echo -e "${GREEN}✓ $ARM64_VM deleted${NC}"
    fi
    
    if vm_exists "$AMD64_VM"; then
        echo -e "${BLUE}Deleting $AMD64_VM...${NC}"
        multipass delete "$AMD64_VM"
        echo -e "${GREEN}✓ $AMD64_VM deleted${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Purging deleted VMs...${NC}"
    multipass purge
    echo -e "${GREEN}✓ All VMs cleaned${NC}"
}

# Main command router
COMMAND="${1:-}"
VM_TARGET="${2:-all}"

case "$COMMAND" in
    list|ls)
        list_vms
        ;;
    status|st)
        show_status
        ;;
    info)
        show_info "$VM_TARGET"
        ;;
    start)
        start_vms "$VM_TARGET"
        ;;
    stop)
        stop_vms "$VM_TARGET"
        ;;
    restart)
        restart_vms "$VM_TARGET"
        ;;
    shell|sh)
        shell_vm "$VM_TARGET"
        ;;
    ssh)
        ssh_vm "$VM_TARGET"
        ;;
    run)
        run_ravact "$VM_TARGET"
        ;;
    sync)
        sync_vms "$VM_TARGET"
        ;;
    clean|delete)
        clean_vms
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
