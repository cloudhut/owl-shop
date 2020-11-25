package fake

import (
	"github.com/brianvoe/gofakeit/v5"
	"github.com/mroth/weightedrand"
	"net/http"
)

type FrontendEvent struct {
	// VersionedStruct
	Version int `json:"version"`

	RequestedURL    string                `json:"requestedUrl"`
	Method          string                `json:"method"`
	CorrelationID   string                `json:"correlationId"`
	IPAddress       string                `json:"ipAddress"`
	RequestDuration int                   `json:"requestDuration"`
	Response        FrontendEventResponse `json:"response"`
	Headers         map[string]string     `json:"headers"`
}

type FrontendEventResponse struct {
	Size       int `json:"size"`
	StatusCode int `json:"statusCode"`
}

func NewFrontendEvent() FrontendEvent {
	return FrontendEvent{
		Version:         0,
		RequestedURL:    gofakeit.URL(),
		Method:          gofakeit.HTTPMethod(),
		CorrelationID:   gofakeit.UUID(),
		IPAddress:       gofakeit.IPv4Address(),
		RequestDuration: gofakeit.Number(1, 1500),
		Response: FrontendEventResponse{
			Size:       gofakeit.Number(40, 2500),
			StatusCode: newStatusCode(),
		},
		Headers: newHTTPHeaders(),
	}
}

func newHTTPHeaders() map[string]string {
	return map[string]string{
		"user-agent":      gofakeit.UserAgent(),
		"accept":          "*/*",
		"accept-encoding": "gzip",
		"cache-control":   "max-age=0",
		"origin":          gofakeit.URL(),
		"referrer":        gofakeit.URL(),
	}
}

// newStatusCode returns a weighted status code
func newStatusCode() int {
	c, err := weightedrand.NewChooser(
		weightedrand.Choice{Item: http.StatusOK, Weight: 940},
		weightedrand.Choice{Item: http.StatusMovedPermanently, Weight: 10},
		weightedrand.Choice{Item: http.StatusInternalServerError, Weight: 5},
		weightedrand.Choice{Item: http.StatusServiceUnavailable, Weight: 5},
		weightedrand.Choice{Item: http.StatusNotFound, Weight: 40},
		weightedrand.Choice{Item: http.StatusBadRequest, Weight: 2},
	)
	if err != nil {
		panic(err)
	}
	statusCode := c.Pick().(int)
	return statusCode
}
