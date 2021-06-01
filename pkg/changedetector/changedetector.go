package changedetector

import (
	"bytes"
	"crypto/sha1"
	"io"
)

// detects changes in content given by it. the first "iteration" is always regarded as changed
type Detector struct {
	previousDigest []byte
}

func New() *Detector {
	return &Detector{}
}

func (c *Detector) ReaderChanged(content io.Reader) (bool, error) {
	return c.WriterChanged(func(sink io.Writer) error {
		_, err := io.Copy(sink, content)
		return err
	})
}

func (c *Detector) WriterChanged(produce func(sink io.Writer) error) (bool, error) {
	digest := sha1.New() // change detection doesn't need to be cryptographic strength

	if err := produce(digest); err != nil {
		return false, err
	}

	newDigest := digest.Sum(nil)

	defer func() {
		c.previousDigest = newDigest
	}()

	changed := !bytes.Equal(newDigest, c.previousDigest)

	return changed, nil
}
