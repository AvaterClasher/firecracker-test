package main

import (
	"context"
	"log"
	"net"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

type runningFirecracker struct {
	vmmCtx    context.Context
	vmmCancel context.CancelFunc
	vmmID     string
	machine   *firecracker.Machine
	ip        net.IP
}

func main() {
	ctx := context.Background()

	log.Println("Attempting to spawn a Firecracker VM...")

	vm, err := createAndStartVM(ctx)
	if err != nil {
		log.Fatalf("Failed to spawn Firecracker VM: %v", err)
	}
	defer vm.shutDown()

	log.Printf("Firecracker VM spawned successfully with ID: %s and IP: %s\n", vm.vmmID, vm.ip)
}
