#include <Arduino.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <DHT.h>
#include <ESP32Servo.h>
#include "config.h"

// ============ 全局对象 ============
DHT dht(DHT_PIN, DHTTYPE);
Servo servo1;
Servo servo2;

// ============ 状态变量 ============
bool shadeActive = false;       // 遮阳状态
bool pumpActive = false;        // 水泵状态
unsigned long lastReportTime = 0;

// ============ 传感器数据结构 ============
struct SensorData {
  float temperature;
  float humidity;
  int soilMoisture;
  int rainAnalog;
  int rainDigital;
};

// ============ 函数声明 ============
void setupWiFi();
void setupPins();
SensorData readSensors();
void controlShade(float temperature);
void controlPump(bool isRaining);
String buildJsonPayload(const SensorData& data);
void sendDataToServer(const SensorData& data);
void processServerResponse(String response);

// ============ Setup ============
void setup() {
  Serial.begin(115200);
  delay(1000);

  Serial.println("\n\n=================================");
  Serial.println("SmartGrow 智能灌溉系统");
  Serial.println("ESP32-S3 固件启动中...");
  Serial.println("=================================\n");

  // 初始化引脚
  setupPins();

  // 初始化 DHT 传感器
  dht.begin();
  Serial.println("[传感器] DHT11 已初始化");

  // 初始化舵机
  servo1.attach(SERVO1_PIN);
  servo2.attach(SERVO2_PIN);
  servo1.write(SERVO_OPEN_ANGLE_1);
  servo2.write(SERVO_OPEN_ANGLE_2);
  Serial.println("[执行器] 舵机已初始化");

  // 初始化 WiFi
  setupWiFi();

  Serial.println("\n[系统] 初始化完成，进入主循环\n");
}

// ============ Loop ============
void loop() {
  unsigned long currentTime = millis();

  // 定时上报数据
  if (currentTime - lastReportTime >= REPORT_INTERVAL) {
    lastReportTime = currentTime;

    // 读取传感器
    SensorData data = readSensors();

    // 打印传感器数据
    Serial.println("\n----- 传感器读数 -----");
    Serial.printf("温度: %.1f°C\n", data.temperature);
    Serial.printf("湿度: %.1f%%\n", data.humidity);
    Serial.printf("土壤: %d ADC\n", data.soilMoisture);
    Serial.printf("雨量(模拟): %d\n", data.rainAnalog);
    Serial.printf("雨量(数字): %d (%s)\n", data.rainDigital,
                  data.rainDigital == 0 ? "下雨" : "干燥");
    Serial.println("----------------------");

    // 控制逻辑
    bool isRaining = (data.rainDigital == 0);
    controlShade(data.temperature);
    controlPump(isRaining);

    // 发送数据到服务器
    if (WiFi.status() == WL_CONNECTED) {
      sendDataToServer(data);
    } else {
      Serial.println("[WiFi] 连接断开，尝试重连...");
      WiFi.reconnect();
    }
  }

  delay(100);
}

// ============ WiFi 设置 ============
void setupWiFi() {
  Serial.print("[WiFi] 连接到: ");
  Serial.println(WIFI_SSID);

  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  int attempts = 0;
  while (WiFi.status() != WL_CONNECTED && attempts < 20) {
    delay(500);
    Serial.print(".");
    attempts++;
  }

  if (WiFi.status() == WL_CONNECTED) {
    Serial.println();
    Serial.println("[WiFi] 连接成功!");
    Serial.print("[WiFi] IP 地址: ");
    Serial.println(WiFi.localIP());
  } else {
    Serial.println();
    Serial.println("[WiFi] 连接失败!");
  }
}

// ============ 引脚设置 ============
void setupPins() {
  pinMode(RAIN_DIGITAL_PIN, INPUT_PULLUP);
  pinMode(PUMP_PIN, OUTPUT);
  digitalWrite(PUMP_PIN, LOW);  // 水泵初始关闭

  Serial.println("[引脚] GPIO 配置完成");
}

// ============ 读取传感器 ============
SensorData readSensors() {
  SensorData data;

  // 读取 DHT11
  data.temperature = dht.readTemperature();
  data.humidity = dht.readHumidity();

  // 检查读取是否成功
  if (isnan(data.temperature) || isnan(data.humidity)) {
    Serial.println("[警告] DHT 读取失败，使用默认值");
    data.temperature = 25.0;
    data.humidity = 60.0;
  }

  // 读取土壤湿度 (ADC)
  data.soilMoisture = analogRead(SOIL_PIN);

  // 读取雨滴传感器
  data.rainAnalog = analogRead(RAIN_ANALOG_PIN);
  data.rainDigital = digitalRead(RAIN_DIGITAL_PIN);

  return data;
}

