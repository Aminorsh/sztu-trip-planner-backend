const express = require('express');
const mongoose = require('mongoose');
const cors = require('cors');
const path = require('path');
require('dotenv').config();

const app = express();

// 中间件
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// 静态文件服务 - 用于访问上传的头像
app.use('/uploads', express.static(path.join(__dirname, 'uploads')));

// 数据库连接
mongoose.connect(process.env.MONGODB_URI || 'mongodb://localhost:27017/sztu-trip-planner')
  .then(() => console.log('MongoDB连接成功'))
  .catch(err => {
    console.error('MongoDB连接失败:', err);
    process.exit(1);
  });
//mongoose.connect(process.env.MONGODB_URI, {
  //useNewUrlParser: true,
  //useUnifiedTopology: true,
//})
//.then(() => console.log('MongoDB连接成功'))
//.catch(err => console.error('MongoDB连接错误:', err));

// 路由
const userRoutes = require('./routes/userRoutes');
const tripRoutes = require('./routes/tripRoutes');
const placeRoutes = require('./routes/placeRoutes');
//app.use('/api/users', require('./routes/userRoutes'));
app.use('/api/users', userRoutes);
app.use('/api/trips', tripRoutes);
app.use('/api/places', placeRoutes);

// 健康检查端点
app.get('/api/health', (req, res) => {
  res.json({
    success: true,
    status: 'OK',
    message: '旅行规划后端服务运行正常',
    service: 'SZTU Trip Planner API',
    timestamp: new Date().toISOString()
  });
});

// 404处理
app.use('*', (req, res) => {
  res.status(404).json({
    success: false,
    message: '接口不存在'
  });
});

// 错误处理中间件
app.use((error, req, res, next) => {
  console.error('服务器错误:', error);
  res.status(500).json({
    success: false,
    message: '内部服务器错误',
    ...(process.env.NODE_ENV === 'development' && { error: err.message })
  });
});

// 根路径
app.get('/', (req, res) => {
  res.json({
    message: '欢迎使用 SZTU Trip Planner API',
    version: '1.0.0',
    endpoints: {
      users: '/api/users',
      trips: '/api/trips',
      places: '/api/places'
    }
  });
});

const PORT = process.env.PORT || 5000;

app.listen(PORT, () => {
  console.log(`服务器运行在端口 ${PORT}`);
  console.log(`环境: ${process.env.NODE_ENV}`);
  console.log(`地址: http://localhost:${PORT}`);
});

module.exports = app;