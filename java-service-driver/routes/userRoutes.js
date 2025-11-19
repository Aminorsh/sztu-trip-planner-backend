const express = require('express');
const router = express.Router();
const userController = require('../controllers/userController');
const auth = require('../middleware/auth');
const upload = require('../middleware/upload');

// 所有路由都需要认证
router.use(auth);

// 修改密码
router.put('/password', userController.changePassword);

// 更新个人资料（支持头像上传）
router.put('/profile', upload.single('avatar'), userController.updateProfile);

// 获取个人资料
router.get('/profile', userController.getProfile);

// 注销账户
router.delete('/account', userController.deleteAccount);

module.exports = router;