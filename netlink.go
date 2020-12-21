package main

import (
	"fmt"
	"time"

	"github.com/vishvananda/netlink"
)

// NetlinkLatencySetter implements LatencySetter by calling the tc command.
type NetlinkLatencySetter struct{}

// SetLatency sets extra latency of the given interface
func (t NetlinkLatencySetter) SetLatency(iname string, latency time.Duration) error {
	link, err := netlink.LinkByName(iname)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return fmt.Errorf("list qdisc: %w", err)
	}

	for _, q := range qdiscs {
		if q.Type() == (&netlink.Netem{}).Type() {
			if err := netlink.QdiscDel(
				netlink.NewNetem(
					netlink.QdiscAttrs{
						LinkIndex: link.Attrs().Index,
						Parent:    netlink.HANDLE_ROOT,
					},
					netlink.NetemQdiscAttrs{},
				),
			); err != nil {
				return fmt.Errorf("qdisc del: %w", err)
			}
		}
	}

	if latency <= 0 {
		return nil
	}

	if err := netlink.QdiscAdd(netlink.NewNetem(
		netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.HANDLE_ROOT,
		},
		netlink.NetemQdiscAttrs{
			Latency: uint32(latency.Microseconds()),
		},
	)); err != nil {
		return fmt.Errorf("qdisc add: %w", err)
	}

	return nil
}

func ptrInt64(v int64) *int64 { return &v }
