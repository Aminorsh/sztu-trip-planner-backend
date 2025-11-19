const User = require('../models/User');
const jwt = require('jsonwebtoken');
const fs = require('fs');
const path = require('path');

// 生成JWT Token
const generateToken = (userId) => {
  return jwt.sign({ id: userId }, process.env.JWT_SECRET, { expiresIn: '30d' });
};

const userController = {
  // 修改密码
  changePassword: async (req, res) => {
    try {
      const { currentPassword, newPassword } = req.body;
      const userId = req.user._id;

      if (!currentPassword || !newPassword) {
        return res.status(400).json({
          success: false,
          message: '请提供当前密码和新密码'
        });
      }

      if (newPassword.length < 6) {
        return res.status(400).json({
          success: false,
          message: '新密码长度至少6位'
        });
      }

      const user = await User.findById(userId);
      
      // 验证当前密码
      const isMatch = await user.comparePassword(currentPassword);
      if (!isMatch) {
        return res.status(400).json({
          success: false,
          message: '当前密码不正确'
        });
      }

      // 更新密码
      user.password = newPassword;
      await user.save();

      res.json({
        success: true,
        message: '密码修改成功'
      });

    } catch (error) {
      console.error('修改密码错误:', error);
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  },

  // 更新个人资料
  updateProfile: async (req, res) => {
    try {
      const { username } = req.body;
      const userId = req.user._id;

      const updateData = {};
      if (username) updateData.username = username;
      
      // 如果有上传头像
      if (req.file) {
        // 删除旧头像文件
        const user = await User.findById(userId);
        if (user.avatar) {
          const oldAvatarPath = path.join(__dirname, '..', user.avatar);
          if (fs.existsSync(oldAvatarPath)) {
            fs.unlinkSync(oldAvatarPath);
          }
        }
        updateData.avatar = req.file.path;
      }

      const updatedUser = await User.findByIdAndUpdate(
        userId,
        updateData,
        { new: true, runValidators: true }
      ).select('-password');

      res.json({
        success: true,
        message: '个人资料更新成功',
        data: {
          id: updatedUser._id,
          username: updatedUser.username,
          avatar: updatedUser.avatar,
          email: updatedUser.email
        }
      });

    } catch (error) {
      console.error('更新个人资料错误:', error);
      if (error.code === 11000) {
        return res.status(400).json({
          success: false,
          message: '用户名已存在'
        });
      }
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  },

  // 获取个人资料
  getProfile: async (req, res) => {
    try {
      const user = await User.findById(req.user._id).select('-password');
      
      res.json({
        success: true,
        data: {
          id: user._id,
          username: user.username,
          avatar: user.avatar,
          email: user.email,
          createdAt: user.createdAt
        }
      });

    } catch (error) {
      console.error('获取个人资料错误:', error);
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  },

  // 注销账户
  deleteAccount: async (req, res) => {
    try {
      const userId = req.user._id;
      
      // 删除用户头像文件
      const user = await User.findById(userId);
      if (user.avatar) {
        const avatarPath = path.join(__dirname, '..', user.avatar);
        if (fs.existsSync(avatarPath)) {
          fs.unlinkSync(avatarPath);
        }
      }

      await User.findByIdAndDelete(userId);

      res.json({
        success: true,
        message: '账户已成功注销'
      });

    } catch (error) {
      console.error('注销账户错误:', error);
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }
};

module.exports = userController;