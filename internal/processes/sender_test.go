package processes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"sreda/internal/config"
	"sreda/internal/lib/logger/sl"

	"github.com/stretchr/testify/assert"
)

func TestSender_Run(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// env := "local"
	env := "nop"
	noplog := sl.SetupLogger(env)

	cfg := &config.Config{
		Env:     env,
		Version: "1.0.0",
		URL:     server.URL,
		Requests: struct {
			Amount    int64 `yaml:"amount"`
			PerSecond int64 `yaml:"per_second"`
		}{
			Amount:    10,
			PerSecond: 1,
		},
	}

	sender := NewSender(noplog, cfg)
	ctx := context.Background()
	err := sender.Run(ctx)
	assert.NoError(t, err)
}

func TestSender_RunCount(t *testing.T) {
	requestCount := int64(0)
	allCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Увеличиваем счетчик запросов.
		atomic.AddInt64(&requestCount, 1)
		atomic.AddInt64(&allCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// env := "local"
	env := "nop"
	log := sl.SetupLogger(env)
	cfg := &config.Config{
		Env:     env,
		Version: "1.0.0",
		URL:     server.URL,
		Requests: struct {
			Amount    int64 `yaml:"amount"`
			PerSecond int64 `yaml:"per_second"`
		}{
			Amount:    100,
			PerSecond: 10,
		},
	}

	sender := NewSender(log, cfg)

	// Создаем канал для сигнала завершения.
	done := make(chan struct{})
	defer close(done)

	ctx := context.Background()

	// Запускаем Sender в отдельной горутине
	go func() {
		err := sender.Run(ctx)
		assert.NoError(t, err)
	}()

	// Ожидаем, пока Sender выполнит все запросы или превысит лимит Amount.
	for {
		select {
		case <-time.After(time.Second):
			// Проверяем, что количество запросов не больше, чем PerSecond в секунду (+1 для погрешности).
			assert.LessOrEqual(t, atomic.LoadInt64(&requestCount), cfg.Requests.PerSecond+1)
			log.Debug("requestCount", "requestCount", atomic.LoadInt64(&requestCount))

			// Если все запросы были отправлены, завершаем цикл
			if atomic.LoadInt64(&allCount) >= cfg.Requests.Amount {
				return
			}

			// Сбрасываем requestCount
			atomic.StoreInt64(&requestCount, 0)

		case <-done:
			return
		}
	}
}
