package xraygeo

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

func TestMarshal(t *testing.T) {
	data, err := Marshal([]Entry{
		{Tag: "ru-providers", Prefixes: []string{"10.0.0.0/24", "10.0.1.0/24"}},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var list routercommon.GeoIPList
	if err := proto.Unmarshal(data, &list); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
	if len(list.Entry) != 1 {
		t.Fatalf("entries = %d", len(list.Entry))
	}
	if list.Entry[0].GetCode() != "ru-providers" {
		t.Fatalf("code = %q", list.Entry[0].GetCode())
	}
	if len(list.Entry[0].GetCidr()) != 2 {
		t.Fatalf("cidr len = %d", len(list.Entry[0].GetCidr()))
	}
}
