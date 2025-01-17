package handlers

import (
	"context"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
)

func HandleSample(f *proxy.Flow, c CryptoClient) int {

	key := "key"
	ctx := context.Background()

	unencryptedBytes, _ := c.DecryptBytes(ctx,
		key,
		f.Response.Body)

	f.Response.Body = unencryptedBytes

	return len(unencryptedBytes)
}
