package client

import (
	"github.com/Trendyol/chaki/config"
	"github.com/Trendyol/chaki/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type driverBuilder struct {
	cfg      *config.Config
	eh       ErrDecoder
	d        *resty.Client
	updaters []DriverWrapper
}

func newDriverBuilder(cfg *config.Config) *driverBuilder {
	return &driverBuilder{
		cfg: cfg,
		d:   resty.New().SetBaseURL(cfg.GetString("baseurl")),
	}
}

func (b *driverBuilder) AddErrDecoder(eh ErrDecoder) *driverBuilder {
	b.eh = eh
	return b
}

func (b *driverBuilder) AddUpdaters(wrappers ...DriverWrapper) *driverBuilder {
	b.updaters = append(b.updaters, wrappers...)
	return b
}

func (b *driverBuilder) build() *resty.Client {
	b.useLogging()

	for _, upd := range b.updaters {
		b.d = upd(b.d)
	}

	b.d.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		return b.eh(r.Request.Context(), r)
	})
	return b.d
}

func (b *driverBuilder) useLogging() {
	b.d.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		logger.From(r.Context()).Info(
			"request sent",
			zap.String("method", r.Method),
			zap.String("path", r.URL),
			zap.Any("body", r.Body),
			zap.Any("headers", r.Header),
			zap.Any("query", r.QueryParam),
		)
		return nil
	})

	b.d.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		logger.From(r.Request.Context()).Info(
			"response got",
			zap.String("method", r.Request.Method),
			zap.String("path", r.Request.URL),
			zap.Int("status", r.StatusCode()),
			zap.String("body", string(r.Body())),
		)
		return nil
	})
}
