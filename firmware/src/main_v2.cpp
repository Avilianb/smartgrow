#include <Arduino.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <DHT.h>
#include <ESP32Servo.h>
#include "config.h"

// ============ 全局对象 ============
DHT dht(DHT_PIN, DHTTYPE);
Servo servo1;
Servo servo2;
WiFiClientSecure secureClient;

// ============ 状态变量 ============
bool shadeActive = false;       // 遮阳状态
bool pumpActive = false;        // 水泵状态
unsigned long lastReportTime = 0;
unsigned long pumpStartTime = 0;  // 水泵启动时间
float pumpDuration = 0;           // 需要浇水的时长（秒）

// ============ 传感器数据结构 ============
struct SensorData {
  float temperature;
  float humidity;
  int soilMoisture;
  int rainAnalog;
  int rainDigital;
};

// ============ 命令结构 ============
struct Command {
  int64_t id;
  String type;
  float volumeL;
  bool valid;
};

// ============ 函数声明 ============
void setupWiFi();
void setupPins();
SensorData readSensors();
void controlShade(float temperature);
void controlPump(bool isRaining);
String buildJsonPayload(const SensorData& data);
void sendDataToServer(const SensorData& data);
void processCommands(JsonArray commands);
void executeIrrigateCommand(const Command& cmd);
void reportCommandStatus(int64_t cmdId, String status, String result);

// ============ Setup ============
void setup() {
  Serial.begin(115200);
  delay(1000);

  Serial.println("\n\n=================================");
  Serial.println("SmartGrow 智能灌溉系统 v2.0");
  Serial.println("ESP32-S3 固件启动中...");
  Serial.println("支持: HTTPS + 命令队列");
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

  // 配置 HTTPS 客户端（跳过证书验证）
  secureClient.setInsecure();
  Serial.println("[HTTPS] 证书验证已禁用（开发模式）");

  // 初始化 WiFi
  setupWiFi();

  Serial.println("\n[系统] 初始化完成，进入主循环");
  Serial.printf("[配置] 上报间隔: %lu 秒\n", REPORT_INTERVAL / 1000);
  Serial.printf("[配置] 服务器: https://%s\n\n", SERVER_DOMAIN);
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

    // 发送数据到服务器并接收命令
    if (WiFi.status() == WL_CONNECTED) {
      sendDataToServer(data);
    } else {
      Serial.println("[WiFi] 连接断开，尝试重连...");
      WiFi.reconnect();
    }
  }

  // 检查水泵是否需要停止
  if (pumpActive && pumpDuration > 0) {
    if ((millis() - pumpStartTime) / 1000.0 >= pumpDuration) {
      digitalWrite(PUMP_PIN, LOW);
      pumpActive = false;
      Serial.println("[水泵] 灌溉完成，自动停止");
      pumpDuration = 0;
    }
  }

  delay(100);
}

// ============ WiFi 初始化 ============
void setupWiFi() {
  Serial.print("[WiFi] 连接到: ");
  Serial.println(WIFI_SSID);

  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  int attempts = 0;
  while (WiFi.status() != WL_CONNECTED && attempts < 30) {
    delay(500);
    Serial.print(".");
    attempts++;
  }

  if (WiFi.status() == WL_CONNECTED) {
    Serial.println("\n[WiFi] 连接成功!");
    Serial.print("[WiFi] IP 地址: ");
    Serial.println(WiFi.localIP());
    Serial.print("[WiFi] 信号强度: ");
    Serial.print(WiFi.RSSI());
    Serial.println(" dBm");
  } else {
    Serial.println("\n[WiFi] 连接失败！将在主循环中重试");
  }
}

// ============ 引脚初始化 ============
void setupPins() {
  pinMode(RAIN_DIGITAL_PIN, INPUT);
  pinMode(PUMP_PIN, OUTPUT);
  digitalWrite(PUMP_PIN, LOW);  // 水泵初始为关闭
  Serial.println("[引脚] 引脚初始化完成");
}

// ============ 读取传感器 ============
SensorData readSensors() {
  SensorData data;

  // 读取 DHT11 温湿度
  data.temperature = dht.readTemperature();
  data.humidity = dht.readHumidity();

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
    pumpDuration = 0;
    Serial.println("[水泵] 检测到降雨，强制关闭");
  }
}

// ============ 构建 JSON 数据 ============
String buildJsonPayload(const SensorData& data) {
  JsonDocument doc;

  doc["device_id"] = DEVICE_ID;
  doc["timestamp"] = "2025-12-06T12:00:00Z";  // TODO: 使用 NTP 获取真实时间
  doc["temperature_c"] = round(data.temperature * 10) / 10.0;  // 保留1位小数
  doc["humidity_pct"] = round(data.humidity * 10) / 10.0;
  doc["soil_raw"] = data.soilMoisture;
  doc["rain_analog"] = data.rainAnalog;
  doc["rain_digital"] = data.rainDigital;
  doc["pump_state"] = pumpActive ? "on" : "off";
  doc["shade_state"] = shadeActive ? "closed" : "open";

  String payload;
  serializeJson(doc, payload);
  return payload;
}

