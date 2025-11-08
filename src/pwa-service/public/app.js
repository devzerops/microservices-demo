/**
 * PWA App Initialization and Offline Manager
 */

class PWAManager {
  constructor() {
    this.swRegistration = null;
    this.deferredPrompt = null;
    this.isOnline = navigator.onLine;

    this.init();
  }

  async init() {
    // Register service worker
    if ('serviceWorker' in navigator) {
      await this.registerServiceWorker();
    }

    // Setup install prompt
    this.setupInstallPrompt();

    // Setup online/offline handlers
    this.setupConnectivityHandlers();

    // Setup push notifications
    this.setupPushNotifications();

    // Initialize offline storage
    this.initializeOfflineStorage();

    console.log('[PWA] Initialization complete');
  }

  async registerServiceWorker() {
    try {
      this.swRegistration = await navigator.serviceWorker.register('/service-worker.js');

      console.log('[PWA] Service Worker registered:', this.swRegistration);

      // Check for updates
      this.swRegistration.addEventListener('updatefound', () => {
        const newWorker = this.swRegistration.installing;

        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            this.showUpdateNotification();
          }
        });
      });

      // Listen for messages from SW
      navigator.serviceWorker.addEventListener('message', (event) => {
        console.log('[PWA] Message from SW:', event.data);
      });

    } catch (error) {
      console.error('[PWA] Service Worker registration failed:', error);
    }
  }

  setupInstallPrompt() {
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault();
      this.deferredPrompt = e;

      console.log('[PWA] Install prompt available');

      // Show install button
      this.showInstallButton();
    });

    window.addEventListener('appinstalled', () => {
      console.log('[PWA] App installed successfully');
      this.deferredPrompt = null;
      this.hideInstallButton();
    });
  }

  showInstallButton() {
    const installButton = document.getElementById('install-button');
    if (installButton) {
      installButton.style.display = 'block';

      installButton.addEventListener('click', async () => {
        if (!this.deferredPrompt) return;

        this.deferredPrompt.prompt();

        const { outcome } = await this.deferredPrompt.userChoice;
        console.log('[PWA] Install prompt outcome:', outcome);

        this.deferredPrompt = null;
        this.hideInstallButton();
      });
    }
  }

  hideInstallButton() {
    const installButton = document.getElementById('install-button');
    if (installButton) {
      installButton.style.display = 'none';
    }
  }

  setupConnectivityHandlers() {
    window.addEventListener('online', () => {
      this.isOnline = true;
      console.log('[PWA] Connection restored');
      this.showNotification('Back online!', 'Your connection has been restored.');

      // Trigger background sync
      this.triggerBackgroundSync();
    });

    window.addEventListener('offline', () => {
      this.isOnline = false;
      console.log('[PWA] Connection lost');
      this.showNotification('You are offline', 'Some features may be limited.');
    });

    // Update UI based on connection status
    this.updateConnectionStatus();
  }

  updateConnectionStatus() {
    const statusElement = document.getElementById('connection-status');
    if (statusElement) {
      statusElement.textContent = this.isOnline ? 'Online' : 'Offline';
      statusElement.className = this.isOnline ? 'online' : 'offline';
    }
  }

  async setupPushNotifications() {
    if (!('PushManager' in window)) {
      console.log('[PWA] Push notifications not supported');
      return;
    }

    // Check permission
    if (Notification.permission === 'default') {
      // Show opt-in UI
      this.showNotificationOptIn();
    } else if (Notification.permission === 'granted') {
      await this.subscribeToPush();
    }
  }

  showNotificationOptIn() {
    const optInElement = document.getElementById('notification-opt-in');
    if (optInElement) {
      optInElement.style.display = 'block';

      const enableButton = document.getElementById('enable-notifications');
      if (enableButton) {
        enableButton.addEventListener('click', async () => {
          const permission = await Notification.requestPermission();

          if (permission === 'granted') {
            await this.subscribeToPush();
            optInElement.style.display = 'none';
          }
        });
      }
    }
  }

  async subscribeToPush() {
    try {
      const subscription = await this.swRegistration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: this.urlBase64ToUint8Array(
          'YOUR_VAPID_PUBLIC_KEY' // Replace with actual key
        )
      });

      console.log('[PWA] Push subscription:', subscription);

      // Send subscription to server
      await fetch('/api/push/subscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(subscription)
      });

    } catch (error) {
      console.error('[PWA] Push subscription failed:', error);
    }
  }

  urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);

    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i);
    }

    return outputArray;
  }

  async triggerBackgroundSync() {
    if ('sync' in this.swRegistration) {
      try {
        await this.swRegistration.sync.register('sync-cart');
        console.log('[PWA] Background sync registered');
      } catch (error) {
        console.error('[PWA] Background sync failed:', error);
      }
    }
  }

  async initializeOfflineStorage() {
    if (!('indexedDB' in window)) {
      console.log('[PWA] IndexedDB not supported');
      return;
    }

    // Open database
    const db = await this.openDB();
    console.log('[PWA] Offline storage initialized');
  }

  openDB() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('OfflineStore', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve(request.result);

      request.onupgradeneeded = (event) => {
        const db = event.target.result;

        if (!db.objectStoreNames.contains('pendingCartItems')) {
          db.createObjectStore('pendingCartItems', { keyPath: 'id', autoIncrement: true });
        }

        if (!db.objectStoreNames.contains('cachedProducts')) {
          db.createObjectStore('cachedProducts', { keyPath: 'id' });
        }

        if (!db.objectStoreNames.contains('userPreferences')) {
          db.createObjectStore('userPreferences', { keyPath: 'key' });
        }
      };
    });
  }

  showNotification(title, message) {
    const notification = document.createElement('div');
    notification.className = 'pwa-notification';
    notification.innerHTML = `
      <strong>${title}</strong>
      <p>${message}</p>
    `;

    document.body.appendChild(notification);

    setTimeout(() => {
      notification.classList.add('show');
    }, 100);

    setTimeout(() => {
      notification.classList.remove('show');
      setTimeout(() => notification.remove(), 300);
    }, 3000);
  }

  showUpdateNotification() {
    const updateNotification = document.getElementById('update-notification');
    if (updateNotification) {
      updateNotification.style.display = 'block';

      const reloadButton = document.getElementById('reload-app');
      if (reloadButton) {
        reloadButton.addEventListener('click', () => {
          // Tell the SW to take control immediately
          if (this.swRegistration.waiting) {
            this.swRegistration.waiting.postMessage({ type: 'SKIP_WAITING' });
          }

          window.location.reload();
        });
      }
    }
  }

  // Public API for adding items to offline cart
  async addToOfflineCart(product) {
    const db = await this.openDB();
    const transaction = db.transaction(['pendingCartItems'], 'readwrite');
    const store = transaction.objectStore('pendingCartItems');

    const item = {
      ...product,
      addedAt: new Date(),
      synced: false
    };

    await store.add(item);

    console.log('[PWA] Added to offline cart:', item);

    // Try to sync immediately if online
    if (this.isOnline) {
      await this.triggerBackgroundSync();
    }
  }

  // Public API for getting offline cart
  async getOfflineCart() {
    const db = await this.openDB();
    const transaction = db.transaction(['pendingCartItems'], 'readonly');
    const store = transaction.objectStore('pendingCartItems');

    return new Promise((resolve, reject) => {
      const request = store.getAll();
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }
}

// Initialize PWA Manager when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    window.pwaManager = new PWAManager();
  });
} else {
  window.pwaManager = new PWAManager();
}

// Export for use in other scripts
window.PWAManager = PWAManager;
