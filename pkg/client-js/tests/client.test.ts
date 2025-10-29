/**
 * Tests for Client class
 */

import { Client } from '../src/client';
import { Config } from '../src/types';
import axios from 'axios';

// Mock axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('Client', () => {
  let mockAxiosInstance: any;

  beforeEach(() => {
    jest.clearAllMocks();

    // Setup mock axios instance
    mockAxiosInstance = {
      get: jest.fn(),
      post: jest.fn(),
      put: jest.fn(),
      delete: jest.fn(),
      head: jest.fn(),
      defaults: {
        headers: {
          common: {} as Record<string, string>,
        },
      },
      interceptors: {
        response: {
          use: jest.fn(),
        },
      },
    };

    mockedAxios.create.mockReturnValue(mockAxiosInstance);
  });

  describe('Configuration', () => {
    it('should create client with valid config', () => {
      // Given: Valid configuration
      const config: Config = {
        baseURL: 'https://test.example.com',
      };

      // When: Creating a client
      const client = new Client(config);

      // Then: Client should be created successfully
      expect(client).toBeInstanceOf(Client);
    });

    it('should throw error with missing base URL', () => {
      // Given: Configuration without baseURL
      const config: Config = {
        baseURL: '',
      };

      // When & Then: Should throw error
      expect(() => new Client(config)).toThrow('baseURL is required');
    });

    it('should remove trailing slash from base URL', () => {
      // Given: Base URL with trailing slash
      const config: Config = {
        baseURL: 'https://test.example.com/',
      };

      // When: Creating client
      const client = new Client(config);

      // Then: Trailing slash should be removed
      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          baseURL: 'https://test.example.com',
        })
      );
    });

    it('should set authentication header when token provided', () => {
      // Given: Configuration with token
      const config: Config = {
        baseURL: 'https://test.example.com',
        token: 'test-token',
      };

      // When: Creating client
      const client = new Client(config);

      // Then: Authorization header should be set
      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer test-token',
          }),
        })
      );
    });

    it('should set custom timeout', () => {
      // Given: Configuration with custom timeout
      const config: Config = {
        baseURL: 'https://test.example.com',
        timeout: 120000,
      };

      // When: Creating client
      const client = new Client(config);

      // Then: Timeout should be set
      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          timeout: 120000,
        })
      );
    });

    it('should set custom user agent', () => {
      // Given: Configuration with custom user agent
      const config: Config = {
        baseURL: 'https://test.example.com',
        userAgent: 'my-app/1.0',
      };

      // When: Creating client
      const client = new Client(config);

      // Then: User-Agent header should be set
      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          headers: expect.objectContaining({
            'User-Agent': 'my-app/1.0',
          }),
        })
      );
    });

    it('should configure insecure mode', () => {
      // Given: Configuration with insecure skip verify
      const config: Config = {
        baseURL: 'https://test.example.com',
        insecureSkipVerify: true,
      };

      // When: Creating client
      const client = new Client(config);

      // Then: HTTPS agent should be configured
      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          httpsAgent: expect.anything(),
        })
      );
    });
  });

  describe('Token Management', () => {
    it('should update authentication token', () => {
      // Given: Client with initial token
      const config: Config = {
        baseURL: 'https://test.example.com',
        token: 'initial-token',
      };
      const client = new Client(config);

      // When: Updating token
      client.setToken('new-token');

      // Then: Authorization header should be updated
      expect(mockAxiosInstance.defaults.headers.common['Authorization']).toBe('Bearer new-token');
    });
  });
});
