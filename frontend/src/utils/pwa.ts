// PWA utilities for service worker and push notifications

export interface PushSubscriptionData {
  endpoint: string;
  p256dh: string;
  auth: string;
}

// VAPID public key - should match backend configuration
// This will need to be set via environment variable
export const VAPID_PUBLIC_KEY = import.meta.env.VITE_VAPID_PUBLIC_KEY || '';

/**
 * Register the service worker
 */
export async function registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
  if (!('serviceWorker' in navigator)) {
    console.warn('Service workers are not supported in this browser');
    return null;
  }

  try {
    const registration = await navigator.serviceWorker.register('/sw.js', {
      scope: '/'
    });

    console.log('Service worker registered successfully:', registration);

    // Wait for the service worker to be ready
    await navigator.serviceWorker.ready;
    console.log('Service worker is ready');

    return registration;
  } catch (error) {
    console.error('Service worker registration failed:', error);
    return null;
  }
}

/**
 * Request notification permission from the user
 */
export async function requestNotificationPermission(): Promise<NotificationPermission> {
  if (!('Notification' in window)) {
    console.warn('Notifications are not supported in this browser');
    return 'denied';
  }

  const permission = await Notification.requestPermission();
  console.log('Notification permission:', permission);
  return permission;
}

/**
 * Subscribe to push notifications
 */
export async function subscribeToPushNotifications(
  registration: ServiceWorkerRegistration
): Promise<PushSubscriptionData | null> {
  try {
    // Check if already subscribed
    let subscription = await registration.pushManager.getSubscription();

    // If not subscribed, create a new subscription
    if (!subscription) {
      const vapidPublicKey = urlBase64ToUint8Array(VAPID_PUBLIC_KEY);

      subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: vapidPublicKey
      });

      console.log('Push subscription created:', subscription);
    } else {
      console.log('Already subscribed to push notifications');
    }

    // Extract subscription data
    const subscriptionData = extractSubscriptionData(subscription);
    return subscriptionData;
  } catch (error) {
    console.error('Failed to subscribe to push notifications:', error);
    return null;
  }
}

/**
 * Unsubscribe from push notifications
 */
export async function unsubscribeFromPushNotifications(
  registration: ServiceWorkerRegistration
): Promise<boolean> {
  try {
    const subscription = await registration.pushManager.getSubscription();

    if (subscription) {
      const success = await subscription.unsubscribe();
      console.log('Unsubscribed from push notifications:', success);
      return success;
    }

    return false;
  } catch (error) {
    console.error('Failed to unsubscribe from push notifications:', error);
    return false;
  }
}

/**
 * Get current push subscription
 */
export async function getCurrentPushSubscription(
  registration: ServiceWorkerRegistration
): Promise<PushSubscriptionData | null> {
  try {
    const subscription = await registration.pushManager.getSubscription();

    if (subscription) {
      return extractSubscriptionData(subscription);
    }

    return null;
  } catch (error) {
    console.error('Failed to get push subscription:', error);
    return null;
  }
}

/**
 * Extract subscription data in a format suitable for backend
 */
function extractSubscriptionData(subscription: PushSubscription): PushSubscriptionData {
  const keys = subscription.toJSON().keys!;

  return {
    endpoint: subscription.endpoint,
    p256dh: keys.p256dh || '',
    auth: keys.auth || ''
  };
}

/**
 * Convert base64 string to Uint8Array for VAPID key
 */
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }

  return outputArray;
}

/**
 * Check if push notifications are supported
 */
export function isPushNotificationSupported(): boolean {
  return (
    'serviceWorker' in navigator &&
    'PushManager' in window &&
    'Notification' in window
  );
}

/**
 * Check if user has granted notification permission
 */
export function hasNotificationPermission(): boolean {
  if (!('Notification' in window)) {
    return false;
  }

  return Notification.permission === 'granted';
}

/**
 * Initialize PWA features
 * Call this once when the app starts
 */
export async function initializePWA(): Promise<ServiceWorkerRegistration | null> {
  console.log('Initializing PWA features...');

  // Register service worker
  const registration = await registerServiceWorker();

  if (!registration) {
    console.warn('Service worker registration failed, PWA features unavailable');
    return null;
  }

  // Check if notifications are supported
  if (!isPushNotificationSupported()) {
    console.warn('Push notifications are not supported in this browser');
    return registration;
  }

  // Check current permission status
  const currentPermission = Notification.permission;
  console.log('Current notification permission:', currentPermission);

  // If permission is default, we'll ask when user tries to subscribe
  // If permission is denied, we can't do anything
  // If permission is granted, we could auto-restore subscriptions

  return registration;
}
