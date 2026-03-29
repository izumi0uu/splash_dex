package constants

const SolChainId = "100000"
const SolChainIdInt = 100000
const TrxChainIdInt = 110000
const EthChainIdInt = 1
const BaseChainIdInt = 8453
const BscChainIdInt = 56
const AtaAccountSize = 165
const MinSolVolume = uint64(1e7)
const AutoSlippage = 2500

const (
	ChainIconEth  = "https://cdn.pumpx.ai/static/img/chain/eth.png"
	ChainIconBsc  = "https://cdn.pumpx.ai/static/img/chain/bsc.png"
	ChainIconBase = "https://cdn.pumpx.ai/static/img/chain/base.png"
	ChainIconSol  = "https://cdn.pumpx.ai/static/img/chain/sol.png"
	ChainIconTron = "https://cdn.pumpx.ai/static/img/chain/trx.png"
)

const SolDecimal = 9

// --- Raydium DEX Programs ---
const ProgramStrRaydiumV4AMM = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8" // Raydium V4 AMM: constant product x*y=k, legacy pools & PumpFun graduates
const ProgramStrRaydiumV4AMMDevnet = "DRaya7Kj3aMWQSy19kSjvmuwq9docCHofyP9kanQGaav"
const ProgramStrRaydiumV4CLMM = "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK" // Raydium CLMM: concentrated liquidity (similar to Uniswap V3)
const ProgramStrRaydiumCPMM = "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C"   // Raydium CPMM: new constant product pools, many new tokens launch here
const ProgramStrRaydiumV2 = "RVKd61ztZW9GUwhRbbLoYVRE5Xf1B2tVscKqwZqXgEr"      // Raydium V2: legacy AMM, very low volume

// --- Aggregator ---
const ProgramStrJupiter = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4" // Jupiter: largest swap aggregator on Solana, routes through underlying DEXes

// --- Other DEX Programs ---
const ProgramStrOrca = "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"         // Orca Whirlpool: concentrated liquidity DEX
const ProgramStrMeteoraDLMM = "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"  // Meteora DLMM: dynamic liquidity market maker
const ProgramStrMeteoraPool = "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB" // Meteora standard AMM pools
const ProgramStrPhoenix = "PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY"      // Phoenix: on-chain order book DEX
const ProgramStrLifinity = "2wT8Yq49kHgDzXuPxZSaeLaH1qbmGXtEyPy64bL7aD3c"    // Lifinity: proactive market maker DEX

// --- Token Launch Platforms ---
const ProgramStrPumpFun = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"  // PumpFun: bonding curve token launchpad
const ProgramStrPumpAmm = "pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"  // PumpSwap: PumpFun's own AMM after token graduation
const ProgramStrMoonshot = "MoonCVVNZFSYkqNXP6bxHLPl6A4UA2iJhgoRDBsydJY" // Moonshot: alternative token launchpad

// --- Solana System Programs ---
const ProgramStrVote = "Vote111111111111111111111111111111111111111"             // Vote program: ~80% of all txs are vote txs
const ProgramStrToken = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"            // SPL Token program
const ProgramStrToken2022 = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"        // Token-2022: extended token standard with transfer fees, etc.
const ProgramStrSystem = "11111111111111111111111111111111"                      // System program
const ProgramStrAssociatedToken = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL" // Associated Token Account program

// --- Token Addresses (Solana) ---
const TokenStrWrapSol = "So11111111111111111111111111111111111111112" // Wrapped SOL
const TokenStrUSDC = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"   // USDC (Circle)
const TokenStrUSDT = "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"   // USDT (Tether)
const TokenStrPYUSD = "2b1kV6DkPAnxd5ixfnxCpjxmKwqjjaYmCZfHsFu24GXo"  // PYUSD (PayPal USD)

const TokenStrWrapTrx = "TNUC9Qb1rRpS5CbWLmNMxXBjyFoydXjWFR"

const SunPumpLogo = "/static/img/brand/sunpump_logo.jpg"
const PumpFunLogo = "/static/img/brand/pump_fun.png"

const TokenStrWTRX = "TNUC9Qb1rRpS5CbWLmNMxXBjyFoydXjWFR"

const (
	TrxContractV2Router  = "TKzxdSv2FZKQrEqkKVgp5DcwEXBEKMg2Ax"
	TrxContractV1Factroy = "TXk8rQSAvPvBBNtqSoY6nCfsXWCSSpTVQF" // JustswapFactory
	TrxContractPsmUsdd   = "TPYmHEhy5n8TCEfYGqW2rPxsghSfzghPDn"
	TrxContractV3Router  = "TQAvWQpT9H916GckwWDJNhYZvQMkuRL7PN"
	TrxContractWtrx      = "TNUC9Qb1rRpS5CbWLmNMxXBjyFoydXjWFR"
)

const WBsc = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"

const (
	BaseTokenAddressSol  = "solana"
	BaseTokenAddressEth  = "eth"
	BaseTokenAddressBase = "base_eth"
	BaseTokenAddressBsc  = "bnb"
)
const (
	BaseTokenSymbolSol  = "SOL"
	BaseTokenSymbolEth  = "ETH"
	BaseTokenSymbolBase = "BASE_ETH"
	BaseTokenSymbolBsc  = "BNB"
)

const (
	MinWithdrawAmountSol = "0.005"
)

const (
	ChainNameEth  = "eth"
	ChainNameBase = "base"
	ChainNameBsc  = "bsc"
)

const EthDecimal = 18
const BaseDecimal = 18
const BscDecimal = 18

const AllChainIdInt = -1

const EvmChains = "1,8453,56"

const (
	SolQuickBuyDefault  = "0.1,0.2,0.5,1,2,5"
	EthQuickBuyDefault  = "0.01,0.02,0.05,0.1,0.5,1"
	BaseQuickBuyDefault = "0.01,0.02,0.05,0.1,0.5,1"
	BscQuickBuyDefault  = "0.03,0.06,0.15,0.3,1,2"
	QuickSellDefault    = "25,33,50,60,75,100"
)
