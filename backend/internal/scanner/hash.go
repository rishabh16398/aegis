package scanner

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type fileHashes struct {
	MD5    string
	SHA256 string
}

// computeHashes reads the file once and produces both hashes in one pass.
func computeHashes(path string) (fileHashes, error) {
	f, err := os.Open(path)
	if err != nil {
		return fileHashes{}, err
	}
	defer f.Close()

	md5h := md5.New()
	sha256h := sha256.New()

	if _, err := io.Copy(io.MultiWriter(md5h, sha256h), f); err != nil {
		return fileHashes{}, err
	}

	return fileHashes{
		MD5:    fmt.Sprintf("%x", md5h.Sum(nil)),
		SHA256: fmt.Sprintf("%x", sha256h.Sum(nil)),
	}, nil
}
