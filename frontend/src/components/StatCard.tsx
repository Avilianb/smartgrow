import React from 'react';
import type { ReactNode } from 'react';

interface StatCardProps {
  title: string;
  value: string | number;
  unit?: string;
  icon: ReactNode;
  trend?: string;
  trendUp?: boolean; // true = up is good/green, false = up is bad/red
  statusColor?: string;
  subtext?: string;
  isExpandable?: boolean;
  isExpanded?: boolean;
}

export const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  unit,
  icon,
  trend,
  trendUp = true,
  statusColor,
  subtext,
  isExpandable = false,
}) => {
  return (
    <div className={`
      relative bg-white rounded-3xl p-6
      shadow-[0_2px_10px_-3px_rgba(6,81,237,0.1)]
      border border-slate-100
      transition-all duration-300 ease-in-out
      hover:-translate-y-1 hover:shadow-[0_20px_25px_-5px_rgba(0,0,0,0.1),0_10px_10px_-5px_rgba(0,0,0,0.04)]
      group ${isExpandable ? 'cursor-pointer' : 'cursor-default'} overflow-hidden
    `}>
      {/* Decorative background circle */}
      <div className="absolute -right-6 -top-6 w-24 h-24 bg-gradient-to-br from-primary-50 to-transparent rounded-full opacity-50 group-hover:scale-150 transition-transform duration-500 ease-out" />

      <div className="flex items-start justify-between relative z-10">
        <div>
          <p className="text-sm font-medium text-slate-500 mb-1">{title}</p>
          <div className="flex items-baseline gap-1">
            <h3 className="text-3xl font-bold text-slate-800 tracking-tight">
              {value}
            </h3>
            {unit && <span className="text-sm font-medium text-slate-400">{unit}</span>}
          </div>
          {subtext && <p className="mt-2 text-xs text-slate-400 font-medium">{subtext}</p>}
        </div>

        <div className={`
          p-3 rounded-2xl
          ${statusColor ? statusColor : 'bg-primary-50 text-primary-600'}
          transition-colors duration-300 group-hover:bg-primary-500 group-hover:text-white
        `}>
          {icon}
        </div>
      </div>

      {trend && (
        <div className="mt-4 flex items-center text-xs font-medium">
          <span className={`
            px-2 py-0.5 rounded-full mr-2
            ${trendUp ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}
          `}>
            {trend}
          </span>
          <span className="text-slate-400">对比24h前</span>
        </div>
      )}

    </div>
  );
};