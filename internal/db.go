package internal

type SavePaymentRequest struct {
	Backend     int    `json:"backend"`
	Amount      int64  `json:"amount"`
	RequestedAt string `json:"requested_at"`
}
