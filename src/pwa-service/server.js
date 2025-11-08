/**
 * PWA Service - Static File Server
 *
 * Serves PWA files including service worker, manifest, and static assets
 */

const express = require('express');
const path = require('path');
const compression = require('compression');
const helmet = require('helmet');

const app = express();
const PORT = process.env.PORT || 8095;

// Security middleware
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "'unsafe-inline'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", 'data:', 'https:'],
      connectSrc: ["'self'", 'ws:', 'wss:'],
      fontSrc: ["'self'"],
      objectSrc: ["'none'"],
      mediaSrc: ["'self'"],
      frameSrc: ["'none'"],
    },
  },
  crossOriginEmbedderPolicy: false,
}));

// Compression middleware
app.use(compression());

// Logging middleware
app.use((req, res, next) => {
  console.log(`[${new Date().toISOString()}] ${req.method} ${req.url}`);
  next();
});

// Serve static files
app.use(express.static(path.join(__dirname, 'public'), {
  maxAge: '1d',
  etag: true,
}));

// Service Worker - no caching
app.get('/service-worker.js', (req, res) => {
  res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
  res.setHeader('Content-Type', 'application/javascript');
  res.sendFile(path.join(__dirname, 'service-worker.js'));
});

// Manifest - cache for 1 day
app.get('/manifest.json', (req, res) => {
  res.setHeader('Cache-Control', 'public, max-age=86400');
  res.setHeader('Content-Type', 'application/json');
  res.sendFile(path.join(__dirname, 'public', 'manifest.json'));
});

// Offline page
app.get('/offline.html', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'offline.html'));
});

// Mock API endpoints for demo/testing

// Cart API
app.use(express.json());

let cartData = {};

app.post('/api/cart', (req, res) => {
  const { userId, item } = req.body;

  if (!cartData[userId]) {
    cartData[userId] = [];
  }

  cartData[userId].push({
    ...item,
    addedAt: new Date().toISOString(),
  });

  console.log(`[Cart] Item added for user ${userId}:`, item);

  res.json({
    success: true,
    cart: cartData[userId],
  });
});

app.get('/api/cart/:userId', (req, res) => {
  const { userId } = req.params;
  res.json({
    cart: cartData[userId] || [],
  });
});

// Push subscription API
let subscriptions = [];

app.post('/api/push/subscribe', (req, res) => {
  const subscription = req.body;

  subscriptions.push(subscription);
  console.log('[Push] New subscription registered');

  res.json({
    success: true,
    message: 'Subscription registered successfully',
  });
});

// Health check
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: 'pwa-service',
    timestamp: new Date().toISOString(),
  });
});

// Catch-all route - serve index.html for SPA routing
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'), (err) => {
    if (err) {
      res.status(404).send('Not Found');
    }
  });
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error('[Error]', err);
  res.status(500).json({
    error: 'Internal Server Error',
    message: err.message,
  });
});

// Start server
app.listen(PORT, () => {
  console.log(`[PWA Service] Server running on port ${PORT}`);
  console.log(`[PWA Service] Environment: ${process.env.NODE_ENV || 'development'}`);
  console.log(`[PWA Service] Ready to serve Progressive Web App`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('[PWA Service] SIGTERM received, shutting down gracefully');
  process.exit(0);
});

process.on('SIGINT', () => {
  console.log('[PWA Service] SIGINT received, shutting down gracefully');
  process.exit(0);
});
