package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
)

// APIProxy is an implementation of the proxier interface that is used to
// forward the request to Vault and get the response.
type APIProxy struct {
	client               *api.Client
	logger               hclog.Logger
	RequireRequestHeader bool
}

type APIProxyConfig struct {
	Client               *api.Client
	Logger               hclog.Logger
	RequireRequestHeader bool
}

func NewAPIProxy(config *APIProxyConfig) (Proxier, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("nil API client")
	}
	return &APIProxy{
		client:               config.Client,
		logger:               config.Logger,
		RequireRequestHeader: config.RequireRequestHeader,
	}, nil
}

const (
	vaultRequestHeader = "Vault-Request"
	preconditionFailed = "Precondition Failed"
)

func (ap *APIProxy) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {

	if ap.RequireRequestHeader {
		// check for the required request header
		val, ok := req.Request.Header[vaultRequestHeader]
		if !ok || !reflect.DeepEqual(val, []string{"true"}) {
			return &SendResponse{
					Response: &api.Response{
						Response: &http.Response{
							StatusCode: http.StatusPreconditionFailed,
							Header:     http.Header{},
							Body: ioutil.NopCloser(bytes.NewReader(
								[]byte(preconditionFailed))),
							Request: req.Request,
						},
					},
				},
				errors.New(preconditionFailed)
		}

		// remove the required request header from the request
		delete(req.Request.Header, vaultRequestHeader)
	}

	client, err := ap.client.Clone()
	if err != nil {
		return nil, err
	}
	client.SetToken(req.Token)

	// http.Transport will transparently request gzip and decompress the response, but only if
	// the client doesn't manually set the header. Removing any Accept-Encoding header allows the
	// transparent compression to occur.
	req.Request.Header.Del("Accept-Encoding")
	client.SetHeaders(req.Request.Header)

	fwReq := client.NewRequest(req.Request.Method, req.Request.URL.Path)
	fwReq.BodyBytes = req.RequestBody

	query := req.Request.URL.Query()
	if len(query) != 0 {
		fwReq.Params = query
	}

	// Make the request to Vault and get the response
	ap.logger.Info("forwarding request", "method", req.Request.Method, "path", req.Request.URL.Path)

	resp, err := client.RawRequestWithContext(ctx, fwReq)
	if resp == nil && err != nil {
		// We don't want to cache nil responses, so we simply return the error
		return nil, err
	}

	// Before error checking from the request call, we'd want to initialize a SendResponse to
	// potentially return
	sendResponse, newErr := NewSendResponse(resp, nil)
	if newErr != nil {
		return nil, newErr
	}

	// Bubble back the api.Response as well for error checking/handling at the handler layer.
	return sendResponse, err
}
