module.exports = {
  testEnvironment: 'node',
  collectCoverageFrom: [
    'charge.js',
    'server.js',
    '!node_modules/**',
    '!index.js'
  ],
  coverageThreshold: {
    global: {
      branches: 75,
      functions: 80,
      lines: 80,
      statements: 80
    }
  },
  testMatch: [
    '**/__tests__/**/*.test.js',
    '**/?(*.)+(spec|test).js'
  ]
};
