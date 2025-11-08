/**
 * Copyright 2018 Google LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Set environment variables before requiring the module
process.env.PORT = '7001';
process.env.DISABLE_PROFILER = 'true';
process.env.ENABLE_TRACING = '0';

const {
  getSupportedCurrencies,
  convert,
  check,
  _getCurrencyData,
  _carry
} = require('../server.js');

describe('CurrencyService Unit Tests', () => {
  describe('_getCurrencyData', () => {
    it('should load currency data from JSON file', (done) => {
      _getCurrencyData((data) => {
        expect(data).toBeDefined();
        expect(typeof data).toBe('object');
        expect(Object.keys(data).length).toBeGreaterThan(0);
        done();
      });
    });

    it('should include EUR as base currency', (done) => {
      _getCurrencyData((data) => {
        expect(data.EUR).toBeDefined();
        expect(data.EUR).toBe('1.0');
        done();
      });
    });

    it('should include major currencies', (done) => {
      _getCurrencyData((data) => {
        expect(data.USD).toBeDefined();
        expect(data.JPY).toBeDefined();
        expect(data.GBP).toBeDefined();
        expect(data.CNY).toBeDefined();
        done();
      });
    });
  });

  describe('_carry', () => {
    it('should handle simple amounts without carry', () => {
      const result = _carry({ units: 10, nanos: 500000000 });
      expect(result.units).toBe(10);
      expect(result.nanos).toBe(500000000);
    });

    it('should carry nanos overflow to units', () => {
      const result = _carry({ units: 5, nanos: 1500000000 });
      expect(result.units).toBe(6);
      expect(result.nanos).toBe(500000000);
    });

    it('should handle fractional units', () => {
      const result = _carry({ units: 10.5, nanos: 0 });
      expect(result.units).toBe(10);
      expect(result.nanos).toBe(500000000);
    });

    it('should handle large nanos values', () => {
      const result = _carry({ units: 0, nanos: 3000000000 });
      expect(result.units).toBe(3);
      expect(result.nanos).toBe(0);
    });

    it('should handle zero values', () => {
      const result = _carry({ units: 0, nanos: 0 });
      expect(result.units).toBe(0);
      expect(result.nanos).toBe(0);
    });

    it('should handle fractional units with existing nanos', () => {
      const result = _carry({ units: 10.75, nanos: 250000000 });
      expect(result.units).toBe(11);
      expect(result.nanos).toBe(0);
    });
  });

  describe('getSupportedCurrencies', () => {
    it('should return list of currency codes', (done) => {
      const mockCall = {};
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response).toBeDefined();
        expect(response.currency_codes).toBeDefined();
        expect(Array.isArray(response.currency_codes)).toBe(true);
        done();
      };

      getSupportedCurrencies(mockCall, callback);
    });

    it('should include EUR, USD, JPY', (done) => {
      const mockCall = {};
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response.currency_codes).toContain('EUR');
        expect(response.currency_codes).toContain('USD');
        expect(response.currency_codes).toContain('JPY');
        done();
      };

      getSupportedCurrencies(mockCall, callback);
    });

    it('should return at least 30 currencies', (done) => {
      const mockCall = {};
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response.currency_codes.length).toBeGreaterThanOrEqual(30);
        done();
      };

      getSupportedCurrencies(mockCall, callback);
    });
  });

  describe('convert', () => {
    it('should convert USD to EUR', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'USD',
            units: 100,
            nanos: 0
          },
          to_code: 'EUR'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('EUR');
        expect(result.units).toBeGreaterThan(80);
        expect(result.units).toBeLessThan(95);
        expect(result.nanos).toBeLessThan(1000000000);
        done();
      };

      convert(mockCall, callback);
    });

    it('should convert EUR to USD', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'EUR',
            units: 100,
            nanos: 0
          },
          to_code: 'USD'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('USD');
        expect(result.units).toBeGreaterThan(110);
        expect(result.units).toBeLessThan(115);
        done();
      };

      convert(mockCall, callback);
    });

    it('should handle JPY to USD conversion', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'JPY',
            units: 10000,
            nanos: 0
          },
          to_code: 'USD'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('USD');
        expect(result.units).toBeGreaterThan(80);
        expect(result.units).toBeLessThan(95);
        done();
      };

      convert(mockCall, callback);
    });

    it('should handle fractional amounts with nanos', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'USD',
            units: 10,
            nanos: 500000000
          },
          to_code: 'EUR'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('EUR');
        expect(result.units).toBeGreaterThan(8);
        expect(result.units).toBeLessThan(11);
        done();
      };

      convert(mockCall, callback);
    });

    it('should handle zero amount conversion', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'USD',
            units: 0,
            nanos: 0
          },
          to_code: 'EUR'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('EUR');
        expect(result.units).toBe(0);
        expect(result.nanos).toBe(0);
        done();
      };

      convert(mockCall, callback);
    });

    it('should handle small fractional amounts', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'USD',
            units: 0,
            nanos: 100000000
          },
          to_code: 'EUR'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('EUR');
        done();
      };

      convert(mockCall, callback);
    });

    it('should convert between non-EUR currencies (CAD to GBP)', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'CAD',
            units: 100,
            nanos: 0
          },
          to_code: 'GBP'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('GBP');
        expect(result.units).toBeGreaterThan(50);
        expect(result.units).toBeLessThan(60);
        done();
      };

      convert(mockCall, callback);
    });

    it('should handle large amounts', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'USD',
            units: 1000000,
            nanos: 0
          },
          to_code: 'JPY'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('JPY');
        expect(result.units).toBeGreaterThan(100000000);
        done();
      };

      convert(mockCall, callback);
    });

    it('should convert CNY to INR', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'CNY',
            units: 1000,
            nanos: 0
          },
          to_code: 'INR'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('INR');
        expect(result.units).toBeGreaterThan(9000);
        expect(result.units).toBeLessThan(11000);
        done();
      };

      convert(mockCall, callback);
    });

    it('should handle amounts near overflow boundary', (done) => {
      const mockCall = {
        request: {
          from: {
            currency_code: 'USD',
            units: 99,
            nanos: 999999999
          },
          to_code: 'EUR'
        }
      };

      const callback = (err, result) => {
        expect(err).toBeNull();
        expect(result).toBeDefined();
        expect(result.currency_code).toBe('EUR');
        expect(result.nanos).toBeLessThan(1000000000);
        done();
      };

      convert(mockCall, callback);
    });
  });

  describe('check (Health Check)', () => {
    it('should return SERVING status', (done) => {
      const mockCall = {};
      const callback = (err, response) => {
        expect(err).toBeNull();
        expect(response).toBeDefined();
        expect(response.status).toBe('SERVING');
        done();
      };

      check(mockCall, callback);
    });
  });
});
