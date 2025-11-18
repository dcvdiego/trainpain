import { apiClient } from './api';
import { RouteReliabilityQuery, RouteReliabilityResponse } from '../types/route';

export const routeService = {
  async getReliability(query: RouteReliabilityQuery): Promise<RouteReliabilityResponse> {
    const response = await apiClient.post<RouteReliabilityResponse>('/routes/reliability', query);
    return response.data;
  },
};
