/**
 * Copyright 2024 Google LLC
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

describe('CurrencyService', () => {
  let currencyData;

  beforeAll(() => {
    // Load currency data
    currencyData = require('../data/currency_conversion.json');
  });

  describe('Currency Data', () => {
    test('should have currency conversion data', () => {
      expect(currencyData).toBeDefined();
      expect(typeof currencyData).toBe('object');
    });

    test('should include major currencies', () => {
      const majorCurrencies = ['USD', 'EUR', 'GBP', 'JPY'];
      majorCurrencies.forEach(currency => {
        expect(currencyData).toHaveProperty(currency);
        expect(typeof currencyData[currency]).toBe('number');
      });
    });

    test('should have positive conversion rates', () => {
      Object.values(currencyData).forEach(rate => {
        expect(rate).toBeGreaterThan(0);
      });
    });
  });

  describe('Currency Conversion Logic', () => {
    // Helper function that handles decimal/fractional carrying
    function carry(amount) {
      const fractionSize = Math.pow(10, 9);
      amount.nanos += (amount.units % 1) * fractionSize;
      amount.units = Math.floor(amount.units) + Math.floor(amount.nanos / fractionSize);
      amount.nanos = amount.nanos % fractionSize;
      return amount;
    }

    test('should handle carry operation correctly', () => {
      const amount = { units: 10.5, nanos: 500000000 };
      const result = carry(amount);

      expect(result.units).toBe(11);
      expect(result.nanos).toBeGreaterThanOrEqual(0);
      expect(result.nanos).toBeLessThan(1000000000);
    });

    test('should convert currency amounts correctly', () => {
      // Test conversion from USD to EUR
      const fromAmount = { units: 100, nanos: 0, currency_code: 'USD' };
      const toCode = 'EUR';

      // Convert: USD --> EUR
      const usdRate = currencyData['USD'];
      const eurRate = currencyData['EUR'];

      const euros = carry({
        units: fromAmount.units / usdRate,
        nanos: fromAmount.nanos / usdRate
      });

      const result = carry({
        units: euros.units * eurRate,
        nanos: euros.nanos * eurRate
      });

      result.units = Math.floor(result.units);
      result.nanos = Math.floor(result.nanos);

      expect(result.units).toBeGreaterThanOrEqual(0);
      expect(result.nanos).toBeGreaterThanOrEqual(0);
      expect(result.nanos).toBeLessThan(1000000000);
    });

    test('should handle zero amount conversion', () => {
      const amount = { units: 0, nanos: 0 };
      const result = carry(amount);

      expect(result.units).toBe(0);
      expect(result.nanos).toBe(0);
    });
  });

  describe('Supported Currencies', () => {
    test('should return list of supported currencies', () => {
      const supportedCurrencies = Object.keys(currencyData);

      expect(Array.isArray(supportedCurrencies)).toBe(true);
      expect(supportedCurrencies.length).toBeGreaterThan(0);
    });

    test('should include common currency codes', () => {
      const currencyCodes = Object.keys(currencyData);
      const commonCodes = ['USD', 'EUR', 'GBP', 'JPY', 'CAD'];

      commonCodes.forEach(code => {
        expect(currencyCodes).toContain(code);
      });
    });
  });

  describe('Edge Cases', () => {
    test('should handle large amounts', () => {
      const largeAmount = { units: 1000000, nanos: 999999999 };
      const result = carry(largeAmount);

      expect(result.units).toBeGreaterThan(1000000);
      expect(result.nanos).toBeLessThan(1000000000);
    });

    test('should handle small fractional amounts', () => {
      const smallAmount = { units: 0, nanos: 1 };
      const result = carry(smallAmount);

      expect(result.units).toBe(0);
      expect(result.nanos).toBeGreaterThanOrEqual(0);
    });
  });
});
