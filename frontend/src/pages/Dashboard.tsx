import React, { useState, useEffect } from 'react';
import { Droplets, Thermometer, CloudRain, Sun, Play, RotateCcw, Umbrella } from 'lucide-react';
import { StatCard } from '../components/StatCard';
import { MultiLineChart } from '../components/Charts';
import { getDeviceStatus, getHistory, triggerIrrigation, recomputePlan } from '../services/api';
import type { DeviceStatus, SensorHistoryPoint } from '../types';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export const Dashboard: React.FC = () => {
  const [status, setStatus] = useState<DeviceStatus | null>(null);
  const [history, setHistory] = useState<SensorHistoryPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [dataAge, setDataAge] = useState<string>('');
  const [expandedCard, setExpandedCard] = useState<string | null>(null);

  useEffect(() => {
    // Fetch data from API
    const fetchData = async () => {
      try {
        const [statusData, historyData] = await Promise.all([
          getDeviceStatus(),
          getHistory()
        ]);
        setStatus(statusData);
        setHistory(historyData);

        // 计算数据年龄
        if (statusData.timestamp) {
          const dataTime = new Date(statusData.timestamp);
          const now = new Date();
          const diffMs = now.getTime() - dataTime.getTime();
          const diffMins = Math.floor(diffMs / 60000);
          const diffHours = Math.floor(diffMins / 60);
          const diffDays = Math.floor(diffHours / 24);

          if (diffDays > 0) {
            setDataAge(`${diffDays}天前`);
          } else if (diffHours > 0) {
            setDataAge(`${diffHours}小时前`);
          } else if (diffMins > 0) {
            setDataAge(`${diffMins}分钟前`);
          } else {
            setDataAge('刚刚');
          }
        }
      } catch (error) {
        console.error('Error fetching data:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchData();

    // 10秒自动刷新
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleWaterNow = async () => {
    const success = await triggerIrrigation(2.0);
    alert(success ? "指令已发送: 触发手动灌溉。" : "发送失败，请重试");
  };

  const handleShadeToggle = () => {
    alert("指令已发送: 切换遮阳机构状态。");
  };

  const handleRecalculate = async () => {
    const success = await recomputePlan();
    alert(success ? "请求已发送: 重新计算15天灌溉计划。" : "计算失败，请重试");
  };

  const getShadeStateText = (state: string) => {
    switch(state) {
        case 'open': return '开启';
        case 'closed': return '关闭';
        case 'partial': return '半开';
        default: return state;
    }
  };

  const toggleCard = (cardId: string) => {
    setExpandedCard(expandedCard === cardId ? null : cardId);
  };

  const renderExpandedChart = (cardId: string) => {
    if (expandedCard !== cardId) return null;

    const chartConfigs = {
      soil: { dataKey: 'soil_raw' as const, color: '#3b82f6', unit: 'ADC', name: '土壤湿度' },
      temp: { dataKey: 'temp' as const, color: '#f97316', unit: '°C', name: '温度' },
      humidity: { dataKey: 'humidity' as const, color: '#06b6d4', unit: '%', name: '空气湿度' },
    };

    const config = chartConfigs[cardId as keyof typeof chartConfigs];
    if (!config) return null;

    return (
      <div
        key={cardId}
        className="bg-white rounded-3xl p-6 shadow-sm border border-slate-100 animate-scaleIn origin-top"
      >
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-bold text-slate-800">{config.name} - 24小时趋势</h2>
          <div className="flex items-center gap-1">
            <span className="w-2 h-2 rounded-full" style={{ backgroundColor: config.color }}></span>
            <span className="text-xs font-medium text-slate-500">{config.name}</span>
          </div>
        </div>
        <div className="h-[300px]">
          {history.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={history} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                <defs>
                  <linearGradient id={`color${cardId}`} x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={config.color} stopOpacity={0.3}/>
                    <stop offset="95%" stopColor={config.color} stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis
                  dataKey="time"
                  tick={{fontSize: 12, fill: '#94a3b8'}}
                  tickLine={false}
                  axisLine={false}
                  interval={4}
                />
                <YAxis
                  tick={{fontSize: 12, fill: '#94a3b8'}}
                  tickLine={false}
                  axisLine={false}
                  unit={config.unit}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#fff',
                    borderRadius: '12px',
                    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                    border: 'none',
                    padding: '8px 12px'
                  }}
                  itemStyle={{ color: '#1e293b', fontWeight: 600 }}
                  labelStyle={{ color: '#64748b', fontSize: '12px', marginBottom: '4px' }}
                  formatter={(value: number) => value.toFixed(1)}
                />
                <Area
                  type="monotone"
                  dataKey={config.dataKey}
                  stroke={config.color}
                  strokeWidth={2.5}
                  fillOpacity={1}
                  fill={`url(#color${cardId})`}
                  animationDuration={1000}
                />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-full text-slate-400">
              <div className="text-center">
                <p className="text-lg font-medium mb-2">暂无图表数据</p>
                <p className="text-sm">等待 ESP32 设备上传传感器数据...</p>
              </div>
            </div>
          )}
        </div>
      </div>
    );
  };

  if (loading || !status) {
    return (
      <div className="flex items-center justify-center h-full min-h-[500px]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-8 animate-fadeIn">
      <header className="mb-8">
        <h1 className="text-2xl font-bold text-slate-800">系统仪表盘</h1>
        <div className="flex items-center gap-3">
          <p className="text-slate-500">设备实时概览: {status.device_id}</p>
          {dataAge && dataAge !== '刚刚' && !dataAge.includes('分钟') && (
            <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-700">
              ⚠️ 数据已过期 ({dataAge})
            </span>
          )}
        </div>
      </header>

      {/* Top Stats Row */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* 土壤湿度卡片 */}
        <div onClick={() => toggleCard('soil')}>
          <StatCard
            title="土壤湿度"
            value={status.soil_raw}
            unit="ADC"
            icon={<Droplets size={24} />}
            statusColor="bg-blue-50 text-blue-600"
            subtext={status.soil_status === 'optimal' ? '湿度适宜' : '需注意'}
            isExpandable={true}
            isExpanded={expandedCard === 'soil'}
          />
        </div>

        {/* 温度卡片 */}
        <div onClick={() => toggleCard('temp')}>
          <StatCard
            title="温度"
            value={status.temperature_c}
            unit="°C"
            icon={<Thermometer size={24} />}
            statusColor="bg-orange-50 text-orange-600"
            trend="+1.2°"
            trendUp={false}
            isExpandable={true}
            isExpanded={expandedCard === 'temp'}
          />
        </div>

        {/* 空气湿度卡片 */}
        <div onClick={() => toggleCard('humidity')}>
          <StatCard
            title="空气湿度"
            value={status.humidity_pct}
            unit="%"
            icon={<CloudRain size={24} />}
            statusColor="bg-cyan-50 text-cyan-600"
            isExpandable={true}
            isExpanded={expandedCard === 'humidity'}
          />
        </div>

        {/* 今日计划卡片 - 不可展开 */}
        <StatCard
          title="今日计划"
          value={status.today_plan.planned_volume_l}
          unit="L"
          icon={<Sun size={24} />}
          subtext={`已完成 ${((status.today_plan.executed_volume_l / status.today_plan.planned_volume_l) * 100).toFixed(0)}%`}
          statusColor="bg-purple-50 text-purple-600"
        />
      </div>

      {/* 展开的图表区域 - 与4个标签同宽 */}
      {expandedCard && (
        <div className="mt-6">
          {renderExpandedChart(expandedCard)}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Chart Section */}
        <div className="lg:col-span-2 bg-white rounded-3xl p-6 shadow-sm border border-slate-100 hover:shadow-lg transition-shadow duration-300">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-lg font-bold text-slate-800">24小时环境趋势</h2>
              {history.length === 0 && (
                <p className="text-xs text-amber-600 mt-1">⚠️ 暂无历史数据</p>
              )}
            </div>
            <div className="flex gap-2">
              <span className="flex items-center gap-1 text-xs font-medium text-slate-500">
                <span className="w-2 h-2 rounded-full bg-orange-500"></span> 温度
              </span>
              <span className="flex items-center gap-1 text-xs font-medium text-slate-500">
                <span className="w-2 h-2 rounded-full bg-blue-500"></span> 湿度
              </span>
            </div>
          </div>
          <div className="h-[300px]">
            {history.length > 0 ? (
              <MultiLineChart data={history} />
            ) : (
              <div className="flex items-center justify-center h-full text-slate-400">
                <div className="text-center">
                  <p className="text-lg font-medium mb-2">暂无图表数据</p>
                  <p className="text-sm">等待 ESP32 设备上传传感器数据...</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Control & Status Panel */}
        <div className="space-y-6">
            {/* Quick Actions */}
            <div className="bg-white rounded-3xl p-6 shadow-sm border border-slate-100">
                <h3 className="text-lg font-bold text-slate-800 mb-4">快捷操作</h3>
                <div className="space-y-3">
                    <button 
                        onClick={handleWaterNow}
                        className="w-full flex items-center justify-between p-4 rounded-xl bg-blue-50 text-blue-700 hover:bg-blue-100 transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-white rounded-lg shadow-sm group-hover:scale-110 transition-transform">
                                <Droplets size={18} className="text-blue-500" />
                            </div>
                            <span className="font-semibold">立即浇水</span>
                        </div>
                        <Play size={16} />
                    </button>

                    <button 
                        onClick={handleShadeToggle}
                        className="w-full flex items-center justify-between p-4 rounded-xl bg-slate-50 text-slate-700 hover:bg-slate-100 transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-white rounded-lg shadow-sm group-hover:scale-110 transition-transform">
                                <Umbrella size={18} className="text-slate-500" />
                            </div>
                            <span className="font-semibold">切换遮阳</span>
                        </div>
                        <span className="text-xs uppercase tracking-wider font-bold bg-white px-2 py-1 rounded text-slate-400">
                            {getShadeStateText(status.shade_state)}
                        </span>
                    </button>

                    <button 
                        onClick={handleRecalculate}
                        className="w-full flex items-center justify-between p-4 rounded-xl bg-primary-50 text-primary-700 hover:bg-primary-100 transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-white rounded-lg shadow-sm group-hover:scale-110 transition-transform">
                                <RotateCcw size={18} className="text-primary-500" />
                            </div>
                            <span className="font-semibold">重新计算计划</span>
                        </div>
                    </button>
                </div>
            </div>
            
            {/* Status Summary */}
            <div className="bg-gradient-to-br from-slate-800 to-slate-900 rounded-3xl p-6 text-white shadow-lg relative overflow-hidden">
                <div className="absolute top-0 right-0 w-32 h-32 bg-white opacity-5 rounded-full -mr-10 -mt-10"></div>
                <h3 className="text-lg font-bold mb-4 relative z-10">系统状态</h3>
                <div className="space-y-4 relative z-10">
                    <div className="flex justify-between items-center border-b border-slate-700 pb-2">
                        <span className="text-slate-400 text-sm">水泵状态</span>
                        <span className={`font-mono font-bold ${status.pump_state === 'on' ? 'text-green-400' : 'text-slate-200'}`}>
                            {status.pump_state === 'on' ? '开启' : '关闭'}
                        </span>
                    </div>
                    <div className="flex justify-between items-center border-b border-slate-700 pb-2">
                        <span className="text-slate-400 text-sm">雨水传感器</span>
                        <span className="font-mono font-bold text-slate-200">{status.rain_status === 'raining' ? '有雨' : '干燥'}</span>
                    </div>
                    <div className="flex justify-between items-center">
                        <span className="text-slate-400 text-sm">最后更新</span>
                        <span className={`text-xs font-medium ${dataAge === '刚刚' || dataAge.includes('分钟') ? 'text-green-400' : 'text-amber-400'}`}>
                            {dataAge || '未知'}
                        </span>
                    </div>
                </div>
            </div>
        </div>
      </div>
    </div>
  );
};