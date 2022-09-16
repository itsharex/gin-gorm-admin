package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/dot123/gin-gorm-admin/internal/config"
	"github.com/dot123/gin-gorm-admin/internal/ginx"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
)

func CopyBodyMiddleware(skippers ...SkipperFunc) gin.HandlerFunc {
	var maxMemory int64 = 64 << 20 // 64 MB
	if v := config.C.HTTP.MaxContentLength; v > 0 {
		maxMemory = v
	}

	return func(c *gin.Context) {
		if SkipHandler(c, skippers...) || c.Request.Body == nil {
			c.Next()
			return
		}

		var requestBody []byte
		isGzip := false
		safe := &io.LimitedReader{R: c.Request.Body, N: maxMemory}

		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(safe)
			if err == nil {
				isGzip = true
				requestBody, _ = ioutil.ReadAll(reader)
			}
		}

		if !isGzip {
			requestBody, _ = ioutil.ReadAll(safe)
		}

		c.Request.Body.Close()
		bf := bytes.NewBuffer(requestBody)
		c.Request.Body = http.MaxBytesReader(c.Writer, ioutil.NopCloser(bf), maxMemory)
		c.Set(ginx.ReqBodyKey, requestBody)

		c.Next()
	}
}