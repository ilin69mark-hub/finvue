import { describe, test, expect } from 'vitest';

const API_BASE = 'http://localhost:8080';

describe('Backend API', () => {
  test('health endpoint should return OK', async () => {
    const response = await fetch(`${API_BASE}/health`);
    expect(response.status).toBe(200);
    const text = await response.text();
    expect(text).toBe('OK');
  });

  test('assets endpoint should return data', async () => {
    const response = await fetch(`${API_BASE}/api/v1/assets`);
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(Array.isArray(data)).toBe(true);
  });

  test('ohlcv endpoint should require asset_id', async () => {
    const response = await fetch(`${API_BASE}/api/v1/ohlcv`);
    expect(response.status).toBe(400);
  });

  test('invalid asset id should return 400', async () => {
    const response = await fetch(`${API_BASE}/api/v1/assets/invalid`);
    expect(response.status).toBe(400);
  });

  test('not found endpoint should return 404', async () => {
    const response = await fetch(`${API_BASE}/nonexistent`);
    expect(response.status).toBe(404);
  });
});