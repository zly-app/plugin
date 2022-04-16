package compress

import (
	"io"

	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zly-app/plugin/honey/config"
)

type Compress interface {
	// 压缩
	Compress(in io.Reader, out io.Writer) error
	// 解压缩
	UnCompress(in io.Reader, out io.Writer) error
}

func MakeCompress(conf *config.Config) Compress {
	switch conf.CompressType {
	case ZStdCompressName:
		return NewZStdCompress()
	}

	logger.Log.Fatal("honey压缩程序类型未定义", zap.String("CompressType", conf.CompressType))
	return nil
}
