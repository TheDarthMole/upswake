package worker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
)

type Pool struct {
	wg      *sync.WaitGroup
	workers []*Worker
}

func NewWorkerPool(ctx context.Context, config *entity.Config, tlsConfig *tls.Config, logger *slog.Logger, url string) (*Pool, error) {
	wg := &sync.WaitGroup{}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	var workers []*Worker

	for _, mapping := range config.NutServers {
		for _, target := range mapping.Targets {
			worker, err := newWorker(ctx, target, client, wg, logger, url)
			if err != nil {
				return &Pool{}, err
			}
			workers = append(workers, worker)
		}
	}

	return &Pool{
		wg:      wg,
		workers: workers,
	}, nil
}

type Worker struct {
	ctx         context.Context
	wg          *sync.WaitGroup
	logger      *slog.Logger
	client      *http.Client
	url         string
	requestBody []byte
	interval    time.Duration
}

func (w *Pool) Start() {
	for _, worker := range w.workers {
		w.wg.Add(1)
		worker.run()
	}
}

func (w *Pool) Wait() {
	w.wg.Wait()
}

func newWorker(ctx context.Context, targetServer *entity.TargetServer, client *http.Client, wg *sync.WaitGroup, logger *slog.Logger, url string) (*Worker, error) {
	jobLogger := logger.With(
		slog.String("type", "serveJob"),
		slog.String("worker_name", targetServer.Name),
	)

	body, err := json.Marshal(map[string]string{"mac": targetServer.MAC})
	if err != nil {
		jobLogger.Error("Error marshalling JSON", slog.Any("error", err))
		return &Worker{}, err
	}

	return &Worker{
		ctx:         ctx,
		client:      client,
		wg:          wg,
		logger:      jobLogger,
		interval:    targetServer.Interval,
		requestBody: body,
		url:         url,
	}, nil
}

func (w *Worker) run() {
	w.logger.Info("Starting worker")

	go func() {
		defer w.wg.Done()

		startupTime := rand.IntN(500-350) + 350 // Stagger initial requests to avoid thundering herd, min 350ms, max 500ms
		time.Sleep(time.Duration(startupTime) * time.Millisecond)

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				w.logger.Info("Gracefully stopping worker")
				return
			case <-ticker.C:
				w.sendWakeRequest()
			}
		}
	}()
}

func (w *Worker) sendWakeRequest() {
	req, err := http.NewRequestWithContext(w.ctx, http.MethodPost, w.url, bytes.NewBuffer(w.requestBody))
	if err != nil {
		w.logger.Error("Error creating HTTP request",
			slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.client.Do(req)

	defer func(resp *http.Response) {
		if resp == nil {
			return
		}
		_, _ = io.Copy(io.Discard, resp.Body) // Drain body to enable connection reuse
		err = resp.Body.Close()
		if err != nil {
			w.logger.Error("Error closing response body",
				slog.Any("error", err))
		}
	}(resp)

	if ctxErr := context.Cause(w.ctx); ctxErr != nil {
		w.logger.Warn("Context cancelled when making request",
			slog.Any("error", ctxErr))
		return
	}

	if err != nil {
		w.logger.Error("Error sending post request",
			slog.Any("error", err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.logger.Error("Unexpected status code from upswake endpoint",
			slog.String("status_code", resp.Status))
		return
	}
}
