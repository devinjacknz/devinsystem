export interface AuthCredentials {
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: string;
    username: string;
  };
}

export interface AuthState {
  isAuthenticated: boolean;
  user: AuthResponse['user'] | null;
  token: string | null;
  error: string | null;
  isLoading: boolean;
}

export interface AuthContextType extends AuthState {
  login: (credentials: AuthCredentials) => Promise<void>;
  logout: () => void;
  register: (credentials: AuthCredentials) => Promise<void>;
}
