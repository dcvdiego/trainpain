import { Station } from './station';

export interface RouteReliabilityQuery {
  origin_crs: string;
  destination_crs: string;
  time_start?: string;
  time_end?: string;
  day_filter?: string;
  analysis_days?: number;
}

export interface RouteReliabilityMetrics {
  id: number;
  route_id: number;
  time_period: string;
  start_time?: string;
  end_time?: string;
  day_filter: string;
  analysis_start_date: string;
  analysis_end_date: string;
  total_services_analyzed: number;
  cancellation_rate: number;
  on_time_rate: number;
  avg_delay_minutes: number;
  median_delay_minutes: number;
  p95_delay_minutes: number;
  p99_delay_minutes: number;
  pct_0_5_min_late: number;
  pct_5_15_min_late: number;
  pct_15_30_min_late: number;
  pct_30_plus_min_late: number;
  reliability_score: number;
  best_day_of_week?: string;
  worst_day_of_week?: string;
  most_common_delay_reason?: string;
  computed_at: string;
}

export interface RouteReliabilityResponse {
  origin: Station;
  destination: Station;
  metrics: RouteReliabilityMetrics;
  query: RouteReliabilityQuery;
}
