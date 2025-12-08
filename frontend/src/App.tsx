import React from 'react';
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { Sidebar } from './components/Sidebar';
import { MobileBottomNav } from './components/MobileBottomNav';
import { Dashboard } from './pages/Dashboard';
import { LocationManager } from './pages/LocationManager';
import { SystemLogs } from './pages/SystemLogs';
import { Login } from './pages/Login';
import { UserManagement } from './pages/UserManagement';
import { ChangePassword } from './pages/ChangePassword';

// 受保护的路由组件
const ProtectedRoute: React.FC<{ children: React.ReactNode; adminOnly?: boolean }> = ({
  children,
  adminOnly = false
}) => {
  const { isAuthenticated, isAdmin, isLoading } = useAuth();

  // 等待认证状态加载完成
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-gray-600">加载中...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (adminOnly && !isAdmin) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="flex min-h-screen bg-[#F8FAFC]">
      <Sidebar />
      <main className="flex-1 pt-16 pb-16 md:pt-0 md:pb-0 md:ml-64 transition-all duration-300">
        <div className="max-w-7xl mx-auto">
          {children}
        </div>
      </main>
      {/* 移动端底部导航栏 */}
      <MobileBottomNav />
    </div>
  );
};

const AppRoutes: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth();

  // 等待认证状态加载完成
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-gray-600">加载中...</div>
      </div>
    );
  }

  return (
    <Routes>
      {/* 登录页面 */}
      <Route
        path="/login"
        element={isAuthenticated ? <Navigate to="/" replace /> : <Login />}
      />

      {/* 需要认证的路由 */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout>
              <Dashboard />
            </Layout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/location"
        element={
          <ProtectedRoute>
            <Layout>
              <LocationManager />
            </Layout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/logs"
        element={
          <ProtectedRoute>
            <Layout>
              <SystemLogs />
            </Layout>
          </ProtectedRoute>
        }
      />

      {/* 管理员专用路由 */}
      <Route
        path="/admin/users"
        element={
          <ProtectedRoute adminOnly>
            <Layout>
              <UserManagement />
            </Layout>
          </ProtectedRoute>
        }
      />

      {/* 用户设置路由 */}
      <Route
        path="/change-password"
        element={
          <ProtectedRoute>
            <Layout>
              <ChangePassword />
            </Layout>
          </ProtectedRoute>
        }
      />

      {/* 404重定向 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <HashRouter>
        <AppRoutes />
      </HashRouter>
    </AuthProvider>
  );
};

export default App;