export enum WalletType {
  Trading = 'trading',
  Profit = 'profit'
}

export interface WalletInfo {
  address: string;
  balance: number;
  type: WalletType;
}

export interface WalletTransfer {
  fromType: WalletType;
  toType: WalletType;
  amount: number;
}

export interface WalletState {
  tradingWallet: WalletInfo | null;
  profitWallet: WalletInfo | null;
  isConnecting: boolean;
  error: string | null;
}
