import { apiClient } from './api';
import { PushSubscriptionData } from '../utils/pwa';

export interface PineappleClass {
  id: number;
  class_name: string;
  instructor: string;
  level: string;
  style: string;
  description: string;
  duration_minutes: number;
  created_at: string;
  updated_at: string;
}

export interface PineappleClassSchedule {
  id: number;
  class_id: number;
  day_of_week: number; // 0=Sunday, 1=Monday, ..., 6=Saturday
  start_time: string; // HH:MM:SS
  end_time: string; // HH:MM:SS
  room: string;
  max_capacity: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PineappleClassWithSchedule extends PineappleClass {
  schedules: PineappleClassSchedule[];
}

export interface PineappleSubscription {
  id: number;
  schedule_id: number;
  push_endpoint: string;
  push_p256dh: string;
  push_auth: string;
  notification_threshold: number;
  notify_2h_before: boolean;
  notify_1h_before: boolean;
  notify_30m_before: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface SubscribeRequest {
  schedule_id: number;
  push_endpoint: string;
  push_p256dh: string;
  push_auth: string;
  notification_threshold?: number;
  notify_2h_before?: boolean;
  notify_1h_before?: boolean;
  notify_30m_before?: boolean;
}

/**
 * Get all classes with their schedules
 */
export async function getAllClasses(): Promise<PineappleClassWithSchedule[]> {
  const response = await apiClient.get<{ classes: PineappleClassWithSchedule[] }>(
    '/pineapple/classes'
  );
  return response.data.classes;
}

/**
 * Get classes for a specific day of the week
 * @param dayOfWeek 0=Sunday, 1=Monday, ..., 6=Saturday
 */
export async function getClassesByDay(dayOfWeek: number): Promise<PineappleClassWithSchedule[]> {
  const response = await apiClient.get<{ classes: PineappleClassWithSchedule[] }>(
    `/pineapple/classes/day/${dayOfWeek}`
  );
  return response.data.classes;
}

/**
 * Subscribe to a class schedule for notifications
 */
export async function subscribe(
  scheduleId: number,
  pushSubscription: PushSubscriptionData,
  options?: {
    notificationThreshold?: number;
    notify2hBefore?: boolean;
    notify1hBefore?: boolean;
    notify30mBefore?: boolean;
  }
): Promise<PineappleSubscription> {
  const request: SubscribeRequest = {
    schedule_id: scheduleId,
    push_endpoint: pushSubscription.endpoint,
    push_p256dh: pushSubscription.p256dh,
    push_auth: pushSubscription.auth,
    notification_threshold: options?.notificationThreshold || 5,
    notify_2h_before: options?.notify2hBefore !== false, // default true
    notify_1h_before: options?.notify1hBefore || false,
    notify_30m_before: options?.notify30mBefore || false
  };

  const response = await apiClient.post<{ subscription: PineappleSubscription }>(
    '/pineapple/subscribe',
    request
  );

  return response.data.subscription;
}

/**
 * Unsubscribe from a class schedule
 */
export async function unsubscribe(subscriptionId: number): Promise<void> {
  await apiClient.delete(`/pineapple/subscribe/${subscriptionId}`);
}

/**
 * Get all subscriptions for a user (identified by push endpoint)
 */
export async function getUserSubscriptions(
  pushEndpoint: string
): Promise<PineappleSubscription[]> {
  const response = await apiClient.get<{ subscriptions: PineappleSubscription[] }>(
    '/pineapple/subscriptions',
    {
      params: { endpoint: pushEndpoint }
    }
  );

  return response.data.subscriptions;
}

/**
 * Trigger manual scrape of classes (admin function)
 */
export async function triggerScrape(): Promise<void> {
  await apiClient.post('/pineapple/scrape');
}

/**
 * Get day of week name
 */
export function getDayName(dayOfWeek: number): string {
  const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  return days[dayOfWeek] || 'Unknown';
}

/**
 * Format time from HH:MM:SS to HH:MM
 */
export function formatTime(time: string): string {
  return time.substring(0, 5);
}

/**
 * Get current day of week (0=Sunday, 1=Monday, ..., 6=Saturday)
 */
export function getCurrentDayOfWeek(): number {
  return new Date().getDay();
}
