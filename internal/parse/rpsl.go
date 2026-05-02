package parse

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"ripex/internal/ripe"
)

type Object struct {
	Type       string
	Key        string
	Attributes map[string][]string
}

func ParseGzipFile(path string, wantTypes map[string]struct{}, visit func(Object) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	return ParseRPSL(gz, wantTypes, visit)
}

func ParseRPSL(r io.Reader, wantTypes map[string]struct{}, visit func(Object) error) error {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 8*1024*1024)

	var lines []string
	flush := func() error {
		if len(lines) == 0 {
			return nil
		}
		obj, err := buildObject(lines)
		lines = lines[:0]
		if err != nil {
			return err
		}
		if obj.Type == "" {
			return nil
		}
		if len(wantTypes) > 0 {
			if _, ok := wantTypes[obj.Type]; !ok {
				return nil
			}
		}
		return visit(obj)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}

func buildObject(lines []string) (Object, error) {
	obj := Object{Attributes: make(map[string][]string)}
	var currentKey string

	for _, raw := range lines {
		if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "%") {
			continue
		}

		if len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t' || raw[0] == '+') {
			if currentKey == "" || len(obj.Attributes[currentKey]) == 0 {
				continue
			}
			last := len(obj.Attributes[currentKey]) - 1
			obj.Attributes[currentKey][last] += " " + strings.TrimSpace(raw)
			continue
		}

		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return Object{}, fmt.Errorf("invalid RPSL line: %q", raw)
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		obj.Attributes[key] = append(obj.Attributes[key], value)
		currentKey = key

		if obj.Type == "" {
			obj.Type = key
			obj.Key = value
		}
	}

	return obj, nil
}

func ToOrganisation(obj Object) ripe.Organisation {
	return ripe.Organisation{
		ID:      first(obj.Attributes["organisation"]),
		Name:    first(obj.Attributes["org-name"]),
		Country: first(obj.Attributes["country"]),
	}
}

func ToInetnum(obj Object) ripe.Inetnum {
	return ripe.Inetnum{
		Key:     first(obj.Attributes["inetnum"]),
		OrgID:   first(obj.Attributes["org"]),
		Country: first(obj.Attributes["country"]),
		NetName: first(obj.Attributes["netname"]),
		Status:  first(obj.Attributes["status"]),
	}
}

func ToAutNum(obj Object) ripe.AutNum {
	return ripe.AutNum{
		ASN:    normalizeASN(first(obj.Attributes["aut-num"])),
		AsName: first(obj.Attributes["as-name"]),
		OrgID:  first(obj.Attributes["org"]),
	}
}

func ToRoute(obj Object) ripe.Route {
	return ripe.Route{
		Prefix: first(obj.Attributes["route"]),
		Origin: normalizeASN(first(obj.Attributes["origin"])),
	}
}

func normalizeASN(v string) string {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "AS") {
		return v
	}
	return "AS" + v
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
