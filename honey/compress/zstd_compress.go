package compress

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

const ZStdCompressName = "zstd"

type ZStdCompress struct{}

func (Z *ZStdCompress) Compress(in io.Reader, out io.Writer) error {
	enc, err := zstd.NewWriter(out)
	if err != nil {
		return err
	}
	_, err = io.Copy(enc, in)
	if err != nil {
		_ = enc.Close()
		return err
	}
	return enc.Close()
}

func (Z *ZStdCompress) UnCompress(in io.Reader, out io.Writer) error {
	d, err := zstd.NewReader(in)
	if err != nil {
		return err
	}
	defer d.Close()

	// Copy content...
	_, err = io.Copy(out, d)
	return err
}

func NewZStdCompress() Compress {
	return &ZStdCompress{}
}
