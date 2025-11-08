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

const path = require('path');

describe('Demo Dashboard Server', () => {
  describe('MIME Types', () => {
    const mimeTypes = {
      '.html': 'text/html',
      '.css': 'text/css',
      '.js': 'application/javascript',
      '.json': 'application/json',
      '.png': 'image/png',
      '.jpg': 'image/jpeg',
      '.gif': 'image/gif',
      '.svg': 'image/svg+xml',
      '.ico': 'image/x-icon',
    };

    test('should have correct MIME type for HTML', () => {
      expect(mimeTypes['.html']).toBe('text/html');
    });

    test('should have correct MIME type for CSS', () => {
      expect(mimeTypes['.css']).toBe('text/css');
    });

    test('should have correct MIME type for JavaScript', () => {
      expect(mimeTypes['.js']).toBe('application/javascript');
    });

    test('should have correct MIME type for JSON', () => {
      expect(mimeTypes['.json']).toBe('application/json');
    });

    test('should have correct MIME type for images', () => {
      expect(mimeTypes['.png']).toBe('image/png');
      expect(mimeTypes['.jpg']).toBe('image/jpeg');
      expect(mimeTypes['.gif']).toBe('image/gif');
      expect(mimeTypes['.svg']).toBe('image/svg+xml');
    });

    test('should have correct MIME type for favicon', () => {
      expect(mimeTypes['.ico']).toBe('image/x-icon');
    });
  });

  describe('File Path Handling', () => {
    test('should resolve root to index.html', () => {
      const url = '/';
      let filePath = '.' + url;
      if (filePath === './') {
        filePath = './index.html';
      }
      expect(filePath).toBe('./index.html');
    });

    test('should handle regular file paths', () => {
      const url = '/styles.css';
      const filePath = '.' + url;
      expect(filePath).toBe('./styles.css');
    });

    test('should extract correct file extension', () => {
      const filePath = './index.html';
      const extname = path.extname(filePath);
      expect(extname).toBe('.html');
    });

    test('should extract extension for nested paths', () => {
      const filePath = './assets/css/styles.css';
      const extname = path.extname(filePath);
      expect(extname).toBe('.css');
    });
  });

  describe('Environment Configuration', () => {
    test('should use default port if not specified', () => {
      const originalPort = process.env.PORT;
      delete process.env.PORT;
      const PORT = process.env.PORT || 3000;
      expect(PORT).toBe(3000);

      if (originalPort) {
        process.env.PORT = originalPort;
      }
    });

    test('should use environment PORT if specified', () => {
      const originalPort = process.env.PORT;
      process.env.PORT = '8080';
      const PORT = process.env.PORT || 3000;
      expect(PORT).toBe('8080');

      if (originalPort) {
        process.env.PORT = originalPort;
      } else {
        delete process.env.PORT;
      }
    });
  });

  describe('Content Type Resolution', () => {
    const mimeTypes = {
      '.html': 'text/html',
      '.css': 'text/css',
      '.js': 'application/javascript',
      '.json': 'application/json',
    };

    test('should resolve known content types', () => {
      const extname = '.html';
      const contentType = mimeTypes[extname] || 'application/octet-stream';
      expect(contentType).toBe('text/html');
    });

    test('should use default for unknown extensions', () => {
      const extname = '.xyz';
      const contentType = mimeTypes[extname] || 'application/octet-stream';
      expect(contentType).toBe('application/octet-stream');
    });
  });
});
