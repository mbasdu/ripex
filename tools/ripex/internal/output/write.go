package output

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Record struct {
	Prefix           string `json:"prefix"`
	Dataset          string `json:"dataset"`
	SourceObjectType string `json:"source_object_type"`
	SourceKey        string `json:"source_key"`
	OrgID            string `json:"org_id,omitempty"`
	OrgName          string `json:"org_name,omitempty"`
	OrgCountry       string `json:"org_country,omitempty"`
	ASN              string `json:"asn,omitempty"`
	AsName           string `json:"as_name,omitempty"`
	SnapshotDate     string `json:"snapshot_date"`
}

func WriteJSONL(path string, records []Record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, record := range records {
		if err := enc.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func WriteCSV(path string, records []Record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"prefix",
		"dataset",
		"source_object_type",
		"source_key",
		"org_id",
		"org_name",
		"org_country",
		"asn",
		"as_name",
		"snapshot_date",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, record := range records {
		row := []string{
			record.Prefix,
			record.Dataset,
			record.SourceObjectType,
			record.SourceKey,
			record.OrgID,
			record.OrgName,
			record.OrgCountry,
			record.ASN,
			record.AsName,
			record.SnapshotDate,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}

func WriteLines(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data := ""
	if len(lines) > 0 {
		data = strings.Join(lines, "\n") + "\n"
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
