package types

type SendTokenRequest struct {
	Token     string
	Amount    int64
	ToAddress string
}

// TxBroadcastResponse Generic tx response
type TxBroadcastResponse struct {
	Ok     bool   `json:"ok"`
	Log    string `json:"log"`
	TxHash string `json:"txHash"`
	Code   uint32 `json:"code"`
	Data   string `json:"data"`
}
