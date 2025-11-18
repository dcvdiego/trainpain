import { apiClient } from './api';
import { StationSearchResult } from '../types/station';

export const stationService = {
  async search(query: string): Promise<StationSearchResult[]> {
    const response = await apiClient.get<{ results: StationSearchResult[]; count: number }>(
      '/stations',
      { params: { q: query } }
    );
    return response.data.results;
  },

  async getByCRS(crs: string) {
    const response = await apiClient.get(`/stations/${crs}`);
    return response.data;
  },
};
