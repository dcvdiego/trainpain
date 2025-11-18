export interface Station {
  id: number;
  crs_code?: string;
  station_name: string;
  latitude?: number;
  longitude?: number;
  station_type?: string;
  tfl_station_id?: string;
  zone?: string;
  created_at: string;
  updated_at: string;
}

export interface StationSearchResult {
  id: number;
  crs_code?: string;
  station_name: string;
  station_type?: string;
}
