import React, { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';

interface User {
  id: number;
  username: string;
  role: 'admin' | 'user';
  created_at: string;
  updated_at: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  deviceId: string | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isLoading: boolean;
  login: (token: string, user: User, deviceId?: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [deviceId, setDeviceId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // 从localStorage加载认证信息
  useEffect(() => {
    const storedToken = localStorage.getItem('auth_token');
    const storedUser = localStorage.getItem('auth_user');
    const storedDeviceId = localStorage.getItem('device_id');

    if (storedToken && storedUser) {
      try {
        const parsedUser = JSON.parse(storedUser);
        setToken(storedToken);
        setUser(parsedUser);
        setDeviceId(storedDeviceId);
      } catch (error) {
        console.error('Failed to parse stored user:', error);
        // 清除无效的存储数据
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
        localStorage.removeItem('device_id');
      }
    }
    setIsLoading(false);
  }, []);

  const login = (newToken: string, newUser: User, newDeviceId?: string) => {
    setToken(newToken);
    setUser(newUser);
    setDeviceId(newDeviceId || null);

    // 保存到localStorage
    localStorage.setItem('auth_token', newToken);
    localStorage.setItem('auth_user', JSON.stringify(newUser));
    if (newDeviceId) {
      localStorage.setItem('device_id', newDeviceId);
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    setDeviceId(null);

    // 清除localStorage
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
    localStorage.removeItem('device_id');
  };

  const value: AuthContextType = {
    user,
    token,
    deviceId,
    isAuthenticated: !!token && !!user,
    isAdmin: user?.role === 'admin',
    isLoading,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
