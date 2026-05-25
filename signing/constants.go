package signing

// L1 (EIP-712) auth domain constants. These never change across networks; only
// the chainId in the domain separator does.
const (
	ClobDomainName = "ClobAuthDomain"
	ClobVersion    = "1"
	MsgToSign      = "This message attests that I control the given wallet"
)
