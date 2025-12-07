// server.js
require('dotenv').config();
const express = require('express');
const cors = require('cors');
const path = require('path');

// 导入数据库连接
require('./config/database');

// 导入模型
require('./models/index');

// 创建Express应用
const app = express();

// 中间件
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// 静态文件服务
app.use('/uploads', express.static(path.join(__dirname, 'uploads')));

// 导入路由
const userRoutes = require('./routes/userRoutes');
const tripRoutes = require('./routes/tripRoutes');

// 使用路由
app.use('/api/users', userRoutes);
app.use('/api/trips', tripRoutes);

// 健康检查端点
app.get('/health', (req, res) => {
  res.json({
    status: 'OK',
    timestamp: new Date().toISOString(),
    service: 'SZTU Trip Planner API',
    database: 'MySQL'
  });
});

// 根路径
app.get('/', (req, res) => {
  res.json({
    message: '欢迎使用 SZTU Trip Planner API',
    version: '1.0.0',
    database: 'MySQL',
    endpoints: {
      users: '/api/users',
      trips: '/api/trips',
      health: '/health'
    }
  });
});

// 404处理
app.use('*', (req, res) => {
  res.status(404).json({
    success: false,
    message: '接口不存在',
    path: req.originalUrl
  });
});

// 错误处理中间件
app.use((err, req, res, next) => {
  console.error('服务器错误:', err);
  
  // Sequelize错误处理
  if (err.name && err.name.includes('Sequelize')) {
    return res.status(400).json({
      success: false,
      message: '数据库操作失败',
      ...(process.env.NODE_ENV === 'development' && { error: err.message })
    });
  }
  
  res.status(err.status || 500).json({
    success: false,
    message: '服务器内部错误',
    ...(process.env.NODE_ENV === 'development' && { error: err.message })
  });
});

// 启动服务器
const PORT = process.env.PORT || 3000;

app.listen(PORT, () => {
  console.log(`🚀 服务器运行在端口 ${PORT}`);
  console.log(`📍 环境: ${process.env.NODE_ENV}`);
  console.log(`🗄️  数据库: MySQL`);
  console.log(`🌐 地址: http://localhost:${PORT}`);
});

module.exports = app;