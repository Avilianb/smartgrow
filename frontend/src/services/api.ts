import type { DeviceStatus, SensorHistoryPoint, LogEntry, WeatherForecast } from '../types';

// 获取认证token
const getAuthToken = (): string | null => {
  return localStorage.getItem('auth_token');
};

// 获取设备ID（从localStorage或使用默认值）
const getDeviceId = (): string => {
  return localStorage.getItem('device_id') || 'esp32s3-1';
};

// 通用fetch请求配置
const getHeaders = (): HeadersInit => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  const token = getAuthToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return headers;
};

// 动态获取API地址
const getApiBase = (): string => {
  // 如果是生产环境（CDN/HTTPS），使用相对路径
  // 这样可以让CDN/反向代理来处理API请求转发
  if (window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    // 使用相对路径，让代理服务器转发到后端
    return '/api';
  }
  // 开发环境使用localhost
  return 'http://localhost:8080/api';
};

const API_BASE = getApiBase();

console.log('API Base URL:', API_BASE); // 调试信息
console.log('Current hostname:', window.location.hostname);
console.log('Current protocol:', window.location.protocol);

// 获取设备状态
export const getDeviceStatus = async (): Promise<DeviceStatus> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(`${API_BASE}/device/${DEVICE_ID}/status`, {
      headers: getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch status');
    const data = await response.json();
    return {
      device_id: data.device_id,
      timestamp: data.timestamp,
      temperature_c: data.temperature_c,
      humidity_pct: data.humidity_pct,
      soil_status: data.soil_status,
      soil_raw: data.soil_raw || 2150,
      rain_status: data.rain_status,
      pump_state: data.pump_state,
      shade_state: data.shade_state,
      today_plan: data.today_plan
    };
  } catch (error) {
    console.error('Error fetching device status:', error);
    // 返回默认值
    return {
      device_id: DEVICE_ID,
      timestamp: new Date().toISOString(),
      temperature_c: 28.5,
      humidity_pct: 65.2,
      soil_status: 'optimal',
      soil_raw: 2150,
      rain_status: 'no_rain',
      pump_state: 'off',
      shade_state: 'closed',
      today_plan: { planned_volume_l: 2.5, executed_volume_l: 1.2 }
    };
  }
};

// 获取历史数据
export const getHistory = async (): Promise<SensorHistoryPoint[]> => {
  const DEVICE_ID = getDeviceId();
  try {
    const endTime = new Date().toISOString();
    const startTime = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    const response = await fetch(
      `${API_BASE}/device/${DEVICE_ID}/history?start_time=${startTime}&end_time=${endTime}&limit=48`,
      { headers: getHeaders() }
    );
    if (!response.ok) throw new Error('Failed to fetch history');
    const result = await response.json();

    if (result.data && result.data.length > 0) {
      // 转换为图表所需格式
      return result.data.map((item: any) => {
        const date = new Date(item.timestamp);
        return {
          time: date.getHours().toString().padStart(2, '0') + ':' +
                date.getMinutes().toString().padStart(2, '0'),
          temp: item.temperature_c || 0,
          humidity: item.humidity_pct || 0,
          soil: item.soil_raw || 0
        };
      });
    } else {
      console.warn('No history data available from backend, using empty array');
      return [];
    }
  } catch (error) {
    console.error('Error fetching history:', error);
    return [];
  }
};

// 获取日志
export const getLogs = async (limit: number = 20, offset: number = 0): Promise<{ data: LogEntry[], total: number }> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(
      `${API_BASE}/device/${DEVICE_ID}/logs?limit=${limit}&offset=${offset}`,
      { headers: getHeaders() }
    );
    if (!response.ok) throw new Error('Failed to fetch logs');
    const result = await response.json();
    return {
      data: result.data || [],
      total: result.total || 0
    };
  } catch (error) {
    console.error('Error fetching logs:', error);
    return {
      data: [],
      total: 0
    };
  }
};

// 获取设备位置
export const getLocation = async (): Promise<{ latitude: number; longitude: number; address?: string } | null> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(`${API_BASE}/location/${DEVICE_ID}`, {
      headers: getHeaders(),
    });
    if (!response.ok) {
      // 404是预期的情况（首次使用时没有保存过位置）
      if (response.status === 404) {
        console.info('📍 暂无保存的位置信息，使用默认位置');
        return null;
      }
      throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
  } catch (error) {
    // 只有非404错误才打印error级别日志
    if (error instanceof Error && !error.message.includes('404')) {
      console.error('获取位置信息失败:', error);
    }
    return null;
  }
};

