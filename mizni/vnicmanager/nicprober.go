/*
Copyright 2019 The Alkaid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package vnicmanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// default wait time for nic appearing
const defaultWaitTime = time.Second * 3

type nicProberWithTimeout struct {
	timeout time.Duration
}

func (n nicProberWithTimeout) DeviceReady(name, nsPath string) error {
	// most of the time, nic should be ready inside ns
	if err := findNicInNS(name, nsPath); err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return fmt.Errorf("unexpected error: %v", err)
		}

		// rarely need to wait till nic appears, if nic not found
		if n.timeout == 0 {
			n.timeout = defaultWaitTime
		}
		ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
		defer cancel()
		return waitforNicInNS(ctx, name, nsPath)
	}

	return nil
}

func findNicInNS(name, nsPath string) error {
	return ns.WithNetNSPath(nsPath, func(nsOrig ns.NetNS) error {
		_, err := netlink.LinkByName(name)
		return err
	})
}

func waitforNicInNS(ctx context.Context, name, nsPath string) error {
	ns, err := netns.GetFromPath(nsPath)
	if err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}

	updates := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	defer close(done)

	if err := netlink.LinkSubscribeAt(ns, updates, done); err != nil {
		close(updates)
		return fmt.Errorf("unexpected error: %v", err)
	}

	if err := findNicInNS(name, nsPath); err == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return errors.New("unexpected closure of netlink communication")
			}
			if name != update.Link.Attrs().Name {
				break
			}
			return nil
		}
	}
}
