import React, { lazy, Suspense } from 'react';
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { Sidebar } from './components/Sidebar';
import { MobileBottomNav } from './components/MobileBottomNav';

// 懒加载页面组件 - 只在需要时才加载
const Dashboard = lazy(() => import('./pages/Dashboard').then(m => ({ default: m.Dashboard })));
const LocationManager = lazy(() => import('./pages/LocationManager').then(m => ({ default: m.LocationManager })));
const SystemLogs = lazy(() => import('./pages/SystemLogs').then(m => ({ default: m.SystemLogs })));
const Login = lazy(() => import('./pages/Login').then(m => ({ default: m.Login })));
const UserManagement = lazy(() => import('./pages/UserManagement').then(m => ({ default: m.UserManagement })));
const ChangePassword = lazy(() => import('./pages/ChangePassword').then(m => ({ default: m.ChangePassword })));

// 加载中组件
const LoadingFallback = () => (
  <div className="flex items-center justify-center h-screen">
    <div className="flex flex-col items-center gap-3">
      <div className="w-12 h-12 border-4 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
      <p className="text-sm text-slate-500">加载中...</p>
    </div>
  </div>
);

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
    return <LoadingFallback />;
  }

  return (
    <Routes>
      {/* 登录页面 */}
      <Route
        path="/login"
        element={
          isAuthenticated ? (
            <Navigate to="/" replace />
          ) : (
            <Suspense fallback={<LoadingFallback />}>
              <Login />
            </Suspense>
          )
        }
      />

      {/* 需要认证的路由 */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout>
              <Suspense fallback={<LoadingFallback />}>
                <Dashboard />
              </Suspense>
            </Layout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/location"
        element={
          <ProtectedRoute>
            <Layout>
              <Suspense fallback={<LoadingFallback />}>
                <LocationManager />
              </Suspense>
            </Layout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/logs"
        element={
          <ProtectedRoute>
            <Layout>
              <Suspense fallback={<LoadingFallback />}>
                <SystemLogs />
              </Suspense>
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
              <Suspense fallback={<LoadingFallback />}>
                <UserManagement />
              </Suspense>
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
              <Suspense fallback={<LoadingFallback />}>
                <ChangePassword />
              </Suspense>
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