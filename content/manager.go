package content

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/abourget/llerrgroup"
)

type ContentRef struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Hash string `json:"hash"`
}

type Manager struct {
	cachePath string
	logger    *zap.Logger
}

func NewManager(cachePath string) *Manager {
	return &Manager{
		cachePath: cachePath,
		logger:    zap.NewNop(),
	}
}

func (c *Manager) SetLogger(logger *zap.Logger) {
	c.logger = logger
}

func (c *Manager) ensureCacheExists() error {
	return os.MkdirAll(c.cachePath, 0777)
}

func (c *Manager) isInCache(ref string) bool {
	fileName := filepath.Join(c.cachePath, replaceAllWeirdities(ref))

	if _, err := os.Stat(fileName); err == nil {
		return true
	}
	return false
}

func (c *Manager) Download(contentRefs []*ContentRef) error {
	if err := c.ensureCacheExists(); err != nil {
		return fmt.Errorf("error creating cache path: %s", err)
	}

	eg := llerrgroup.New(10)
	for _, contentRef := range contentRefs {
		if eg.Stop() {
			continue
		}

		contentRef := contentRef
		eg.Go(func() error {
			if err := c.downloadURL(contentRef.URL, contentRef.Hash); err != nil {
				return fmt.Errorf("content %q: %s", contentRef.Name, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Manager) downloadURL(ref string, hash string) error {
	if hash != "" && c.isInCache(ref) {
		return nil
	}

	cnt, err := c.downloadRef(ref)
	if err != nil {
		return err
	}

	if hash != "" {
		h := sha256.New()
		_, _ = h.Write(cnt)
		contentHash := hex.EncodeToString(h.Sum(nil))

		if contentHash != hash {
			return fmt.Errorf("hash in boot sequence [%q] not equal to computed hash on downloaded file [%q]", hash, contentHash)
		}
	}

	c.logger.Info("Caching content.", zap.String("ref", ref))
	if err := c.writeToCache(ref, cnt); err != nil {
		return err
	}

	return nil
}

func (c *Manager) downloadRef(ref string) ([]byte, error) {
	c.logger.Debug("Downloading content", zap.String("from", ref))
	if _, err := os.Stat(ref); err == nil {
		return c.downloadLocalFile(ref)
	}

	destURL, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("ref %q is not a valid URL: %s", ref, err)
	}

	switch destURL.Scheme {
	case "file":
		return c.downloadFileURL(destURL)
	case "http", "https":
		return c.downloadHTTPURL(destURL)
	default:
		return nil, fmt.Errorf("don't know how to handle scheme %q (from ref %q)", destURL.Scheme, destURL)
	}
}

func (c *Manager) downloadLocalFile(ref string) ([]byte, error) {
	return ioutil.ReadFile(ref)
}

func (c *Manager) downloadFileURL(destURL *url.URL) ([]byte, error) {
	fmt.Printf("Path %s, Raw path: %s\n", destURL.Path, destURL.RawPath)
	return []byte{}, nil
}

func (c *Manager) downloadHTTPURL(destURL *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", destURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("download attempts failed")
	}
	defer resp.Body.Close()

	cnt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		if len(cnt) > 50 {
			cnt = cnt[:50]
		}
		return nil, fmt.Errorf("couldn't get %s, return code: %d, server error: %q", destURL, resp.StatusCode, cnt)
	}

	return cnt, nil
}

func (c *Manager) writeToCache(ref string, content []byte) error {
	fileName := replaceAllWeirdities(ref)
	return ioutil.WriteFile(filepath.Join(c.cachePath, fileName), content, 0666)
}

func (c *Manager) ReadFromCache(ref string) ([]byte, error) {
	fileName := replaceAllWeirdities(ref)
	return ioutil.ReadFile(filepath.Join(c.cachePath, fileName))
}

func (c *Manager) ReaderFromCache(ref string) (io.ReadCloser, error) {
	fileName := replaceAllWeirdities(ref)
	return os.Open(filepath.Join(c.cachePath, fileName))
}

func (c *Manager) FileNameFromCache(ref string) string {
	fileName := replaceAllWeirdities(ref)
	return filepath.Join(c.cachePath, fileName)
}
