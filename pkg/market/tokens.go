package market

// TokenMap maps token symbols to their Solana addresses
var TokenMap = map[string]string{
	"SOL":  "So11111111111111111111111111111111111111112",
	"WSOL": "So11111111111111111111111111111111111111112",
	"USDC": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
	"USDT": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
}

// GetTokenAddress returns the Solana address for a given token symbol
func GetTokenAddress(symbol string) string {
	if address, ok := TokenMap[symbol]; ok {
		return address
	}
	return symbol
}

// GetTokenSymbol returns the symbol for a given token address
func GetTokenSymbol(address string) string {
	for symbol, addr := range TokenMap {
		if addr == address {
			return symbol
		}
	}
	return address
}
