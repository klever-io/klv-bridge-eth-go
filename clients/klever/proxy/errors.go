package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrInvalidAddress signals that the provided address is invalid
var ErrInvalidAddress = errors.New("invalid address")

// ErrNilAddress signals that the provided address is nil
var ErrNilAddress = errors.New("nil address")

// ErrNilNetworkConfigs signals that the provided network configs is nil
var ErrNilNetworkConfigs = errors.New("nil network configs")

// ErrInvalidCacherDuration signals that the provided caching duration is invalid
var ErrInvalidCacherDuration = errors.New("invalid caching duration")

// ErrInvalidAllowedDeltaToFinal signals that an invalid allowed delta to final value has been provided
var ErrInvalidAllowedDeltaToFinal = errors.New("invalid allowed delta to final value")

// ErrNilHTTPClientWrapper signals that a nil HTTP client wrapper was provided
var ErrNilHTTPClientWrapper = errors.New("nil HTTP client wrapper")

// ErrHTTPStatusCodeIsNotOK signals that the returned HTTP status code is not OK
var ErrHTTPStatusCodeIsNotOK = errors.New("HTTP status code is not OK")

// ErrNilEndpointProvider signals that a nil endpoint provider was provided
var ErrNilEndpointProvider = errors.New("nil endpoint provider")

// ErrInvalidEndpointProvider signals that an invalid endpoint provider was provided
var ErrInvalidEndpointProvider = errors.New("invalid endpoint provider")

// ErrNilNetworkStatus signals that nil network status was received
var ErrNilNetworkStatus = errors.New("nil network status")

// ErrNilRequest signals that a nil request was provided
var ErrNilRequest = errors.New("nil request")

// ErrNilProxy signals that a nil proxy has been provided
var ErrNilProxy = errors.New("nil proxy")

// ErrNotUint64Bytes signals that the provided bytes do not represent a valid uint64 number
var ErrNotUint64Bytes = errors.New("provided bytes do not represent a valid uint64 number")

// GenericAPIResponse represents the generic response structure from the Klever node/proxy
type GenericAPIResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  string      `json:"code"`
}

func createHTTPStatusError(httpStatusCode int, err error) error {
	if err == nil {
		err = ErrHTTPStatusCodeIsNotOK
	}

	return fmt.Errorf("%w, returned http status: %d, %s",
		err, httpStatusCode, http.StatusText(httpStatusCode))
}

func createHTTPStatusErrorWithBody(httpStatusCode int, err error, responseBody []byte) error {
	if err == nil {
		err = ErrHTTPStatusCodeIsNotOK
	}

	apiError := extractAPIError(responseBody)
	if apiError != "" {
		return fmt.Errorf("%w, returned http status: %d, %s, api error: %s",
			err, httpStatusCode, http.StatusText(httpStatusCode), apiError)
	}

	return fmt.Errorf("%w, returned http status: %d, %s",
		err, httpStatusCode, http.StatusText(httpStatusCode))
}

func extractAPIError(responseBody []byte) string {
	if len(responseBody) == 0 {
		return ""
	}

	var response GenericAPIResponse
	if jsonErr := json.Unmarshal(responseBody, &response); jsonErr != nil {
		return ""
	}

	return response.Error
}
