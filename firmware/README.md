# ESP32 固件配置指南

## 📋 概述

本固件适用于 ESP32-S3 开发板，用于 SmartGrow 智能灌溉系统的设备端。

## 🔧 硬件要求

### 开发板
- ESP32-S3-DevKitC-1

### 传感器
- DHT11 温湿度传感器
- 土壤湿度传感器（电容式）
- 雨滴传感器（模拟+数字输出）

### 执行器
- 2个舵机（用于遮阳装置）
- 1个继电器模块（用于控制水泵）

## 📌 引脚连接

| 功能 | 引脚 | 说明 |
|------|------|------|
| DHT11 | GPIO 4 | 温湿度传感器 |
| 土壤湿度 | GPIO 1 | ADC读取 |
| 雨滴（模拟） | GPIO 2 | ADC读取 |
| 雨滴（数字） | GPIO 3 | 数字输入（下拉） |
| 舵机1 | GPIO 5 | 遮阳装置 |
| 舵机2 | GPIO 6 | 遮阳装置 |
| 水泵继电器 | GPIO 7 | 数字输出 |

## ⚙️ 配置说明

### 1. WiFi 配置

编辑 `include/config.h` 文件：

```cpp
#define WIFI_SSID "你的WiFi名称"
#define WIFI_PASSWORD "你的WiFi密码"
```

### 2. 服务器配置

**生产环境配置（当前）：**
```cpp
#define SERVER_HOST "202.155.123.28"
#define SERVER_PORT "8080"
#define USE_HTTPS false
```

**如果使用域名：**
```cpp
#define SERVER_HOST "iot.netr0.com"
#define SERVER_PORT "443"
#define USE_HTTPS true
```

### 3. 设备API密钥

⚠️ **重要**：此密钥必须与后端配置一致！

```cpp
#define DEVICE_API_KEY "8dc77f5ea8df913a7a99027bc1975011c9b92cc8a67d67784652e3d3e3830b84"
```

**在后端查看密钥：**
```bash
cat backend/configs/config.yaml | grep device_api_key
```

### 4. 设备ID

```cpp
#define DEVICE_ID "esp32s3-1"
```

如果有多个设备，修改为唯一ID（如 `esp32s3-2`, `esp32s3-3`）

## 🚀 编译和上传

### 使用 PlatformIO

```bash
# 进入固件目录
cd firmware

# 编译
pio run

# 上传到开发板
pio run --target upload

# 查看串口输出
pio device monitor
```

### 使用 Arduino IDE

1. 打开 `src/main.cpp`
2. 安装依赖库：
   - DHT sensor library
   - Adafruit Unified Sensor
   - ArduinoJson (v7.x)
   - ESP32Servo
3. 选择开发板：ESP32-S3-DevKitC-1
4. 点击上传

## 📊 串口输出示例

```
=================================
SmartGrow 智能灌溉系统
ESP32-S3 固件启动中...
=================================

[引脚] GPIO 配置完成
[传感器] DHT11 已初始化
[执行器] 舵机已初始化
[WiFi] 连接到: test
[WiFi] 连接成功!
[WiFi] IP 地址: 192.168.1.100

[系统] 初始化完成，进入主循环

----- 传感器读数 -----
温度: 25.3°C
湿度: 62.5%
土壤: 2150 ADC
雨量(模拟): 1023
雨量(数字): 1 (干燥)
----------------------
[HTTP] POST http://202.155.123.28:8080/api/device/data
[HTTP] 发送: {"device_id":"esp32s3-1",...}
[HTTP] 响应码: 200
[HTTP] 响应: {"success":true,"message":"Data received"}
```

## 🔍 故障排查

### WiFi连接失败
- 检查WiFi SSID和密码是否正确
- 确保WiFi信号强度足够
- 检查路由器是否支持2.4GHz

### HTTP请求失败
- 检查服务器IP地址和端口是否正确
- 检查服务器防火墙是否开放8080端口
- 使用 `curl` 测试服务器是否可访问：
  ```bash
  curl http://202.155.123.28:8080/health
  ```

### 认证失败（401/403）
- 检查DEVICE_API_KEY是否与后端配置一致
- 确保HTTP请求头包含 `X-Device-API-Key`

### 传感器读取失败
- 检查传感器连接是否正确
- 使用万用表测试传感器电源
- 查看串口输出的错误信息

## 📝 开发说明

### 添加新传感器

1. 在 `config.h` 中定义引脚：
```cpp
#define NEW_SENSOR_PIN 8
```

2. 在 `SensorData` 结构体中添加字段：
```cpp
struct SensorData {
  // ... 现有字段
  int newSensorValue;
};
```

3. 在 `readSensors()` 函数中读取：
```cpp
data.newSensorValue = analogRead(NEW_SENSOR_PIN);
```

4. 在 `buildJsonPayload()` 中添加到JSON：
```cpp
doc["new_sensor"] = data.newSensorValue;
```

### 修改上报间隔

编辑 `config.h`：
```cpp
const unsigned long REPORT_INTERVAL = 30000;  // 30秒上报一次
```

### 修改温度阈值

编辑 `config.h`：
```cpp
const float SHADE_ON_TEMP = 35.0;   // 高于35°C开启遮阳
const float SHADE_OFF_TEMP = 32.0;  // 低于32°C关闭遮阳
```

## 🔗 相关链接

- **后端配置文档**: `backend/configs/config.yaml`
- **API文档**: 查看 `README.md` 中的API部分
- **服务器端代码**: `backend/internal/handler/handler.go`

## ⚠️ 安全注意事项

1. **不要将包含真实WiFi密码和API密钥的config.h提交到公开仓库**
2. 建议创建 `config.example.h` 作为模板，将真实配置保存在 `config.h` 中
3. 在 `.gitignore` 中添加：
   ```
   firmware/include/config.h
   ```

## 📞 技术支持

如遇问题：
1. 查看串口输出日志
2. 检查服务器端日志：`tail -f /root/smart-grow/logs/server.log`
3. 提交Issue到GitHub仓库