// 更新设备位置
export const updateLocation = async (latitude: number, longitude: number): Promise<boolean> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(`${API_BASE}/location/${DEVICE_ID}`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ latitude, longitude })
    });
    return response.ok;
  } catch (error) {
    console.error('Error updating location:', error);
    return false;
  }
};

// 更新天气预报（需要经纬度）
export const updateForecast = async (latitude?: number, longitude?: number): Promise<boolean> => {
  const DEVICE_ID = getDeviceId();
  try {
    // 如果没有提供经纬度，先获取设备位置
    if (latitude === undefined || longitude === undefined) {
      const location = await getLocation();
      if (!location) {
        console.info('⛅ 暂无位置信息，无法更新天气预报');
        return false;
      }
      latitude = location.latitude;
      longitude = location.longitude;
    }

    const response = await fetch(`${API_BASE}/forecast/update`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        device_id: DEVICE_ID,
        latitude: latitude,
        longitude: longitude
      })
    });
    return response.ok;
  } catch (error) {
    console.error('更新天气预报失败:', error);
    return false;
  }
};

// 获取天气预报数据（从数据库）
export const getForecastData = async (): Promise<any[]> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(`${API_BASE}/device/${DEVICE_ID}/status`, {
      headers: getHeaders(),
    });
    if (!response.ok) return [];
    await response.json();
    // TODO: 后端需要返回天气预报数据
    return [];
  } catch (error) {
    console.error('Error fetching forecast data:', error);
    return [];
  }
};

// 获取天气预报
export const getForecast = async (): Promise<WeatherForecast[]> => {
  try {
    // 先触发更新天气预报（静默失败）
    await updateForecast();

    // 等待一小段时间让后端处理
    await new Promise(resolve => setTimeout(resolve, 500));

    // 从数据库查询天气预报
    const response = await fetch(`${API_BASE}/forecast?days=5`, {
      headers: getHeaders(),
    });
    if (!response.ok) throw new Error('Failed to fetch forecast');

    const result = await response.json();

    if (result.success && result.data && result.data.length > 0) {
      return result.data.map((item: any) => {
        const date = new Date(item.date);
        return {
          date: date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric', weekday: 'short' }),
          temp_max: item.temp_max || 0,
          temp_min: item.temp_min || 0,
          condition: getWeatherCondition(item.raw_json),
          precip_mm: item.precip_mm || 0
        };
      });
    }
  } catch (error) {
    console.info('⛅ 暂无天气数据，显示默认天气信息');
  }

  // 返回默认数据
  const forecasts: WeatherForecast[] = [];
  for (let i = 0; i < 5; i++) {
    const d = new Date();
    d.setDate(d.getDate() + i);
    forecasts.push({
      date: d.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric', weekday: 'short' }),
      temp_max: 28,
      temp_min: 20,
      condition: '晴朗',
      precip_mm: 0
    });
  }
  return forecasts;
};

// 辅助函数：从原始JSON中提取天气状况
const getWeatherCondition = (rawJson: string): string => {
  try {
    const data = JSON.parse(rawJson);
    return data.textDay || '晴朗';
  } catch {
    return '晴朗';
  }
};

// 手动灌溉
export const triggerIrrigation = async (volume: number): Promise<boolean> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(`${API_BASE}/device/${DEVICE_ID}/irrigate`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ volume_l: volume, reason: 'manual_trigger' })
    });
    return response.ok;
  } catch (error) {
    console.error('Error triggering irrigation:', error);
    return false;
  }
};

// 重新计算计划
export const recomputePlan = async (): Promise<boolean> => {
  const DEVICE_ID = getDeviceId();
  try {
    const response = await fetch(`${API_BASE}/plan/recompute?device_id=${DEVICE_ID}`, {
      method: 'POST',
      headers: getHeaders(),
    });
    return response.ok;
  } catch (error) {
    console.error('Error recomputing plan:', error);
    return false;
  }
};

// 兼容原有的 mock 接口名称
export const getMockDeviceStatus = getDeviceStatus;
export const getMockHistory = getHistory;
export const getMockLogs = getLogs;
export const getMockForecast = getForecast;
export const getMockLocation = getLocation;
