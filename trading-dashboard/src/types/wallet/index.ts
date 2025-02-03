export interface WalletInfo {
  address: string;
  balance: number;
  type: 'trading' | 'profit';
}

export interface WalletTransfer {
  fromType: WalletInfo['type'];
  toType: WalletInfo['type'];
  amount: number;
}

export interface WalletState {
  tradingWallet: WalletInfo | null;
  profitWallet: WalletInfo | null;
  isConnecting: boolean;
  error: string | null;
}
