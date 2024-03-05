package types

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

type Status string

const (
	New        Status = "NEW"
	Registered Status = "REGISTERED"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)

type Order struct {
	Number     string    `db:"number"      json:"number"`
	UserID     string    `db:"user_id"     json:"user_id,omitempty"`
	Status     Status    `db:"status"      json:"status"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
	Accrual    float64   `db:"accrual"     json:"accrual,omitempty"`
}

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		*Alias
		UploadedAt string `json:"uploaded_at"`
	}{
		Alias:      (*Alias)(o),
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	})
}

func ValidateOrder(order string) error {
	num, err := strconv.ParseInt(order, 10, 64)
	if err != nil {
		return err
	}

	if !validLuhn(num) {
		return errors.New("incorrect order number")
	}
	return nil
}

func validLuhn(number int64) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

type Withdrawal struct {
	UserID      string    `db:"user_id" json:"user_id,omitempty"`
	Number      string    `db:"number" json:"number"`
	ProcessedAt time.Time `db:"processed_at" json:"processed_at"`
	Sum         float64   `db:"sum" json:"sum"`
}

func (w *Withdrawal) MarshalJSON() ([]byte, error) {
	type Alias Withdrawal
	return json.Marshal(&struct {
		*Alias
		ProcessedAt string `json:"processed_at"`
	}{
		Alias:       (*Alias)(w),
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	})
}

type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
