package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

func copy(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dst, data, 0644)
	return err
}

func (vm runningFirecracker) shutDown() {
	log.WithField("ip", vm.ip).Info("stopping")
	vm.machine.StopVMM()
	err := os.Remove(vm.machine.Cfg.SocketPath)
	if err != nil {
		log.WithError(err).Error("Failed to delete firecracker socket")
	}
	err = os.Remove("/tmp/rootfs-" + vm.vmmID + ".ext4")
	if err != nil {
		log.WithError(err).Error("Failed to delete firecracker rootfs")
	}
}

func createAndStartVM(ctx context.Context) (*runningFirecracker, error) {
	vmmID := xid.New().String()

	copy("./rootfs/rootfs.ext4", "/tmp/rootfs-"+vmmID+".ext4")

	fcCfg, err := getFirecrackerConfig(vmmID)
	if err != nil {
		log.Errorf("Error: %s", err)
		return nil, err
	}
	logger := log.New()

	if false {
		log.SetLevel(log.DebugLevel)
		logger.SetLevel(log.DebugLevel)
	}

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.NewEntry(logger)),
	}

	firecrackerBinary, err := exec.LookPath("firecracker")
	if err != nil {
		return nil, err
	}

	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("binary %q does not exist: %v", firecrackerBinary, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return nil, fmt.Errorf("binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&0111 == 0 {
		return nil, fmt.Errorf("binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}

	vmmCtx, vmmCancel := context.WithCancel(ctx)

	m, err := firecracker.NewMachine(vmmCtx, fcCfg, machineOpts...)
	if err != nil {
		vmmCancel()
		return nil, fmt.Errorf("failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		vmmCancel()
		return nil, fmt.Errorf("failed to start machine: %v", err)
	}

	log.WithField("ip", m.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP).Info("machine started")

	return &runningFirecracker{
		vmmCtx:    vmmCtx,
		vmmCancel: vmmCancel,
		vmmID:     vmmID,
		machine:   m,
		ip:        m.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP,
	}, nil
}
