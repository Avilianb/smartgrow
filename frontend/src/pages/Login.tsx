import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isAdminLogin, setIsAdminLogin] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [usernameFocused, setUsernameFocused] = useState(false);
  const [passwordFocused, setPasswordFocused] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const trimmedUsername = username.trim();
    const trimmedPassword = password.trim();

    if (!trimmedUsername || !trimmedPassword) {
      setError('用户名和密码不能为空');
      setLoading(false);
      return;
    }

    try {
      const endpoint = isAdminLogin ? '/api/auth/admin/login' : '/api/auth/login';
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: trimmedUsername,
          password: trimmedPassword
        }),
      });

      const data = await response.json();

      if (!response.ok || !data.success) {
        setError(data.message || '登录失败');
        setLoading(false);
        return;
      }

      let deviceId: string | undefined;
      if (data.user.role === 'user') {
        try {
          const tokenParts = data.token.split('.');
          if (tokenParts.length === 3) {
            const payload = JSON.parse(atob(tokenParts[1]));
            deviceId = payload.device_id;
          }
        } catch (err) {
          console.error('Failed to extract device_id from token:', err);
        }
      }

      login(data.token, data.user, deviceId);

      if (data.user.role === 'admin') {
        navigate('/admin/users');
      } else {
        navigate('/');
      }
    } catch (err) {
      console.error('Login error:', err);
      setError('网络错误，请检查服务器连接');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#0A0E27] relative overflow-hidden flex items-center justify-center">
      {/* 装饰性几何形状 */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {/* 左上角粉色圆角矩形 */}
        <div className="absolute -top-20 -left-20 w-80 h-80 bg-gradient-to-br from-pink-400 to-red-400 rounded-[60px] transform rotate-12 opacity-90"></div>

        {/* 右上角黄色三角形 */}
        <div className="absolute -top-10 right-20 w-0 h-0 border-l-[200px] border-l-transparent border-r-[200px] border-r-transparent border-b-[300px] border-b-yellow-400 opacity-90"></div>

        {/* 右下角粉红色半圆 */}
        <div className="absolute -bottom-32 -right-32 w-96 h-96 bg-gradient-to-tl from-pink-500 to-red-500 rounded-full opacity-80"></div>

        {/* 左下角蓝色三角形 */}
        <div className="absolute bottom-10 left-10 w-0 h-0 border-t-[150px] border-t-transparent border-b-[150px] border-b-transparent border-r-[200px] border-r-blue-500 opacity-70"></div>

        {/* 中间黄色圆形 */}
        <div className="absolute bottom-40 left-1/3 w-64 h-64 bg-gradient-to-br from-yellow-300 to-yellow-500 rounded-full opacity-40"></div>

        {/* 小装饰圆点 */}
        <div className="absolute top-1/4 left-20 w-16 h-16 bg-yellow-300 rounded-full opacity-60"></div>
        <div className="absolute bottom-1/3 right-1/4 w-12 h-12 bg-pink-400 rounded-full opacity-50"></div>
        <div className="absolute top-2/3 left-1/4 w-8 h-8 bg-blue-400 rounded-full opacity-60"></div>
      </div>

      {/* 右上角仪表盘预览卡片 */}
      <div className="absolute top-20 right-20 w-96 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-3xl p-6 shadow-2xl transform rotate-3 hover:rotate-0 transition-transform duration-500 hidden lg:block z-10">
        <div className="text-white space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">系统仪表盘预览</h3>
            <div className="w-8 h-8 bg-white/20 rounded-full flex items-center justify-center">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                <path fillRule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clipRule="evenodd"/>
              </svg>
            </div>
          </div>
          <div className="bg-white/10 rounded-2xl p-4 backdrop-blur-sm">
            <div className="flex gap-3 mb-3">
              <div className="flex-1 bg-green-400/30 rounded-xl p-3 text-center">
                <div className="text-xs opacity-80">温度</div>
                <div className="text-xl font-bold">28°C</div>
              </div>
              <div className="flex-1 bg-blue-400/30 rounded-xl p-3 text-center">
                <div className="text-xs opacity-80">湿度</div>
                <div className="text-xl font-bold">65%</div>
              </div>
            </div>
            <div className="flex gap-2">
              <div className="w-3 h-3 bg-blue-400 rounded-full"></div>
              <div className="w-3 h-3 bg-pink-400 rounded-full"></div>
              <div className="w-3 h-3 bg-yellow-400 rounded-full"></div>
              <div className="w-3 h-3 bg-green-400 rounded-full"></div>
            </div>
          </div>
        </div>
      </div>

      {/* 登录卡片 */}
      <div className="relative z-20 w-full max-w-md mx-4">
        <div className="bg-white rounded-3xl shadow-2xl p-10 backdrop-blur-lg">
          {/* Logo */}
          <div className="flex items-center gap-2 mb-8">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-lg flex items-center justify-center">
              <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <span className="text-xl font-bold text-gray-800">SmartGrow</span>
          </div>

          {/* 标题 */}
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            欢迎使用
          </h1>
          <h2 className="text-3xl font-bold text-gray-900 mb-6">
            SmartGrow
          </h2>
          <p className="text-gray-500 mb-8">
            {isAdminLogin ? '管理员登录' : '用户登录'}
          </p>

          {error && (
            <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 rounded-r-xl">
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
                <span className="text-sm text-red-700">{error}</span>
              </div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* 用户名输入框 */}
            <div className="relative">
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onFocus={() => setUsernameFocused(true)}
                onBlur={() => setUsernameFocused(false)}
                required
                className="w-full px-0 py-3 bg-transparent border-0 border-b-2 border-gray-300
                         focus:border-blue-600 focus:outline-none transition-colors duration-300
                         text-gray-900 text-base peer placeholder-transparent"
                placeholder="用户名"
              />
              <label
                htmlFor="username"
                className="absolute left-0 top-3 text-base pointer-events-none flex"
              >
                {'用户名'.split('').map((char, index) => (
                  <span
                    key={index}
                    className="inline-block transition-all duration-300 ease-out"
                    style={{
                      transform: username ? 'translateY(-32px) scale(0.75)' : 'translateY(0) scale(1)',
                      color: username ? '#2563eb' : '#9ca3af',
                      transitionDelay: username ? `${index * 50}ms` : `${(2 - index) * 50}ms`,
                      transformOrigin: 'left center'
                    }}
                  >
                    {char}
                  </span>
                ))}
              </label>
              <div className={`absolute bottom-0 left-0 h-0.5 bg-gradient-to-r from-blue-600 to-indigo-600 transition-all duration-300 ${usernameFocused ? 'w-full' : 'w-0'}`}></div>
            </div>

            {/* 密码输入框 */}
            <div className="relative">
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onFocus={() => setPasswordFocused(true)}
                onBlur={() => setPasswordFocused(false)}
                required
                className="w-full px-0 py-3 bg-transparent border-0 border-b-2 border-gray-300
                         focus:border-blue-600 focus:outline-none transition-colors duration-300
                         text-gray-900 text-base peer placeholder-transparent"
                placeholder="密码"
              />
              <label
                htmlFor="password"
                className="absolute left-0 top-3 text-base pointer-events-none flex"
              >
                {'密码'.split('').map((char, index) => (
                  <span
                    key={index}
                    className="inline-block transition-all duration-300 ease-out"
                    style={{
                      transform: password ? 'translateY(-32px) scale(0.75)' : 'translateY(0) scale(1)',
                      color: password ? '#2563eb' : '#9ca3af',
                      transitionDelay: password ? `${index * 50}ms` : `${(1 - index) * 50}ms`,
                      transformOrigin: 'left center'
                    }}
                  >
                    {char}
                  </span>
                ))}
              </label>
              <div className={`absolute bottom-0 left-0 h-0.5 bg-gradient-to-r from-blue-600 to-indigo-600 transition-all duration-300 ${passwordFocused ? 'w-full' : 'w-0'}`}></div>
            </div>

            {/* 登录按钮 */}
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white py-4 rounded-2xl font-semibold text-lg
                       hover:shadow-2xl hover:shadow-blue-500/50 transition-all duration-300 transform hover:scale-[1.02] active:scale-[0.98]
                       disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100
                       relative overflow-hidden group mt-8"
            >
              <span className="relative z-10 flex items-center justify-center gap-2">
                {loading ? (
                  <>
                    <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    登录中...
                  </>
                ) : (
                  '登录'
                )}
              </span>
              <div className="absolute inset-0 bg-gradient-to-r from-indigo-600 to-blue-600 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            </button>

            {/* 切换登录类型 */}
            <div className="text-center pt-4">
              <button
                type="button"
                onClick={() => {
                  setIsAdminLogin(!isAdminLogin);
                  setError('');
                }}
                className="text-blue-600 hover:text-blue-700 font-medium text-sm transition-colors inline-flex items-center gap-1"
              >
                {isAdminLogin ? '切换到用户登录' : '管理员登录'}
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>
          </form>
        </div>

        {/* 底部版权 */}
        <div className="mt-6 text-center text-sm text-gray-400">
          © 2025 SmartGrow. All rights reserved.
        </div>
      </div>

      {/* 添加样式 */}
      <style>{`
        @keyframes float {
          0%, 100% { transform: translateY(0px); }
          50% { transform: translateY(-20px); }
        }
        .animate-float {
          animation: float 6s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
};
