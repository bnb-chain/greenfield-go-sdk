package types

type TxResult struct {
	Hash string `json:"hash"`
	Log  string `json:"log"`
	Data string `json:"data"`
	Code int32  `json:"code"`
}

type SendTokenRequest struct {
	Token     string
	Amount    int64
	ToAddress string
}

type TxBroadcastResponse struct {
	Ok     bool   `json:"ok"`
	Log    string `json:"log"`
	TxHash string `json:"txHash"`
	Code   uint32 `json:"code"`
	Data   string `json:"data"`
}
