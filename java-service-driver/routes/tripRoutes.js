// routes/tripRoutes.js
const express = require('express');
const router = express.Router();
const tripController = require('../controllers/tripController');
const auth = require('../middleware/auth');

// 所有行程路由都需要认证
router.use(auth);

// 行程详情路由
router.get('/:id', tripController.getTrip);
router.put('/:id', tripController.saveTrip);

// 行程项管理路由
router.post('/:id/items', tripController.addTripItem);
router.put('/:id/items/:itemId', tripController.updateTripItem);
router.delete('/:id/items/:itemId', tripController.deleteTripItem);

// 行程项排序
router.put('/:id/items/reorder', tripController.reorderTripItems);

// 打卡完成
router.post('/:id/items/:itemId/visit', tripController.markAsVisited);

// 行程功能路由
router.post('/:id/optimize-route', tripController.optimizeRoute);
router.post('/:id/share', tripController.shareTrip);
router.post('/:id/export', tripController.exportTrip);

module.exports = router;