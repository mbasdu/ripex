package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var RequiredFiles = []string{
	"ripe.db.organisation.gz",
	"ripe.db.inetnum.gz",
	"ripe.db.aut-num.gz",
	"ripe.db.route.gz",
}

type Config struct {
	BaseURL   string
	CacheDir  string
	UserAgent string
	Timeout   time.Duration
}

func Run(cfg Config) error {
	if err := os.MkdirAll(cfg.CacheDir, 0o755); err != nil {
		return err
	}

	client := &http.Client{Timeout: cfg.Timeout}
	for _, name := range RequiredFiles {
		if err := downloadFile(client, cfg, name); err != nil {
			return err
		}
	}
	return nil
}

func downloadFile(client *http.Client, cfg Config, name string) error {
	url := stringsTrimRight(cfg.BaseURL, "/") + "/" + name
	dst := filepath.Join(cfg.CacheDir, name)
	tmp := dst + ".tmp"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", name, resp.Status)
	}

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}

	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func stringsTrimRight(s, cutset string) string {
	for len(s) > 0 && len(cutset) > 0 {
		last := s[len(s)-1]
		match := false
		for i := 0; i < len(cutset); i++ {
			if byte(cutset[i]) == last {
				match = true
				break
			}
		}
		if !match {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}
