package sp

const (
	HTTPHeaderAuthorization = "Authorization"
	SignAlgorithm           = "ECDSA-secp256k1"
	AuthV1                  = "authTypeV1"
	AuthV2                  = "authTypeV2"
)

// AuthInfo is the authorization info of requests
type AuthInfo struct {
	SignType      string // if using wallet sign, set authV2
	WalletSignStr string
}

// NewAuthInfo returns the AuthInfo which need to pass to api
// useWalletSign indicates whether you need use wallet to sign
// signStr indicates the wallet signature or jwt token
func NewAuthInfo(useWalletSign bool, signStr string) AuthInfo {
	if !useWalletSign {
		return AuthInfo{
			SignType:      AuthV1,
			WalletSignStr: "",
		}
	} else {
		return AuthInfo{
			SignType:      AuthV2,
			WalletSignStr: signStr,
		}
	}
}
