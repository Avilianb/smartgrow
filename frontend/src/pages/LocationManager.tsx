import React, { useState, useEffect } from 'react';
import { MapPin, Save, Cloud, Droplets, Locate, Map as MapIcon, Loader2 } from 'lucide-react';
import { getForecast, getLocation, updateLocation } from '../services/api';
import { MapContainer, TileLayer, Marker, Popup, useMap, useMapEvents } from 'react-leaflet';
import L from 'leaflet';
import type { WeatherForecast } from '../types';

// Fix for default Leaflet icons in webpack/importmap environments
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
});

// Component to handle map clicks
const LocationSelector = ({ onLocationSelect }: { onLocationSelect: (lat: number, lng: number) => void }) => {
  useMapEvents({
    click(e) {
      onLocationSelect(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
};

// Component to recenter map when coords change externally
const RecenterMap = ({ lat, lng }: { lat: number; lng: number }) => {
  const map = useMap();
  useEffect(() => {
    map.setView([lat, lng], map.getZoom());
  }, [lat, lng, map]);
  return null;
};

export const LocationManager: React.FC = () => {
  const [lat, setLat] = useState<string>("39.92");
  const [lng, setLng] = useState<string>("116.41");
  const [address, setAddress] = useState<string>("加载中...");
  const [isLocating, setIsLocating] = useState(false);
  const [forecast, setForecast] = useState<WeatherForecast[]>([]);
  const [isSaving, setIsSaving] = useState(false);

  // 初始加载位置和天气数据
  useEffect(() => {
    const loadData = async () => {
      // 加载位置
      const location = await getLocation();
      if (location) {
        setLat(location.latitude.toFixed(4));
        setLng(location.longitude.toFixed(4));
        // 根据坐标范围显示大致地区
        setAddress(getRegionName(location.latitude, location.longitude));
      } else {
        setAddress("中国, 北京 (默认)");
      }

      // 加载天气预报
      const weatherData = await getForecast();
      setForecast(weatherData);
    };

    loadData();
  }, []);

  // 根据经纬度返回大致地区名称
  const getRegionName = (lat: number, lng: number): string => {
    // 简单的地区判断
    if (lat >= 39 && lat <= 41 && lng >= 115 && lng <= 118) {
      return "中国, 北京";
    } else if (lat >= 30 && lat <= 32 && lng >= 120 && lng <= 122) {
      return "中国, 上海";
    } else if (lat >= 22 && lat <= 24 && lng >= 113 && lng <= 115) {
      return "中国, 广州";
    } else if (lat >= 29 && lat <= 31 && lng >= 119 && lng <= 121) {
      return "中国, 杭州";
    } else if (lat >= 30 && lat <= 32 && lng >= 103 && lng <= 105) {
      return "中国, 成都";
    } else if (lat >= 28 && lat <= 30 && lng >= 120 && lng <= 122) {
      return "中国, 温州地区";
    } else {
      return `自定义位置 (${lat.toFixed(2)}, ${lng.toFixed(2)})`;
    }
  };

  // Helper to safely parse coordinates
  const getCoords = () => {
    const l = parseFloat(lat);
    const n = parseFloat(lng);
    return {
      lat: isNaN(l) ? 39.92 : l,
      lng: isNaN(n) ? 116.41 : n
    };
  };

  const handleMapClick = (newLat: number, newLng: number) => {
    setLat(newLat.toFixed(4));
    setLng(newLng.toFixed(4));
    setAddress(getRegionName(newLat, newLng));
  };

  const handleCurrentLocation = () => {
    if (!navigator.geolocation) {
      alert("浏览器不支持地理定位。");
      return;
    }

    // 检查是否为安全上下文（HTTPS或localhost）
    const isSecureContext = window.isSecureContext ||
                           window.location.protocol === 'https:' ||
                           window.location.hostname === 'localhost' ||
                           window.location.hostname === '127.0.0.1';

    if (!isSecureContext) {
      alert("⚠️ GPS定位功能需要HTTPS连接\n\n" +
            "当前使用的是HTTP连接，浏览器出于安全考虑限制了GPS功能。\n\n" +
            "解决方案：\n" +
            "1. 使用HTTPS访问此页面\n" +
            "2. 或者手动在地图上点击选择位置\n" +
            "3. 或者手动输入经纬度");
      return;
    }

    setIsLocating(true);

    navigator.geolocation.getCurrentPosition(
      (position) => {
        const newLat = position.coords.latitude;
        const newLng = position.coords.longitude;
        setLat(newLat.toFixed(4));
        setLng(newLng.toFixed(4));
        setAddress(`GPS定位: ${getRegionName(newLat, newLng)}`);
        setIsLocating(false);
      },
      (error) => {
        let msg = "无法获取您的位置";
        switch(error.code) {
          case error.PERMISSION_DENIED:
            msg = "❌ 用户拒绝了定位请求\n\n" +
                  "请在浏览器设置中允许访问位置信息：\n" +
                  "• Chrome: 地址栏左侧锁图标 → 网站设置 → 位置\n" +
                  "• Firefox: 地址栏左侧锁图标 → 权限 → 位置\n\n" +
                  "或者您可以手动在地图上点击选择位置。";
            break;
          case error.POSITION_UNAVAILABLE:
            msg = "❌ 位置信息不可用\n\n" +
                  "可能的原因：\n" +
                  "• 设备没有GPS模块\n" +
                  "• GPS信号弱或被屏蔽\n" +
                  "• 定位服务未开启\n\n" +
                  "建议：手动在地图上点击选择位置。";
            break;
          case error.TIMEOUT:
            msg = "⏱️ 定位请求超时\n\n" +
                  "GPS定位耗时过长，请重试或手动选择位置。";
            break;
        }
        alert(msg);
        setIsLocating(false);
      },
      {
        enableHighAccuracy: true, // 启用高精度GPS
        timeout: 10000,
        maximumAge: 0
      }
    );
  };

  const handleSaveLocation = async () => {
    setIsSaving(true);
    const latitude = parseFloat(lat);
    const longitude = parseFloat(lng);

    if (isNaN(latitude) || isNaN(longitude)) {
      alert("❌ 请输入有效的经纬度");
      setIsSaving(false);
      return;
    }

    try {
      console.log('Saving location:', { latitude, longitude, apiUrl: window.location.origin });
      const success = await updateLocation(latitude, longitude);
      if (success) {
        alert("✅ 位置已保存！正在获取最新天气预报...");
        // 重新获取天气预报
        const weatherData = await getForecast();
        setForecast(weatherData);
      } else {
        alert("❌ 保存失败\n\n" +
              "可能的原因：\n" +
              "• 后端服务器未启动\n" +
              "• 网络连接问题\n" +
              "• API地址配置错误\n\n" +
              `当前API地址: ${window.location.protocol}//${window.location.hostname}:8080\n\n` +
              "请检查后端服务器是否运行在8080端口。");
      }
    } catch (error) {
      console.error('Save location error:', error);
      alert("❌ 保存位置时发生错误\n\n" +
            `错误信息: ${error}\n\n` +
            `API地址: ${window.location.protocol}//${window.location.hostname}:8080`);
    }
    setIsSaving(false);
  };

  const coords = getCoords();

  return (
    <div className="p-8 space-y-8 animate-[fadeIn_0.5s_ease-out]">
      <header>
        <h1 className="text-2xl font-bold text-slate-800">位置与天气</h1>
        <p className="text-slate-500">管理设备位置以获取准确天气数据</p>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Map Card */}
        <div className="lg:col-span-2 bg-white rounded-3xl p-6 shadow-sm border border-slate-100 flex flex-col min-h-[600px]">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-lg font-bold text-slate-800 flex items-center gap-2">
              <MapIcon size={20} className="text-primary-500"/>
              设备位置
            </h2>
            <button
              onClick={handleCurrentLocation}
              disabled={isLocating}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-semibold transition-colors
                ${isLocating
                  ? 'bg-slate-100 text-slate-400 cursor-wait'
                  : 'bg-primary-50 text-primary-700 hover:bg-primary-100'
                }`}
            >
              {isLocating ? <Loader2 size={16} className="animate-spin" /> : <Locate size={16} />}
              {isLocating ? '正在定位...' : '使用当前精确位置'}
            </button>
          </div>

          <div className="flex-1 rounded-2xl overflow-hidden relative border border-slate-200 shadow-inner z-0">
             {/* Map Container */}
             <MapContainer
                center={[coords.lat, coords.lng]}
                zoom={10}
                scrollWheelZoom={true}
                style={{ height: "100%", width: "100%" }}
             >
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
                <Marker position={[coords.lat, coords.lng]}>
                  <Popup>
                    设备位置<br />
                    {coords.lat.toFixed(4)}, {coords.lng.toFixed(4)}
                  </Popup>
                </Marker>
                <LocationSelector onLocationSelect={handleMapClick} />
                <RecenterMap lat={coords.lat} lng={coords.lng} />
             </MapContainer>

             {/* Floating Address Badge */}
             <div className="absolute top-4 left-1/2 -translate-x-1/2 z-[400] bg-white/90 backdrop-blur-md px-4 py-2 rounded-full shadow-lg border border-slate-100 flex items-center gap-2 pointer-events-none">
                <MapPin size={16} className="text-red-500 fill-red-500" />
                <span className="text-sm font-bold text-slate-700">{address}</span>
                <span className="text-xs text-slate-400 border-l border-slate-200 pl-2 ml-1">
                  {coords.lat.toFixed(2)}, {coords.lng.toFixed(2)}
                </span>
             </div>
          </div>

          <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-4">
             <div>
                <label className="block text-xs font-semibold text-slate-500 mb-1">纬度 (Latitude)</label>
                <input
                    type="number"
                    step="0.0001"
                    value={lat}
                    onChange={e => setLat(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2 text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 transition-all font-mono text-sm"
                />
             </div>
             <div>
                <label className="block text-xs font-semibold text-slate-500 mb-1">经度 (Longitude)</label>
                <input
                    type="number"
                    step="0.0001"
                    value={lng}
                    onChange={e => setLng(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2 text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 transition-all font-mono text-sm"
                />
             </div>
             <div className="flex items-end">
                <button
                  onClick={handleSaveLocation}
                  disabled={isSaving}
                  className={`w-full font-semibold py-2 px-4 rounded-xl flex items-center justify-center gap-2 transition-colors shadow-lg
                    ${isSaving
                      ? 'bg-slate-400 text-white cursor-wait'
                      : 'bg-primary-600 hover:bg-primary-700 text-white shadow-primary-500/30'
                    }`}
                >
                    {isSaving ? <Loader2 size={18} className="animate-spin" /> : <Save size={18} />}
                    {isSaving ? '保存中...' : '保存坐标'}
                </button>
             </div>
          </div>
        </div>

        {/* Forecast Card */}
        <div className="bg-white rounded-3xl p-6 shadow-sm border border-slate-100 h-fit">
            <h2 className="text-lg font-bold text-slate-800 mb-6 flex items-center gap-2">
                <Cloud size={20} className="text-blue-500"/>
                未来5天预报
            </h2>
            <div className="space-y-4">
                {forecast.map((day, idx) => (
                    <div key={idx} className="group flex items-center justify-between p-3 rounded-2xl hover:bg-slate-50 transition-colors border border-transparent hover:border-slate-100 cursor-default">
                        <div className="flex items-center gap-4">
                            <div className="w-12 text-sm font-bold text-slate-400">{day.date}</div>
                            <div className="flex flex-col">
                                <span className="font-semibold text-slate-700 text-sm">{day.condition}</span>
                                <div className="flex items-center gap-1 text-xs text-slate-400">
                                    <Droplets size={10} />
                                    {day.precip_mm.toFixed(1)}mm
                                </div>
                            </div>
                        </div>
                        <div className="text-right">
                            <span className="block font-bold text-slate-800">{day.temp_max.toFixed(0)}°</span>
                            <span className="block text-xs text-slate-400">{day.temp_min.toFixed(0)}°</span>
                        </div>
                    </div>
                ))}
            </div>
            <div className="mt-6 pt-6 border-t border-slate-100">
                <div className="bg-blue-50 rounded-xl p-4 border border-blue-100">
                    <p className="text-xs text-blue-700 font-medium leading-relaxed">
                        <span className="font-bold block mb-1">AI 建议</span>
                        根据该地区的天气预报，预计周三有大雨。系统将跳过预定的灌溉计划以节约用水。
                    </p>
                </div>
            </div>
        </div>
      </div>
    </div>
  );
};
