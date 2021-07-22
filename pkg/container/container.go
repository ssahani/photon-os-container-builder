// SPDX-License-Identifier: BSD-2
// Copyright 2021 VMware, Inc.

package container

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/photon-os-container-builder/pkg/conf"
	"github.com/photon-os-container-builder/pkg/nspawn"
	"github.com/photon-os-container-builder/pkg/rpm"
	"github.com/photon-os-container-builder/pkg/set"
	"github.com/photon-os-container-builder/pkg/system"
	"github.com/photon-os-container-builder/pkg/systemd"
)

func Spawn(base string, c string, release string, packages string, dir bool) error {
	d := path.Join(base, c)

	if err := system.CreateDirectory(base, c); err != nil {
		fmt.Printf("Failed to create container image dir: %+v\n", err)
		return errors.New("dir exists")
	}

	s := set.New()
	s.AddAll(packages)

	if err := rpm.ConstructOSTree(release, d, s); err != nil {
		defer system.RemoveDir(d)

		fmt.Printf("Failed to construct container root directory '%s': +%v\n", d, err)
		return err
	}

	if err := system.ExecAndDisplay(os.Stdout, "/usr/bin/systemd-machine-id-setup", "--root", d); err != nil {
		fmt.Printf("Failed to execute systemd-machine-id-setup for '%s': %+v\n", c, err)
		return err
	}

	if err := systemd.SetupContainerService(c); err != nil {
		fmt.Printf("Failed to create unit file for '%s': %+v\n", c, err)
		return err
	}

	return nspawn.Spawn(d, dir)
}

func JumpStart(c *conf.Config, base string, container string, link string, ephemeral bool, machine bool, network bool) error {
	dir := path.Join(base, container)

	if !system.PathExists(dir) {
		fmt.Printf("Container '%s' does not exist\n", container)
		return errors.New("not exist")
	}

	return nspawn.ThunderBolt(c, dir, link, ephemeral, machine, network)
}

func Boot(c *conf.Config, storage string, container string, link string, ephemeral bool, network bool) error {
	dir := path.Join(storage, container)

	if !system.PathExists(dir) {
		fmt.Printf("Container '%s' does not exist\n", container)
		return errors.New("not exist")
	}

	return nspawn.Boot(c, dir, link, ephemeral, network)
}
