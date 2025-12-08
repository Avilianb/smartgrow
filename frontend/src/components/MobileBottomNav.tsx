import React from 'react';
import { NavLink } from 'react-router-dom';
import { LayoutDashboard, MapPin, ScrollText, User } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

/**
 * 移动端底部导航栏组件
 * 为手机用户提供更便捷的导航体验
 */
export const MobileBottomNav: React.FC = () => {
  const { isAdmin } = useAuth();

  const navItems = [
    { to: '/', icon: LayoutDashboard, label: '仪表盘' },
    { to: '/location', icon: MapPin, label: '位置' },
    { to: '/logs', icon: ScrollText, label: '日志' },
    { to: isAdmin ? '/admin/users' : '/change-password', icon: User, label: isAdmin ? '用户' : '我的' },
  ];

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-slate-200 z-40 safe-area-bottom">
      <div className="grid grid-cols-4 h-16">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) => `
              flex flex-col items-center justify-center gap-1 transition-colors
              ${isActive
                ? 'text-primary-600 font-semibold'
                : 'text-slate-400 hover:text-slate-600'
              }
            `}
          >
            {({ isActive }) => (
              <>
                <item.icon size={22} strokeWidth={isActive ? 2.5 : 2} />
                <span className="text-xs">{item.label}</span>
              </>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  );
};
