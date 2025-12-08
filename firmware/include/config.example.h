#ifndef CONFIG_H
#define CONFIG_H

// ============ WiFi 配置 ============
#define WIFI_SSID "你的WiFi名称"
#define WIFI_PASSWORD "你的WiFi密码"

// ============ 服务器配置 ============
// 生产环境服务器
#define SERVER_HOST "202.155.123.28"          // 服务器IP地址
#define SERVER_PORT "8080"                    // 服务器端口
#define SERVER_DOMAIN "iot.netr0.com"         // 服务器域名（备用）
#define USE_HTTPS false                       // 暂不使用HTTPS（可通过Nginx启用）

// API端点
#define API_ENDPOINT "/api/device/data"
#define API_CMD_STATUS "/api/device/command/status"

// 设备API密钥（与后端配置一致）
#define DEVICE_API_KEY "你的设备API密钥"

// ============ 引脚定义 ============
// 传感器引脚
#define DHT_PIN 4           // DHT11 温湿度传感器
#define SOIL_PIN 1          // 土壤湿度传感器 (ADC)
#define RAIN_ANALOG_PIN 2   // 雨滴传感器模拟输出 (ADC)
#define RAIN_DIGITAL_PIN 3  // 雨滴传感器数字输出

// 执行器引脚
#define SERVO1_PIN 5        // 舵机1 (遮阳)
#define SERVO2_PIN 6        // 舵机2 (遮阳)
#define PUMP_PIN 7          // 水泵继电器

// ============ DHT 传感器类型 ============
#define DHTTYPE DHT11

// ============ 舵机角度定义 ============
const uint8_t SERVO_OPEN_ANGLE_1 = 0;      // 不遮阳
const uint8_t SERVO_OPEN_ANGLE_2 = 0;
const uint8_t SERVO_SHADE_ANGLE_1 = 135;   // 遮阳
const uint8_t SERVO_SHADE_ANGLE_2 = 110;

// ============ 温度阈值 (滞后控制) ============
const float SHADE_ON_TEMP = 30.0;   // 高于此温度开启遮阳
const float SHADE_OFF_TEMP = 28.0;  // 低于此温度关闭遮阳

// ============ 土壤湿度阈值 ============
const int SOIL_DRY_THRESHOLD = 1500;   // 土壤干燥阈值

// ============ 设备ID ============
#define DEVICE_ID "esp32s3-1"

// ============ 上报间隔 ============
const unsigned long REPORT_INTERVAL = 10000;  // 10秒上报一次

#endif