// ============ 发送数据到服务器 (HTTPS) ============
void sendDataToServer(const SensorData& data) {
  HTTPClient https;

  String url = String("https://") + SERVER_DOMAIN + API_ENDPOINT;
  Serial.print("[HTTPS] POST ");
  Serial.println(url);

  // 使用 WiFiClientSecure
  if (!https.begin(secureClient, url)) {
    Serial.println("[HTTPS] 连接失败");
    return;
  }

  https.addHeader("Content-Type", "application/json");

  String payload = buildJsonPayload(data);
  Serial.print("[HTTPS] 发送: ");
  Serial.println(payload);

  int httpCode = https.POST(payload);

  if (httpCode > 0) {
    Serial.printf("[HTTPS] 响应码: %d\n", httpCode);

    if (httpCode == HTTP_CODE_OK) {
      String response = https.getString();
      Serial.print("[HTTPS] 响应: ");
      Serial.println(response);

      // 解析响应中的命令
      JsonDocument doc;
      DeserializationError error = deserializeJson(doc, response);

      if (!error && doc["success"] == true) {
        if (doc.containsKey("commands")) {
          JsonArray commands = doc["commands"].as<JsonArray>();
          if (commands.size() > 0) {
            Serial.printf("[命令] 收到 %d 条待执行命令\n", commands.size());
            processCommands(commands);
          }
        }
      }
    }
  } else {
    Serial.printf("[HTTPS] 请求失败: %s\n", https.errorToString(httpCode).c_str());
  }

  https.end();
}

// ============ 处理服务器下发的命令 ============
void processCommands(JsonArray commands) {
  for (JsonObject cmd : commands) {
    Command command;
    command.valid = false;

    if (cmd.containsKey("id")) {
      command.id = cmd["id"].as<int64_t>();
      command.type = cmd["command_type"].as<String>();

      Serial.println("\n----- 收到命令 -----");
      Serial.printf("命令ID: %lld\n", command.id);
      Serial.printf("类型: %s\n", command.type.c_str());

      // 解析参数
      if (cmd.containsKey("parameters") && !cmd["parameters"].isNull()) {
        String paramsStr = cmd["parameters"].as<String>();
        JsonDocument paramsDoc;
        if (deserializeJson(paramsDoc, paramsStr) == DeserializationError::Ok) {
          if (command.type == "irrigate") {
            command.volumeL = paramsDoc["volume_l"] | 0.0;
            command.valid = true;
            Serial.printf("参数: volume_l=%.2fL\n", command.volumeL);
          }
        }
      }
      Serial.println("-------------------");

      // 执行命令
      if (command.valid) {
        if (command.type == "irrigate") {
          executeIrrigateCommand(command);
        }
      } else {
        Serial.println("[命令] 参数无效或命令类型不支持");
        reportCommandStatus(command.id, "failed", "Invalid parameters");
      }
    }
  }
}

// ============ 执行灌溉命令 ============
void executeIrrigateCommand(const Command& cmd) {
  Serial.printf("\n[灌溉] 开始执行命令 ID=%lld\n", cmd.id);

  // 上报状态：executing
  reportCommandStatus(cmd.id, "executing", "Starting irrigation");

  // 计算灌溉时长（假设流速 0.5L/秒）
  const float FLOW_RATE = 0.5;  // L/s
  pumpDuration = cmd.volumeL / FLOW_RATE;

  Serial.printf("[灌溉] 目标水量: %.2fL\n", cmd.volumeL);
  Serial.printf("[灌溉] 预计时长: %.1f秒\n", pumpDuration);

  // 启动水泵
  digitalWrite(PUMP_PIN, HIGH);
  pumpActive = true;
  pumpStartTime = millis();

  Serial.println("[水泵] 已启动");

  // 等待灌溉完成
  delay((unsigned long)(pumpDuration * 1000));

  // 停止水泵
  digitalWrite(PUMP_PIN, LOW);
  pumpActive = false;
  pumpDuration = 0;

  Serial.println("[水泵] 已停止");
  Serial.println("[灌溉] 命令执行完成\n");

  // 上报状态：completed
  String result = "Irrigation completed: " + String(cmd.volumeL, 2) + "L";
  reportCommandStatus(cmd.id, "completed", result);
}

// ============ 上报命令执行状态 ============
void reportCommandStatus(int64_t cmdId, String status, String result) {
  HTTPClient https;

  String url = String("https://") + SERVER_DOMAIN + API_CMD_STATUS;
  Serial.printf("[HTTPS] POST %s\n", url.c_str());

  if (!https.begin(secureClient, url)) {
    Serial.println("[HTTPS] 连接失败");
    return;
  }

  https.addHeader("Content-Type", "application/json");

  // 构建请求体
  JsonDocument doc;
  doc["command_id"] = cmdId;
  doc["status"] = status;
  doc["result"] = result;

  String payload;
  serializeJson(doc, payload);

  Serial.print("[HTTPS] 发送: ");
  Serial.println(payload);

  int httpCode = https.POST(payload);

  if (httpCode > 0) {
    Serial.printf("[HTTPS] 响应码: %d\n", httpCode);
    if (httpCode == HTTP_CODE_OK) {
      Serial.println("[命令] 状态上报成功");
    }
  } else {
    Serial.printf("[HTTPS] 请求失败: %s\n", https.errorToString(httpCode).c_str());
  }

  https.end();
}
