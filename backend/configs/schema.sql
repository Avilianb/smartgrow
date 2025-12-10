-- 传感器数据表
CREATE TABLE IF NOT EXISTS sensor_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    temperature_c REAL,
    humidity_pct REAL,
    soil_raw INTEGER,
    rain_analog INTEGER,
    rain_digital INTEGER,
    pump_state TEXT,
    shade_state TEXT
);

CREATE INDEX IF NOT EXISTS idx_sensor_timestamp ON sensor_data(device_id, timestamp DESC);

-- 天气预报表
CREATE TABLE IF NOT EXISTS rain_forecast (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL UNIQUE,
    temp_max REAL,
    temp_min REAL,
    precip_mm REAL,
    humidity_pct REAL,
    raw_json TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_forecast_date ON rain_forecast(date);

-- 灌溉计划表
CREATE TABLE IF NOT EXISTS irrigation_plan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    date TEXT NOT NULL,
    planned_volume_l REAL,
    created_at TEXT NOT NULL,
    UNIQUE(device_id, date)
);

CREATE INDEX IF NOT EXISTS idx_plan_device_date ON irrigation_plan(device_id, date);

-- 设备日志表
CREATE TABLE IF NOT EXISTS device_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT,
    extra TEXT
);

CREATE INDEX IF NOT EXISTS idx_log_timestamp ON device_log(device_id, timestamp DESC);

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',  -- 'admin' 或 'user'
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 创建默认管理员账户 (密码: admin123)
INSERT OR IGNORE INTO users (username, password_hash, role, created_at, updated_at)
VALUES ('admin', '$2a$10$A8ZdAbiFcGgdWNO.gTL9tuMYUeIAkAmgl5S7SmdKtFQohLALgGNUG', 'admin', datetime('now'), datetime('now'));

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- 设备表
CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT UNIQUE NOT NULL,
    user_id INTEGER,
    device_name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(device_id);

-- 设备位置表
CREATE TABLE IF NOT EXISTS device_locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT UNIQUE NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    address TEXT,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE
);

-- 设备命令表
CREATE TABLE IF NOT EXISTS device_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    command_type TEXT NOT NULL,  -- 'irrigate', 'update_config', 'toggle_shade'
    parameters TEXT,              -- JSON格式参数，如: {"duration_minutes": 5}
    status TEXT DEFAULT 'pending', -- 'pending', 'executing', 'completed', 'failed'
    created_at TEXT NOT NULL,
    executed_at TEXT,
    result TEXT                   -- 执行结果或错误信息
);

CREATE INDEX IF NOT EXISTS idx_command_status ON device_commands(device_id, status);
