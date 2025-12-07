const express = require('express');
const router = express.Router();
const userController = require('../controllers/userController');
const auth = require('../middleware/auth');
const upload = require('../middleware/upload');

// 公开路由：注册和登录（不需要JWT认证）
router.post('/register', userController.register);
router.post('/login', userController.login);

// ====== 以下所有路由都需要JWT认证保护[citation:1] ======
router.use(auth);

// 修改密码 (PUT /api/users/password)
router.put('/password', userController.changePassword);
// 更新个人资料，支持上传单个文件字段名为'avatar' (PUT /api/users/profile)
router.put('/profile', upload.single('avatar'), userController.updateProfile);
// 获取个人资料 (GET /api/users/profile)
router.get('/profile', userController.getProfile);
// 注销账户 (DELETE /api/users/account)
router.delete('/account', userController.deleteAccount);

module.exports = router;