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

const charge = require('../charge');

describe('charge function', () => {
  const validAmount = {
    currency_code: 'USD',
    units: 100,
    nanos: 0
  };

  // Get future date for valid credit cards
  const futureYear = new Date().getFullYear() + 2;
  const futureMonth = 12;

  describe('Valid credit cards', () => {
    it('should process valid VISA card', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111', // Valid VISA test card
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
      expect(typeof result.transaction_id).toBe('string');
      expect(result.transaction_id.length).toBeGreaterThan(0);
    });

    it('should process valid MasterCard', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '5555555555554444', // Valid MasterCard test card
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
      expect(typeof result.transaction_id).toBe('string');
    });

    it('should generate unique transaction IDs', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result1 = charge(request);
      const result2 = charge(request);

      expect(result1.transaction_id).not.toBe(result2.transaction_id);
    });

    it('should process VISA card with different format', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4012888888881881', // Another valid VISA
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });

    it('should process MasterCard with different format', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '5105105105105100', // Another valid MasterCard
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });
  });

  describe('Invalid credit card numbers', () => {
    it('should reject invalid card number', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '1234567890123456', // Invalid card number
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      expect(() => charge(request)).toThrow('Credit card info is invalid');
    });

    it('should reject empty card number', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '',
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      expect(() => charge(request)).toThrow(); // Will throw error for invalid card
    });

    it('should reject card number with wrong checksum', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111112', // Wrong checksum
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      expect(() => charge(request)).toThrow('Credit card info is invalid');
    });
  });

  describe('Unaccepted card types', () => {
    it('should reject American Express card', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '378282246310005', // Valid AMEX
          credit_card_cvv: 1234,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      expect(() => charge(request)).toThrow(/cannot process.*credit cards/);
      expect(() => charge(request)).toThrow(/Only VISA or MasterCard is accepted/);
    });

    it('should reject Discover card', () => {
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '6011111111111117', // Valid Discover
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      expect(() => charge(request)).toThrow(/cannot process.*credit cards/);
    });
  });

  describe('Expired credit cards', () => {
    it('should reject card expired last year', () => {
      const lastYear = new Date().getFullYear() - 1;
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: lastYear,
          credit_card_expiration_month: 12
        }
      };

      expect(() => charge(request)).toThrow(/expired on/);
    });

    it('should reject card expired last month', () => {
      const currentDate = new Date();
      const lastMonth = currentDate.getMonth(); // getMonth() is 0-indexed, so this is actually last month
      const currentYear = currentDate.getFullYear();

      // If we're in January, use December of last year
      const expYear = lastMonth === 0 ? currentYear - 1 : currentYear;
      const expMonth = lastMonth === 0 ? 12 : lastMonth;

      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: expYear,
          credit_card_expiration_month: expMonth
        }
      };

      expect(() => charge(request)).toThrow(/expired on/);
    });

    it('should include card last 4 digits in expiry error message', () => {
      const lastYear = new Date().getFullYear() - 1;
      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: lastYear,
          credit_card_expiration_month: 12
        }
      };

      expect(() => charge(request)).toThrow(/ending 1111/);
    });

    it('should accept card expiring this month', () => {
      const currentDate = new Date();
      const currentMonth = currentDate.getMonth() + 1; // getMonth() is 0-indexed
      const currentYear = currentDate.getFullYear();

      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: currentYear,
          credit_card_expiration_month: currentMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });

    it('should accept card expiring next month', () => {
      const currentDate = new Date();
      let nextMonth = currentDate.getMonth() + 2; // +1 to convert to 1-indexed, +1 for next month
      let year = currentDate.getFullYear();

      if (nextMonth > 12) {
        nextMonth = 1;
        year += 1;
      }

      const request = {
        amount: validAmount,
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: year,
          credit_card_expiration_month: nextMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });
  });

  describe('Different amounts', () => {
    it('should process small amount', () => {
      const request = {
        amount: {
          currency_code: 'USD',
          units: 1,
          nanos: 0
        },
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });

    it('should process large amount', () => {
      const request = {
        amount: {
          currency_code: 'USD',
          units: 999999,
          nanos: 999999999
        },
        credit_card: {
          credit_card_number: '4111111111111111',
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });

    it('should process different currency', () => {
      const request = {
        amount: {
          currency_code: 'EUR',
          units: 50,
          nanos: 0
        },
        credit_card: {
          credit_card_number: '5555555555554444',
          credit_card_cvv: 123,
          credit_card_expiration_year: futureYear,
          credit_card_expiration_month: futureMonth
        }
      };

      const result = charge(request);
      expect(result).toBeDefined();
      expect(result.transaction_id).toBeDefined();
    });
  });
});
