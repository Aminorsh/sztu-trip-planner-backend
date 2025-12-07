const pool = require('../config/database');
const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const fs = require('fs').promises;
const path = require('path');

// 辅助函数：生成JWT Token[citation:1]
const generateToken = (userId) => {
  return jwt.sign({ userId }, process.env.JWT_SECRET, { expiresIn: '7d' }); // Token有效期7天
};

const userController = {
  // 1. 修改密码
  changePassword: async (req, res) => {
    const { newPassword } = req.body; // 通过认证中间件后，req.userId是当前登录用户
    const userId = req.userId;
    if (!newPassword) {
      return res.status(400).json({ success: false, message: '请提供新密码' });
    }
    try {
      const hashedPassword = await bcrypt.hash(newPassword, 10); // 加密新密码
      const [result] = await pool.execute(
        'UPDATE users SET password = ? WHERE id = ?',
        [hashedPassword, userId]
      );
      if (result.affectedRows === 0) {
        return res.status(404).json({ success: false, message: '用户不存在' });
      }
      res.json({ success: true, message: '密码修改成功' });
    } catch (error) {
      res.status(500).json({ success: false, message: '服务器错误' });
    }
  },

  // 2. 更新个人资料 (头像和用户名)
  updateProfile: async (req, res) => {
    const userId = req.userId;
    const { username } = req.body;
    let avatarUrl = null;

    // 处理上传的头像
    if (req.file) {
      avatarUrl = `/uploads/${req.file.filename}`; // 头像访问路径
    }

    try {
      // 构建更新字段
      let updateFields = [];
      let queryParams = [];
      if (username) {
        updateFields.push('username = ?');
        queryParams.push(username);
      }
      if (avatarUrl) {
        updateFields.push('avatar_url = ?');
        queryParams.push(avatarUrl);
      }
      if (updateFields.length === 0) {
        return res.status(400).json({ success: false, message: '无有效更新内容' });
      }
      queryParams.push(userId); // WHERE 条件
      const query = `UPDATE users SET ${updateFields.join(', ')} WHERE id = ?`;
      await pool.execute(query, queryParams);
      res.json({ success: true, message: '资料更新成功', data: { username, avatarUrl } });
    } catch (error) {
      if (error.code === 'ER_DUP_ENTRY') {
        return res.status(400).json({ success: false, message: '用户名已存在' });
      }
      res.status(500).json({ success: false, message: '服务器错误' });
    }
  },

  // 3. 获取个人资料
  getProfile: async (req, res) => {
    const userId = req.userId;
    try {
      const [rows] = await pool.execute(
        'SELECT id, username, email, avatar_url FROM users WHERE id = ?',
        [userId]
      );
      if (rows.length === 0) {
        return res.status(404).json({ success: false, message: '用户不存在' });
      }
      const user = rows[0];
      res.json({
        success: true,
        data: {
          id: user.id,
          username: user.username,
          avatar: user.avatar_url // 直接返回数据库存储的路径
        }
      });
    } catch (error) {
      res.status(500).json({ success: false, message: '服务器错误' });
    }
  },

  // 4. 注销账户
  deleteAccount: async (req, res) => {
    const userId = req.userId;
    const connection = await pool.getConnection(); // 获取一个连接，用于事务
    try {
      await connection.beginTransaction(); // 开始事务
      // 先查询用户信息，获取头像路径以便删除文件
      const [userRows] = await connection.execute('SELECT avatar_url FROM users WHERE id = ?', [userId]);
      if (userRows.length === 0) {
        await connection.rollback();
        return res.status(404).json({ success: false, message: '用户不存在' });
      }
      const avatarUrl = userRows[0].avatar_url;
      // 删除数据库中的用户记录
      await connection.execute('DELETE FROM users WHERE id = ?', [userId]);
      // 如果用户有头像，删除本地头像文件
      if (avatarUrl) {
        const avatarPath = path.join(__dirname, '..', avatarUrl);
        try {
          await fs.unlink(avatarPath);
        } catch (fsError) {
          console.warn('删除头像文件失败，可能文件不存在:', fsError);
        }
      }
      await connection.commit(); // 提交事务
      res.json({ success: true, message: '账户已成功注销' });
    } catch (error) {
      await connection.rollback(); // 回滚事务
      res.status(500).json({ success: false, message: '服务器错误' });
    } finally {
      connection.release(); // 释放连接回连接池
    }
  },

  // （额外补充）用户注册 - 这是测试所有功能的前提
  register: async (req, res) => {
    const { username, email, password } = req.body;
    if (!username || !email || !password) {
      return res.status(400).json({ success: false, message: '请填写所有必填字段' });
    }
    try {
      const hashedPassword = await bcrypt.hash(password, 10);
      const [result] = await pool.execute(
        'INSERT INTO users (username, email, password) VALUES (?, ?, ?)',
        [username, email, hashedPassword]
      );
      const token = generateToken(result.insertId); // 注册后直接登录，生成Token
      res.status(201).json({
        success: true,
        message: '注册成功',
        data: { token, userId: result.insertId, username }
      });
    } catch (error) {
      if (error.code === 'ER_DUP_ENTRY') {
        return res.status(400).json({ success: false, message: '用户名或邮箱已存在' });
      }
      res.status(500).json({ success: false, message: '服务器错误' });
    }
  },

  // （额外补充）用户登录 - 这是获取Token的唯一途径
  login: async (req, res) => {
    const { email, password } = req.body;
    try {
      const [rows] = await pool.execute('SELECT * FROM users WHERE email = ?', [email]);
      if (rows.length === 0) {
        return res.status(401).json({ success: false, message: '邮箱或密码错误' });
      }
      const user = rows[0];
      const isPasswordValid = await bcrypt.compare(password, user.password);
      if (!isPasswordValid) {
        return res.status(401).json({ success: false, message: '邮箱或密码错误' });
      }
      const token = generateToken(user.id);
      res.json({
        success: true,
        message: '登录成功',
        data: { token, userId: user.id, username: user.username }
      });
    } catch (error) {
      res.status(500).json({ success: false, message: '服务器错误' });
    }
  }
};

module.exports = userController;