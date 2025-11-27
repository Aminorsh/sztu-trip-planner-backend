// routes/placeRoutes.js
const express = require('express');
const router = express.Router();
const tripController = require('../controllers/tripController');

// 地点搜索（可以不需要认证）
router.get('/search', tripController.searchPlaces);

module.exports = router;