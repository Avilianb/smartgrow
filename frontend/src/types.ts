export interface DeviceStatus {
  device_id: string;
  timestamp: string;
  temperature_c: number;
  humidity_pct: number;
  soil_status: 'dry' | 'optimal' | 'wet';
  soil_raw: number;
  rain_status: 'raining' | 'no_rain';
  pump_state: 'on' | 'off';
  shade_state: 'open' | 'closed' | 'partial';
  today_plan: {
    planned_volume_l: number;
    executed_volume_l: number;
  };
}

export interface SensorHistoryPoint {
  time: string;
  temp: number;
  humidity: number;
  soil: number;
}

export interface LogEntry {
  id: number;
  timestamp: string;
  level: 'INFO' | 'WARN' | 'ERROR';
  message: string;
  device_id: string;
}

export interface LocationConfig {
  latitude: number;
  longitude: number;
  address: string;
}

export interface WeatherForecast {
  date: string;
  temp_max: number;
  temp_min: number;
  condition: string;
  precip_mm: number;
}