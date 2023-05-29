package models

type Release struct {
	TagName string `json:"tag_name"`
}
type Message struct {
	Type        string `json:"@type"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"amount"`
}

type Body struct {
	Messages                    []Message     `json:"messages"`
	Memo                        string        `json:"memo"`
	TimeoutHeight               string        `json:"timeout_height"`
	ExtensionOptions            []interface{} `json:"extension_options"`
	NonCriticalExtensionOptions []interface{} `json:"non_critical_extension_options"`
}

type Fee struct {
	Amount []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"amount"`
	GasLimit string `json:"gas_limit"`
	Payer    string `json:"payer"`
	Granter  string `json:"granter"`
}

type AuthInfo struct {
	SignerInfos []interface{} `json:"signer_infos"`
	Fee         Fee           `json:"fee"`
}

type Tx struct {
	Body       Body          `json:"body"`
	AuthInfo   AuthInfo      `json:"auth_info"`
	Signatures []interface{} `json:"signatures"`
}
