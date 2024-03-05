package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"github.com/shevchukeugeni/gofermart/internal/store"
	"github.com/shevchukeugeni/gofermart/internal/types"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Worker struct {
	logger            *zap.Logger
	db                *sql.DB
	accrualSystemAddr string
	order             store.Order

	client *resty.Client
}

func NewWorker(logger *zap.Logger, db *sql.DB, accrualSystemAddr string, order store.Order, client *resty.Client) *Worker {
	return &Worker{
		logger:            logger.Named("Worker"),
		db:                db,
		accrualSystemAddr: accrualSystemAddr,
		order:             order,
		client:            client,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
		case <-ticker.C:
			w.logger.Info("worker started")
			orders, err := w.order.GetPendingOrdersNumbers(ctx)
			if err != nil {
				w.logger.Error("Unable to get pending orders", zap.Error(err))
			}

		Loop:
			for _, order := range orders {
				res, err := w.requestAccrual(order.Number)
				if err != nil {
					w.logger.Error("Failed accrual system request", zap.Error(err))
					continue
				}

				switch res.StatusCode() {
				case http.StatusNoContent:
					//наверное спросить позже
				case http.StatusTooManyRequests:
					retryAfterStr := res.Header().Get("Retry-After")
					retryAfter, err := strconv.ParseInt(retryAfterStr, 10, 64)
					if err != nil {
						w.logger.Error("unable to get retry time", zap.Error(err))
						break Loop
					}
					ticker = time.NewTicker(time.Duration(retryAfter) * time.Second)
					break Loop
				case http.StatusOK:
					var resp types.AccrualResponse

					err = json.Unmarshal(res.Body(), &resp)
					if err != nil {
						w.logger.Error("unable decode accrual response", zap.Error(err))
						continue
					}

					switch types.Status(resp.Status) {
					case types.Registered, types.Processing:
						continue
					case types.Invalid, types.Processed:
						err = w.order.UpdateOrder(ctx, resp.Order, resp.Status, resp.Accrual)
						if err != nil {
							w.logger.Error("unable to update order", zap.Error(err))
							continue
						}
					}
				}
			}

			w.logger.Info("worker finished")
		}
	}
}

func (w *Worker) requestAccrual(number string) (*resty.Response, error) {
	req := w.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	res := new(resty.Response)
	var innerErr error
	err := withRetry(func() error {
		res, innerErr = req.Get(fmt.Sprintf("http://%s/api/orders/%s", w.accrualSystemAddr, number))
		if innerErr != nil {
			return innerErr
		}
		return nil
	}, "failed to send metric")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func withRetry(fn func() error, warn string) error {
	interval := time.Second
	return retry.Do(fn,
		retry.Attempts(3),
		retry.Delay(interval),
		retry.OnRetry(func(n uint, err error) {
			log.Println(warn, zap.Uint("attempt", n), zap.Error(err))
			interval += 2 * time.Second
		}))
}
