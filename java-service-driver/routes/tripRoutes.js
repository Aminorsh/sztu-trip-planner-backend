// routes/tripRoutes.js
const express = require('express');
const router = express.Router();
const tripController = require('../controllers/tripController');
const auth = require('../middleware/auth');

// 所有行程路由都需要认证
router.use(auth);

// 1. 获取行程详情
router.get('/:id', tripController.getTrip);

// 2. 保存行程
router.put('/:id', tripController.saveTrip);

// 3. 新增行程项
router.post('/:id/items', tripController.addTripItem);

// 4. 更新行程项
router.put('/:id/items/:itemId', tripController.updateTripItem);

// 5. 删除行程项
router.delete('/:id/items/:itemId', tripController.deleteTripItem);

// 6. 行程项排序
router.put('/:id/items/reorder', tripController.reorderTripItems);

// 7. 打卡完成
router.post('/:id/items/:itemId/visit', tripController.markAsVisited);

// 8. 分享行程
router.post('/:id/share', tripController.shareTrip);

// 9. 导出文件
router.post('/:id/export', tripController.exportTrip);

module.exports = router;