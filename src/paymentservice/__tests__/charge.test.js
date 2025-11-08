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

const charge = require('../charge');

describe('PaymentService - Charge', () => {
  const validAmount = {
    currency_code: 'USD',
    units: 100,
    nanos: 0
  };

  describe('Valid Credit Card Processing', () => {
    test('should process valid VISA card', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111', // Valid VISA test number
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);

      expect(result).toHaveProperty('transaction_id');
      expect(result.transaction_id).toBeTruthy();
      expect(typeof result.transaction_id).toBe('string');
    });

    test('should process valid MasterCard', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '5555555555554444', // Valid MasterCard test number
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);

      expect(result).toHaveProperty('transaction_id');
      expect(result.transaction_id).toBeTruthy();
    });

    test('should generate unique transaction IDs', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result1 = charge(request);
      const result2 = charge(request);

      expect(result1.transaction_id).not.toBe(result2.transaction_id);
    });
  });

  describe('Invalid Credit Card Handling', () => {
    test('should reject invalid card number', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '1234567890123456', // Invalid number
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      expect(() => charge(request)).toThrow('Credit card info is invalid');
    });

    test('should reject AMEX card', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '378282246310005', // Valid AMEX test number
          credit_card_cvv: 1234,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      expect(() => charge(request)).toThrow('cannot process');
    });

    test('should reject expired card', () => {
      const currentYear = new Date().getFullYear();
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: currentYear - 1,
          credit_card_expiration_month: 12
        }
      };

      expect(() => charge(request)).toThrow('expired on');
    });

    test('should reject card expiring this month if current year', () => {
      const currentYear = new Date().getFullYear();
      const currentMonth = new Date().getMonth() + 1;

      // Card that expires last month
      const lastMonth = currentMonth === 1 ? 12 : currentMonth - 1;
      const yearForLastMonth = currentMonth === 1 ? currentYear - 1 : currentYear;

      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: yearForLastMonth,
          credit_card_expiration_month: lastMonth
        }
      };

      expect(() => charge(request)).toThrow('expired on');
    });
  });

  describe('Edge Cases', () => {
    test('should handle large transaction amounts', () => {
      const request = {
        amount: {
          currency_code: 'USD',
          units: 999999,
          nanos: 999999999
        },
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);
      expect(result).toHaveProperty('transaction_id');
    });

    test('should handle zero amount transactions', () => {
      const request = {
        amount: {
          currency_code: 'USD',
          units: 0,
          nanos: 0
        },
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);
      expect(result).toHaveProperty('transaction_id');
    });

    test('should handle card expiring in distant future', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '5555555555554444',
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 10,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);
      expect(result).toHaveProperty('transaction_id');
    });
  });

  describe('Different Currency Codes', () => {
    test('should process EUR currency', () => {
      const request = {
        amount: {
          currency_code: 'EUR',
          units: 50,
          nanos: 0
        },
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);
      expect(result).toHaveProperty('transaction_id');
    });

    test('should process JPY currency', () => {
      const request = {
        amount: {
          currency_code: 'JPY',
          units: 10000,
          nanos: 0
        },
        credit_card: {
          credit_card_number: '5555555555554444',
          credit_card_cvv: 123,
          credit_card_expiration_year: new Date().getFullYear() + 2,
          credit_card_expiration_month: 12
        }
      };

      const result = charge(request);
      expect(result).toHaveProperty('transaction_id');
    });
  });
});
