import '@testing-library/jest-dom'
import { configure } from '@testing-library/react'
import { cleanup } from '@testing-library/react'
import React from 'react'

declare global {
  interface Window {
    env: {
      VITE_API_URL: string;
      VITE_WS_URL: string;
    };
  }
  namespace NodeJS {
    interface Global {
      importMeta: {
        env: {
          VITE_API_URL: string;
          MODE: string;
        }
      }
    }
  }
}

configure({
  testIdAttribute: 'data-testid',
  asyncUtilTimeout: 30000
})

jest.mock('./hooks/websocket/useWebSocket')
jest.mock('./hooks/wallet/useWallet')
jest.mock('./hooks/ai/useAIAnalysis')
jest.mock('./hooks/auth/useAuth')

process.env.VITE_API_URL = 'http://localhost:8080'
process.env.MODE = 'test'

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  configurable: true,
  value: jest.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})

Object.defineProperty(window, 'env', {
  writable: true,
  configurable: true,
  value: {
    VITE_API_URL: 'http://localhost:8080',
    VITE_WS_URL: 'ws://localhost:8080',
  },
})

window.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

window.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Setup Vite environment
(global as any).import = {
  meta: {
    env: {
      VITE_API_URL: 'http://localhost:8080',
      MODE: 'test'
    }
  }
};

// Mock Recharts components
jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => 
    React.createElement('div', { 'data-testid': 'responsive-container' }, children),
  LineChart: ({ children }: { children: React.ReactNode }) => 
    React.createElement('div', { 'data-testid': 'line-chart' }, children),
  Line: () => React.createElement('div', { 'data-testid': 'line' }),
  XAxis: () => React.createElement('div', { 'data-testid': 'x-axis' }),
  YAxis: () => React.createElement('div', { 'data-testid': 'y-axis' }),
  CartesianGrid: () => React.createElement('div', { 'data-testid': 'cartesian-grid' }),
  Tooltip: () => React.createElement('div', { 'data-testid': 'tooltip' }),
  Legend: () => React.createElement('div', { 'data-testid': 'legend' })
}));

configure({
  asyncUtilTimeout: 30000,
  testIdAttribute: 'data-testid'
})

class MockWebSocket implements WebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  readonly CONNECTING = MockWebSocket.CONNECTING;
  readonly OPEN = MockWebSocket.OPEN;
  readonly CLOSING = MockWebSocket.CLOSING;
  readonly CLOSED = MockWebSocket.CLOSED;

  url: string;
  readyState: number;
  bufferedAmount = 0;
  extensions = '';
  protocol = '';
  binaryType: BinaryType = 'blob';

  onopen: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;

  constructor(url: string, protocols?: string | string[]) {
    this.url = url;
    this.readyState = MockWebSocket.CONNECTING;
    if (protocols) {
      this.protocol = Array.isArray(protocols) ? protocols[0] : protocols;
    }
    
    queueMicrotask(() => {
      if (this.readyState !== MockWebSocket.CLOSED) {
        this.readyState = MockWebSocket.OPEN;
        const event = new Event('open');
        this.onopen?.(event);
        this.dispatchEvent(event);
      }
    });
  }

  close = jest.fn((code?: number, reason?: string): void => {
    if (this.readyState === MockWebSocket.CLOSED) return;
    this.readyState = MockWebSocket.CLOSING;
    queueMicrotask(() => {
      this.readyState = MockWebSocket.CLOSED;
      const closeEvent = new CloseEvent('close', {
        wasClean: true,
        code: code || 1000,
        reason: reason || ''
      });
      this.onclose?.(closeEvent);
      this.dispatchEvent(closeEvent);
    });
  });

  send = jest.fn((_data: string | ArrayBufferLike | Blob | ArrayBufferView): void => {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
  });

  addEventListener = jest.fn((type: string, listener: EventListenerOrEventListenerObject, _options?: boolean | AddEventListenerOptions): void => {
    const handler = typeof listener === 'function' ? listener : listener.handleEvent;
    switch (type) {
      case 'open':
        this.onopen = handler as (event: Event) => void;
        if (this.readyState === MockWebSocket.OPEN) {
          const event = new Event('open');
          Object.defineProperty(event, 'target', { value: this });
          queueMicrotask(() => handler(event));
        }
        break;
      case 'close':
        this.onclose = handler as (event: CloseEvent) => void;
        break;
      case 'message':
        this.onmessage = handler as (event: MessageEvent) => void;
        break;
      case 'error':
        this.onerror = handler as (event: Event) => void;
        break;
    }
  });

  removeEventListener = jest.fn((type: string, _listener: EventListenerOrEventListenerObject, _options?: boolean | EventListenerOptions): void => {
    switch (type) {
      case 'open':
        this.onopen = null;
        break;
      case 'close':
        this.onclose = null;
        break;
      case 'message':
        this.onmessage = null;
        break;
      case 'error':
        this.onerror = null;
        break;
    }
  });

  dispatchEvent = jest.fn((event: Event): boolean => {
    switch (event.type) {
      case 'open':
        this.onopen?.(event);
        break;
      case 'close':
        this.onclose?.(event as CloseEvent);
        break;
      case 'message':
        this.onmessage?.(event as MessageEvent);
        break;
      case 'error':
        this.onerror?.(event);
        break;
    }
    return true;
  });
}

// Setup global mocks
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
  root: null,
  rootMargin: '',
  thresholds: [1]
}));

global.WebSocket = MockWebSocket as any;

global.matchMedia = jest.fn().mockImplementation(query => ({
  matches: false,
  media: query,
  onchange: null,
  addListener: jest.fn(),
  removeListener: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  dispatchEvent: jest.fn(),
}));

// Reset mocks and cleanup after each test
beforeAll(() => {
  jest.useFakeTimers();
});

beforeEach(() => {
  jest.clearAllMocks();
  jest.clearAllTimers();
  jest.resetModules();
  cleanup();
});

afterEach(() => {
  cleanup();
  jest.clearAllTimers();
  jest.clearAllMocks();
  jest.resetModules();
});

afterAll(() => {
  jest.useRealTimers();
  jest.restoreAllMocks();
});

// Mock Recharts components
jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => 
    React.createElement('div', { 'data-testid': 'responsive-container' }, children),
  LineChart: ({ children }: { children: React.ReactNode }) => 
    React.createElement('div', { 'data-testid': 'line-chart' }, children),
  Line: () => React.createElement('div', { 'data-testid': 'line' }),
  XAxis: () => React.createElement('div', { 'data-testid': 'x-axis' }),
  YAxis: () => React.createElement('div', { 'data-testid': 'y-axis' }),
  CartesianGrid: () => React.createElement('div', { 'data-testid': 'cartesian-grid' }),
  Tooltip: () => React.createElement('div', { 'data-testid': 'tooltip' }),
  Legend: () => React.createElement('div', { 'data-testid': 'legend' })
}));
