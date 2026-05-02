package xraygeo

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

type Entry struct {
	Tag      string
	Prefixes []string
}

func Build(entries []Entry) (*routercommon.GeoIPList, error) {
	list := &routercommon.GeoIPList{}
	for _, entry := range entries {
		if entry.Tag == "" {
			return nil, fmt.Errorf("empty xray geoip tag")
		}
		geoip := &routercommon.GeoIP{
			CountryCode: entry.Tag,
			Code:        entry.Tag,
		}
		for _, raw := range entry.Prefixes {
			prefix, err := netip.ParsePrefix(raw)
			if err != nil {
				return nil, fmt.Errorf("tag %s: parse prefix %q: %w", entry.Tag, raw, err)
			}
			if !prefix.Addr().Is4() {
				continue
			}
			addr := prefix.Addr().As4()
			geoip.Cidr = append(geoip.Cidr, &routercommon.CIDR{
				Ip:     addr[:],
				Prefix: uint32(prefix.Bits()),
				IpAddr: prefix.Addr().String(),
			})
		}
		list.Entry = append(list.Entry, geoip)
	}
	return list, nil
}

func Marshal(entries []Entry) ([]byte, error) {
	list, err := Build(entries)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(list)
}

func WriteFile(path string, entries []Entry) error {
	data, err := Marshal(entries)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
