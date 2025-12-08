import type { DeviceStatus, SensorHistoryPoint, LogEntry, WeatherForecast } from '../types';

// Simulate backend data
export const getMockDeviceStatus = (): DeviceStatus => ({
  device_id: "esp32s3-1",
  timestamp: new Date().toISOString(),
  temperature_c: 28.5,
  humidity_pct: 65.2,
  soil_status: 'optimal',
  soil_raw: 2150,
  rain_status: 'no_rain',
  pump_state: 'off',
  shade_state: 'closed',
  today_plan: {
    planned_volume_l: 2.5,
    executed_volume_l: 1.2
  }
});

export const getMockHistory = (): SensorHistoryPoint[] => {
  const data: SensorHistoryPoint[] = [];
  const now = new Date();
  for (let i = 24; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60 * 60 * 1000);
    data.push({
      time: time.getHours().toString().padStart(2, '0') + ':00',
      temp: 20 + Math.random() * 10 - (i > 18 || i < 6 ? 5 : 0), // Cooler at night
      humidity: 50 + Math.random() * 20,
      soil: 2000 + Math.random() * 500
    });
  }
  return data;
};

export const getMockLogs = (): LogEntry[] => [
  { id: 1, timestamp: '2025-11-25 14:30:01', level: 'INFO', message: '收到心跳包', device_id: 'esp32s3-1' },
  { id: 2, timestamp: '2025-11-25 14:15:00', level: 'INFO', message: '计划灌溉已启动', device_id: 'esp32s3-1' },
  { id: 3, timestamp: '2025-11-25 13:00:00', level: 'WARN', message: '检测到高温 (31°C), 正在调整遮阳', device_id: 'esp32s3-1' },
  { id: 4, timestamp: '2025-11-25 10:00:00', level: 'ERROR', message: '天气 API 请求超时', device_id: 'server' },
  { id: 5, timestamp: '2025-11-25 09:30:00', level: 'INFO', message: '每日计划已重新计算', device_id: 'planner' },
];

export const getMockForecast = (): WeatherForecast[] => Array.from({ length: 5 }).map((_, i) => {
  const d = new Date();
  d.setDate(d.getDate() + i);
  // Simple mapping for demo
  const isRain = Math.random() > 0.7;
  return {
    date: d.toLocaleDateString('zh-CN', { weekday: 'short', month: 'numeric', day: 'numeric' }),
    temp_max: 28 + Math.random() * 5,
    temp_min: 18 + Math.random() * 3,
    condition: isRain ? '小雨' : '晴朗',
    precip_mm: isRain ? Math.random() * 10 : 0
  };
});