import React from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import type { SensorHistoryPoint } from '../types';

interface ChartProps {
  data: SensorHistoryPoint[];
  dataKey: keyof SensorHistoryPoint;
  color: string;
  unit: string;
}

export const TrendChart: React.FC<ChartProps> = ({ data, dataKey, color, unit }) => {
  // 自定义 Tooltip 格式化函数，保留一位小数
  const formatTooltipValue = (value: number) => {
    return value.toFixed(1);
  };

  return (
    <div className="w-full h-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id={`color${dataKey}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.2}/>
              <stop offset="95%" stopColor={color} stopOpacity={0}/>
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
            unit={unit}
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
            formatter={formatTooltipValue}
          />
          <Area
            type="monotone"
            dataKey={dataKey}
            stroke={color}
            strokeWidth={3}
            fillOpacity={1}
            fill={`url(#color${dataKey})`}
            animationDuration={1500}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
};

// 新增：多线图表组件
interface MultiLineChartProps {
  data: SensorHistoryPoint[];
}

export const MultiLineChart: React.FC<MultiLineChartProps> = ({ data }) => {
  // 自定义 Tooltip 格式化函数，保留一位小数
  const formatTooltipValue = (value: number) => {
    return value.toFixed(1);
  };

  return (
    <div className="w-full h-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id="colorTemp" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#f97316" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#f97316" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorHumidity" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
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
            yAxisId="left"
            tick={{fontSize: 12, fill: '#94a3b8'}}
            tickLine={false}
            axisLine={false}
            unit="°C"
          />
          <YAxis
            yAxisId="right"
            orientation="right"
            tick={{fontSize: 12, fill: '#94a3b8'}}
            tickLine={false}
            axisLine={false}
            unit="%"
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
            formatter={formatTooltipValue}
          />
          <Area
            yAxisId="left"
            type="monotone"
            dataKey="temp"
            stroke="#f97316"
            strokeWidth={2.5}
            fillOpacity={1}
            fill="url(#colorTemp)"
            animationDuration={1500}
            name="温度"
          />
          <Area
            yAxisId="right"
            type="monotone"
            dataKey="humidity"
            stroke="#3b82f6"
            strokeWidth={2.5}
            fillOpacity={1}
            fill="url(#colorHumidity)"
            animationDuration={1500}
            name="湿度"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
};