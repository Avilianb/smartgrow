import { useState, useEffect } from 'react';

export type DeviceType = 'mobile' | 'tablet' | 'desktop';

interface DeviceInfo {
  type: DeviceType;
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isTouchDevice: boolean;
  screenWidth: number;
  screenHeight: number;
}

/**
 * Hook to detect device type and capabilities
 * 自动检测设备类型（手机/平板/电脑）
 */
export const useDevice = (): DeviceInfo => {
  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo>(() => getDeviceInfo());

  useEffect(() => {
    const handleResize = () => {
      setDeviceInfo(getDeviceInfo());
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return deviceInfo;
};

function getDeviceInfo(): DeviceInfo {
  const width = window.innerWidth;
  const height = window.innerHeight;

  // 检测是否为触摸设备
  const isTouchDevice =
    'ontouchstart' in window ||
    navigator.maxTouchPoints > 0;

  // 根据屏幕宽度判断设备类型
  // Tailwind breakpoints: sm: 640px, md: 768px, lg: 1024px
  let type: DeviceType;
  if (width < 768) {
    type = 'mobile';
  } else if (width < 1024) {
    type = 'tablet';
  } else {
    type = 'desktop';
  }

  // 额外的移动设备检测（通过User Agent）
  const userAgent = navigator.userAgent.toLowerCase();
  const isMobileUA = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini/i.test(userAgent);

  // 如果User Agent表明是移动设备，即使屏幕较大也认为是移动设备
  if (isMobileUA && type === 'tablet') {
    // iPad等平板设备
    type = 'tablet';
  } else if (isMobileUA && type === 'desktop') {
    // 横屏手机可能被识别为desktop
    type = width < 1024 ? 'tablet' : 'desktop';
  }

  return {
    type,
    isMobile: type === 'mobile',
    isTablet: type === 'tablet',
    isDesktop: type === 'desktop',
    isTouchDevice,
    screenWidth: width,
    screenHeight: height,
  };
}

/**
 * 获取当前视口信息（不使用hook，可在非组件中使用）
 */
export const getViewportInfo = () => {
  return {
    width: window.innerWidth,
    height: window.innerHeight,
    isMobile: window.innerWidth < 768,
    isTablet: window.innerWidth >= 768 && window.innerWidth < 1024,
    isDesktop: window.innerWidth >= 1024,
  };
};
