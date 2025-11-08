// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

const path = require('path');
const HipsterShopServer = require('../server');

// Mock the charge module
jest.mock('../charge', () => {
  return jest.fn((request) => {
    return { transaction_id: 'test-transaction-id-12345' };
  });
});

const charge = require('../charge');

describe('HipsterShopServer', () => {
  let server;
  const protoRoot = path.join(__dirname, '../proto');

  beforeEach(() => {
    jest.clearAllMocks();
    server = new HipsterShopServer(protoRoot, 8080);
  });

  describe('Constructor', () => {
    it('should create server instance without specifying port', () => {
      // When no port is provided, it uses HipsterShopServer.PORT (from process.env.PORT)
      const serverWithDefaultPort = new HipsterShopServer(protoRoot);
      expect(serverWithDefaultPort).toBeDefined();
      // Port might be undefined if process.env.PORT is not set, which is fine
    });

    it('should create server instance with custom port', () => {
      expect(server).toBeDefined();
      expect(server.port).toBe(8080);
    });

    it('should load proto packages', () => {
      expect(server.packages).toBeDefined();
      expect(server.packages.hipsterShop).toBeDefined();
      expect(server.packages.health).toBeDefined();
    });

    it('should create gRPC server', () => {
      expect(server.server).toBeDefined();
    });
  });

  describe('loadProto', () => {
    it('should load a proto file', () => {
      const protoPath = path.join(protoRoot, 'demo.proto');
      const loaded = server.loadProto(protoPath);
      expect(loaded).toBeDefined();
    });
  });

  describe('ChargeServiceHandler', () => {
    it('should call charge function with request', (done) => {
      const mockRequest = {
        amount: {
          currency_code: 'USD',
          units: 100,
          nanos: 0
        },
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: 2030,
          credit_card_expiration_month: 12
        }
      };

      const mockCall = { request: mockRequest };
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response).toBeDefined();
        expect(response.transaction_id).toBe('test-transaction-id-12345');
        expect(charge).toHaveBeenCalledWith(mockRequest);
        done();
      };

      HipsterShopServer.ChargeServiceHandler(mockCall, callback);
    });

    it('should handle charge function errors', (done) => {
      const mockError = new Error('Invalid credit card');
      charge.mockImplementationOnce(() => {
        throw mockError;
      });

      const mockCall = {
        request: {
          amount: { currency_code: 'USD', units: 100, nanos: 0 },
          credit_card: {
            credit_card_number: '1234567890123456',
            credit_card_cvv: 123,
            credit_card_expiration_year: 2030,
            credit_card_expiration_month: 12
          }
        }
      };

      const callback = (err, response) => {
        expect(err).toBeDefined();
        expect(err).toBe(mockError);
        expect(response).toBeUndefined();
        done();
      };

      HipsterShopServer.ChargeServiceHandler(mockCall, callback);
    });

    it('should handle different request formats', (done) => {
      const mockRequest = {
        amount: {
          currency_code: 'EUR',
          units: 50,
          nanos: 500000000
        },
        credit_card: {
          credit_card_number: '5555555555554444',
          credit_card_cvv: 456,
          credit_card_expiration_year: 2028,
          credit_card_expiration_month: 6
        }
      };

      const mockCall = { request: mockRequest };
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response).toBeDefined();
        expect(charge).toHaveBeenCalledWith(mockRequest);
        done();
      };

      HipsterShopServer.ChargeServiceHandler(mockCall, callback);
    });
  });

  describe('CheckHandler', () => {
    it('should return SERVING status', (done) => {
      const mockCall = {};
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response).toBeDefined();
        expect(response.status).toBe('SERVING');
        done();
      };

      HipsterShopServer.CheckHandler(mockCall, callback);
    });

    it('should always return SERVING status regardless of input', (done) => {
      const mockCall = { request: { service: 'PaymentService' } };
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response.status).toBe('SERVING');
        done();
      };

      HipsterShopServer.CheckHandler(mockCall, callback);
    });
  });

  describe('loadAllProtos', () => {
    it('should have PaymentService loaded after construction', () => {
      // Server already loaded in beforeEach via constructor
      expect(server.server).toBeDefined();
      expect(server.packages.hipsterShop).toBeDefined();
      expect(server.packages.health).toBeDefined();
      // Services are already registered in constructor, so we can't test registration again
      // without recreating the server
    });
  });
});