// ============ 控制遮阳 (滞后控制) ============
void controlShade(float temperature) {
  if (temperature >= SHADE_ON_TEMP && !shadeActive) {
    // 开启遮阳
    servo1.write(SERVO_SHADE_ANGLE_1);
    servo2.write(SERVO_SHADE_ANGLE_2);
    shadeActive = true;
    Serial.printf("[遮阳] 开启 (温度: %.1f°C >= %.1f°C)\n", temperature, SHADE_ON_TEMP);
  }
  else if (temperature <= SHADE_OFF_TEMP && shadeActive) {
    // 关闭遮阳
    servo1.write(SERVO_OPEN_ANGLE_1);
    servo2.write(SERVO_OPEN_ANGLE_2);
    shadeActive = false;
    Serial.printf("[遮阳] 关闭 (温度: %.1f°C <= %.1f°C)\n", temperature, SHADE_OFF_TEMP);
  }
}

// ============ 控制水泵 ============
void controlPump(bool isRaining) {
  // 安全规则：下雨时强制关闭水泵
  if (isRaining && pumpActive) {
    digitalWrite(PUMP_PIN, LOW);
    pumpActive = false;
    Serial.println("[水泵] 检测到降雨，强制关闭");
  }
}

// ============ 构建 JSON 数据 ============
String buildJsonPayload(const SensorData& data) {
  JsonDocument doc;

  doc["device_id"] = DEVICE_ID;
  doc["timestamp"] = "2025-12-01T12:00:00Z";  // TODO: 使用 NTP 获取真实时间
  doc["temperature_c"] = data.temperature;
  doc["humidity_pct"] = data.humidity;
  doc["soil_raw"] = data.soilMoisture;
  doc["rain_analog"] = data.rainAnalog;
  doc["rain_digital"] = data.rainDigital;
  doc["pump_state"] = pumpActive ? "on" : "off";
  doc["shade_state"] = shadeActive ? "closed" : "open";

  String payload;
  serializeJson(doc, payload);
  return payload;
}

// ============ 发送数据到服务器 ============
void sendDataToServer(const SensorData& data) {
  HTTPClient http;

  String url = String("http://") + SERVER_HOST + ":" + SERVER_PORT + API_ENDPOINT;
  Serial.print("[HTTP] POST ");
  Serial.println(url);

  http.begin(url);
  http.addHeader("Content-Type", "application/json");

  String payload = buildJsonPayload(data);
  Serial.print("[HTTP] 发送: ");
  Serial.println(payload);

  int httpCode = http.POST(payload);

  if (httpCode > 0) {
    Serial.printf("[HTTP] 响应码: %d\n", httpCode);

    if (httpCode == HTTP_CODE_OK) {
      String response = http.getString();
      Serial.print("[HTTP] 响应: ");
      Serial.println(response);
      processServerResponse(response);
    }
  } else {
    Serial.printf("[HTTP] 请求失败: %s\n", http.errorToString(httpCode).c_str());
  }

  http.end();
}

// ============ 处理服务器响应 ============
void processServerResponse(String response) {
  JsonDocument doc;
  DeserializationError error = deserializeJson(doc, response);

  if (error) {
    Serial.println("[JSON] 解析失败");
    return;
  }

  // 检查是否有命令
  if (!doc["commands"].isNull()) {
    JsonObject commands = doc["commands"];

    // 手动灌溉命令
    if (commands["irrigate_now"] == true) {
      float volume = commands["irrigate_volume_l"];
      Serial.printf("[命令] 收到手动灌溉命令: %.1fL\n", volume);

      // 检查安全条件
      int rainStatus = digitalRead(RAIN_DIGITAL_PIN);
      if (rainStatus == 0) {
        Serial.println("[水泵] 正在下雨，拒绝执行");
      } else {
        digitalWrite(PUMP_PIN, HIGH);
        pumpActive = true;
        Serial.println("[水泵] 开启灌溉");

        // TODO: 根据 volume 计算运行时间
        // 示例：2L = 10秒
        unsigned long runtime = (unsigned long)(volume * 5000);
        delay(runtime);

        digitalWrite(PUMP_PIN, LOW);
        pumpActive = false;
        Serial.println("[水泵] 灌溉完成");
      }
    }

    // 强制遮阳状态
    if (!commands["force_shade_state"].isNull()) {
      const char* shadeCmd = commands["force_shade_state"];
      if (strcmp(shadeCmd, "auto") != 0) {
        Serial.printf("[命令] 收到遮阳命令: %s\n", shadeCmd);
      }
    }
  }
}
