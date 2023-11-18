package processes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"sreda/internal/config"
	"sreda/internal/models"

	"golang.org/x/sync/errgroup"
)

const requestPath = "/api/request"

// MySender структура для хранения данных отправителя.
type MySender struct {
	log    *slog.Logger
	config *config.Config
	client *http.Client

	path string
}

func NewSender(log *slog.Logger, config *config.Config) *MySender {
	return &MySender{
		log:    log,
		config: config,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		path: config.URL + requestPath,
	}
}

// sendRequest отправляет POST-запрос на указанный URL с заданным телом.
func (s *MySender) sendRequest(ctx context.Context, iteration int64) error {
	item := models.IterationEntry{Iteration: iteration}
	jsonBody, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error marshalling iteration %d: %v", iteration, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.path, bytes.NewBufferString(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("error creating request %d: %v", iteration, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request %d: %v", iteration, err)
	}
	defer resp.Body.Close()

	s.log.Debug("Request sent",
		"iteration", iteration,
		"status_code", resp.StatusCode,
		"time", time.Now().String(),
	)

	if ctx.Err() != nil {
		return fmt.Errorf("request canceled for iteration %d: %v", iteration, ctx.Err())
	}

	return nil
}

// Run запускает отправку запросов в соответствии с конфигурацией.
func (s *MySender) Run(inCtx context.Context) error {
	s.log.Debug("Run iterations",
		"amount", s.config.Requests.Amount,
		"per_second", s.config.Requests.PerSecond,
	)

	var (
		g, ctx = errgroup.WithContext(inCtx)
		wg     sync.WaitGroup
	)

	// Ограничиваем количество одновременно выполняемых запросов.
	semaphore := make(chan struct{}, s.config.Requests.PerSecond)

	// Создаем тикер для управления частотой запросов.
	ticker := time.NewTicker(time.Second / time.Duration(s.config.Requests.PerSecond))
	defer ticker.Stop()

	for i := int64(1); i <= s.config.Requests.Amount; i++ {
		wg.Add(1)

		g.Go(func(iteration int64) func() error {
			return func() error {
				defer wg.Done()

				// Ожидаем тикер перед отправкой запроса.
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return ctx.Err()
				}

				// Отправляем запрос.
				return s.sendRequest(ctx, iteration)
			}
		}(i))
	}

	// Запускаем горутину, которая будет ждать завершения всех запросов.
	go func() {
		wg.Wait()
		close(semaphore)
	}()

	return g.Wait()
}
