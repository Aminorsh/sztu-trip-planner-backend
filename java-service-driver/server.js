const express = require('express');
const cors = require('cors');
const path = require('path');
require('dotenv').config(); // 加载环境变量

const app = express();
const PORT = process.env.PORT || 5000;

// 中间件
app.use(cors()); // 允许跨域请求
app.use(express.json()); // 解析JSON格式的请求体
app.use(express.urlencoded({ extended: true })); // 解析URL编码的请求体

// 静态文件服务：让外部可以访问 uploads 文件夹下的头像图片
app.use('/uploads', express.static(path.join(__dirname, 'uploads')));

// 导入路由
const userRoutes = require('./routes/userRoutes');
// 使用路由，所有用户相关的API都以 /api/users 开头
app.use('/api/users', userRoutes);

// 一个简单的根路径测试
app.get('/', (req, res) => {
  res.send('SZTU Trip Planner 后端服务正在运行...');
});

// 启动服务器
app.listen(PORT, () => {
  console.log(`🚀 服务器已启动，正在监听 http://localhost:${PORT}`);
});