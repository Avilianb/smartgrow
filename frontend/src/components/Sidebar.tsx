import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, MapPin, ScrollText, Sprout, Users, LogOut, Key } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

export const Sidebar: React.FC = () => {
  const { isAdmin, user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const navItems = [
    { to: '/', icon: <LayoutDashboard size={20} />, label: '仪表盘', show: true },
    { to: '/location', icon: <MapPin size={20} />, label: '位置管理', show: true },
    { to: '/logs', icon: <ScrollText size={20} />, label: '系统日志', show: true },
    { to: '/admin/users', icon: <Users size={20} />, label: '用户管理', show: isAdmin },
    { to: '/change-password', icon: <Key size={20} />, label: '修改密码', show: true },
  ];

  return (
    <>
      {/* 移动端顶部栏 */}
      <div className="fixed top-0 left-0 right-0 h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 md:hidden z-50">
        <div className="flex items-center gap-2 text-primary-600">
          <div className="bg-primary-50 p-1.5 rounded-lg">
            <Sprout size={20} className="text-primary-600" />
          </div>
          <span className="font-bold text-lg tracking-tight text-slate-800">SmartGrow</span>
        </div>

        <button
          onClick={handleLogout}
          className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-red-500 hover:bg-red-50 transition-colors"
          aria-label="退出登录"
        >
          <LogOut size={18} />
          <span className="text-sm font-medium">退出</span>
        </button>
      </div>

      {/* 侧边栏 - 仅桌面端显示 */}
      <aside className="hidden md:flex fixed left-0 top-0 h-screen w-64 bg-white border-r border-slate-100 flex-col z-50">
        {/* Brand - 桌面端显示 */}
        <div className="h-20 flex items-center px-8 border-b border-slate-50">
          <div className="flex items-center gap-3 text-primary-600">
            <div className="bg-primary-50 p-2 rounded-xl">
              <Sprout size={24} className="text-primary-600" />
            </div>
            <span className="font-bold text-xl tracking-tight text-slate-800">SmartGrow</span>
          </div>
        </div>

        {/* User Info */}
        {user && (
          <div className="px-4 py-4 border-b border-slate-50">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-full flex items-center justify-center text-white font-semibold">
                {user.username.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0">
                <div className="font-medium text-slate-800 truncate">{user.username}</div>
                <div className="text-xs text-slate-500">
                  {isAdmin ? '管理员' : '用户'}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Navigation */}
        <nav className="flex-1 py-8 px-4 space-y-2 overflow-y-auto">
          {navItems.filter(item => item.show).map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => `
                flex items-center gap-3 px-4 py-3.5 rounded-xl transition-all duration-200 group
                ${isActive
                  ? 'bg-primary-50 text-primary-700 font-semibold shadow-sm'
                  : 'text-slate-500 hover:bg-slate-50 hover:text-slate-900'}
              `}
            >
              {item.icon}
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        {/* Footer Actions */}
        <div className="p-4 border-t border-slate-50 space-y-2">
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-3 text-red-400 hover:text-red-500 hover:bg-red-50 rounded-xl w-full transition-all"
          >
            <LogOut size={18} />
            <span className="text-sm font-medium">退出登录</span>
          </button>
        </div>
      </aside>
    </>
  );
};
