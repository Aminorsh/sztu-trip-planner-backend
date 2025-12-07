const jwt = require('jsonwebtoken');
require('dotenv').config();

const auth = async (req, res, next) => {
  try {
    // 从请求头获取Token
    const token = req.header('Authorization')?.replace('Bearer ', '');
    if (!token) {
      return res.status(401).json({ success: false, message: '请提供认证Token' });
    }

    // 验证Token
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    // 将解码出的用户ID（我们之后会存到token里）挂载到req对象上，方便后续使用
    req.userId = decoded.userId;
    next(); // 验证通过，放行到下一个处理函数
  } catch (error) {
    res.status(401).json({ success: false, message: 'Token无效或已过期' });
  }
};

module.exports = auth;