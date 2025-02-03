import '@testing-library/jest-dom'
import { cleanup } from '@testing-library/react'
import React from 'react'
import { configure } from '@testing-library/react'

configure({
  asyncUtilTimeout: 30000,
  testIdAttribute: 'data-testid'
})

interface MockWebSocketInstance extends WebSocket {
  close: jest.Mock;
  send: jest.Mock;
  addEventListener: jest.Mock;
  removeEventListener: jest.Mock;
  binaryType: BinaryType;
}

class MockWebSocket implements MockWebSocketInstance {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  url: string;
  readyState: number;
  onopen: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  binaryType: BinaryType = 'blob';
  bufferedAmount: number = 0;
  extensions: string = '';
  protocol: string = '';

  constructor(url: string, protocols?: string | string[]) {
    this.url = url;
    this.readyState = MockWebSocket.CONNECTING;
    if (protocols) {
      this.protocol = Array.isArray(protocols) ? protocols[0] : protocols;
    }
    jest.advanceTimersByTime(0);
    this.readyState = MockWebSocket.OPEN;
    if (this.onopen) {
      this.onopen(new Event('open'));
    }
  }

  close = jest.fn((code?: number, reason?: string) => {
    if (this.readyState === MockWebSocket.CLOSED) return;
    this.readyState = MockWebSocket.CLOSING;
    queueMicrotask(() => {
      this.readyState = MockWebSocket.CLOSED;
      this.onclose?.({ 
        type: 'close',
        target: this,
        code: code || 1000,
        reason: reason || '',
        wasClean: true
      });
    });
  });

  send = jest.fn(() => {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }
  });

  addEventListener = jest.fn((type: string, handler: (event: any) => void) => {
    switch (type) {
      case 'open':
        this.onopen = handler;
        if (this.readyState === MockWebSocket.OPEN) {
          queueMicrotask(() => handler({ type: 'open', target: this }));
        }
        break;
      case 'close':
        this.onclose = handler;
        break;
      case 'message':
        this.onmessage = handler;
        break;
      case 'error':
        this.onerror = handler;
        break;
    }
  });

  removeEventListener = jest.fn((type: string) => {
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
