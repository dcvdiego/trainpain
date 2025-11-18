import { apiClient } from './api';
import type { RouteReliabilityQuery, RouteReliabilityResponse } from '../types';

export const routeService = {
  async getReliability(query: RouteReliabilityQuery): Promise<RouteReliabilityResponse> {
    const response = await apiClient.post<RouteReliabilityResponse>('/routes/reliability', query);
    return response.data;
  },
};
