// controllers/tripController.js
const Trip = require('../models/Trip');
const TripItem = require('../models/TripItem');
const mongoose = require('mongoose');

// 1. 获取行程详情 - GET /api/trips/{id}
exports.getTrip = async (req, res) => {
  try {
    const { id } = req.params;

    // 验证ID格式
    if (!mongoose.Types.ObjectId.isValid(id)) {
      return res.status(400).json({
        success: false,
        message: '行程ID格式错误'
      });
    }

    const trip = await Trip.findById(id);
    
    if (!trip) {
      return res.status(404).json({
        success: false,
        message: '行程不存在'
      });
    }

    // 获取行程项并按排序顺序排列
    const items = await TripItem.find({ tripId: id }).sort({ sortOrder: 1 });

    res.json({
      success: true,
      data: {
        trip: {
          ...trip.toObject(),
          items
        }
      }
    });
  } catch (error) {
    console.error('获取行程详情错误:', error);
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};

// 2. 保存行程 - PUT /api/trips/{id}
exports.saveTrip = async (req, res) => {
  try {
    const { id } = req.params;
    const { title, description, startDate, endDate, isPublic, status } = req.body;

    // 验证ID格式
    if (!mongoose.Types.ObjectId.isValid(id)) {
      return res.status(400).json({
        success: false,
        message: '行程ID格式错误'
      });
    }

    const trip = await Trip.findByIdAndUpdate(
      id,
      {
        title,
        description,
        startDate,
        endDate,
        isPublic: isPublic || false,
        status: status || 'planning'
      },
      { 
        new: true, 
        runValidators: true 
      }
    );

    if (!trip) {
      return res.status(404).json({
        success: false,
        message: '行程不存在'
      });
    }

    res.json({
      success: true,
      message: '行程保存成功',
      data: { trip }
    });
  } catch (error) {
    console.error('保存行程错误:', error);
    if (error.name === 'ValidationError') {
      return res.status(400).json({
        success: false,
        message: '输入数据验证失败',
        errors: Object.values(error.errors).map(err => err.message)
      });
    }
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};

// 3. 新增行程项 - POST /api/trips/{id}/items
exports.addTripItem = async (req, res) => {
  try {
    const { id } = req.params;
    const { 
      title, 
      description, 
      location, 
      address, 
      latitude, 
      longitude, 
      startTime, 
      endTime, 
      category,
      cost,
      notes
    } = req.body;

    // 验证行程是否存在
    const trip = await Trip.findById(id);
    if (!trip) {
      return res.status(404).json({
        success: false,
        message: '行程不存在'
      });
    }

    // 获取当前最大排序值
    const lastItem = await TripItem.findOne({ tripId: id })
      .sort({ sortOrder: -1 });
    const sortOrder = lastItem ? lastItem.sortOrder + 1 : 0;

    const tripItem = new TripItem({
      title,
      description,
      location,
      address,
      latitude,
      longitude,
      startTime,
      endTime,
      sortOrder,
      category: category || 'other',
      cost: cost || 0,
      notes: notes || '',
      tripId: id
    });

    await tripItem.save();

    res.status(201).json({
      success: true,
      message: '行程项添加成功',
      data: { tripItem }
    });
  } catch (error) {
    console.error('新增行程项错误:', error);
    if (error.name === 'ValidationError') {
      return res.status(400).json({
        success: false,
        message: '输入数据验证失败',
        errors: Object.values(error.errors).map(err => err.message)
      });
    }
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};

// 4. 更新行程项 - PUT /api/trips/{id}/items/{itemId}
exports.updateTripItem = async (req, res) => {
  try {
    const { id, itemId } = req.params;
    const updateData = req.body;

    // 移除不允许更新的字段
    delete updateData.tripId;
    delete updateData._id;

    const tripItem = await TripItem.findOneAndUpdate(
      { _id: itemId, tripId: id },
      updateData,
      { 
        new: true, 
        runValidators: true 
      }
    );

    if (!tripItem) {
      return res.status(404).json({
        success: false,
        message: '行程项不存在'
      });
    }

    res.json({
      success: true,
      message: '行程项更新成功',
      data: { tripItem }
    });
  } catch (error) {
    console.error('更新行程项错误:', error);
    if (error.name === 'ValidationError') {
      return res.status(400).json({
        success: false,
        message: '输入数据验证失败',
        errors: Object.values(error.errors).map(err => err.message)
      });
    }
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};

// 5. 删除行程项 - DELETE /api/trips/{id}/items/{itemId}
exports.deleteTripItem = async (req, res) => {
  try {
    const { id, itemId } = req.params;

    const tripItem = await TripItem.findOneAndDelete({
      _id: itemId,
      tripId: id
    });

    if (!tripItem) {
      return res.status(404).json({
        success: false,
        message: '行程项不存在'
      });
    }

    res.json({
      success: true,
      message: '行程项删除成功'
    });
  } catch (error) {
    console.error('删除行程项错误:', error);
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};

// 6. 行程项排序 - PUT /api/trips/{id}/items/reorder
exports.reorderTripItems = async (req, res) => {
  try {
    const { id } = req.params;
    const { itemIds } = req.body;

    if (!Array.isArray(itemIds)) {
      return res.status(400).json({
        success: false,
        message: 'itemIds 必须是数组'
      });
    }

    // 验证行程是否存在
    const trip = await Trip.findById(id);
    if (!trip) {
      return res.status(404).json({
        success: false,
        message: '行程不存在'
      });
    }

    // 批量更新排序顺序
    const bulkOperations = itemIds.map((itemId, index) => ({
      updateOne: {
        filter: { _id: itemId, tripId: id },
        update: { sortOrder: index }
      }
    }));

    await TripItem.bulkWrite(bulkOperations);

    res.json({
      success: true,
      message: '行程项排序更新成功'
    });
  } catch (error) {
    console.error('行程项排序错误:', error);
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};

// 7. 打卡完成 - POST /api/trips/{id}/items/{itemId}/visit
exports.markAsVisited = async (req, res) => {
  try {
    const { id, itemId } = req.params;

    const tripItem = await TripItem.findOneAndUpdate(
      { _id: itemId, tripId: id },
      { 
        visited: true,
        visitedAt: new Date()
      },
      { new: true }
    );

    if (!tripItem) {
      return res.status(404).json({
        success: false,
        message: '行程项不存在'
      });
    }

    res.json({
      success: true,
      message: '打卡成功',
      data: { tripItem }
    });
  } catch (error) {
    console.error('打卡完成错误:', error);
    res.status(500).json({
      success: false,
      message: '服务器错误，请稍后重试'
    });
  }
};


// 10. 分享行程 - POST /api/trips/{id}/share
exports.shareTrip = async (req, res) => {
  try {
    const { id } = req.params;
    const { expiresIn = '7' } = req.body; // 默认7天有效期

    const trip = await Trip.findById(id);
    if (!trip) {
      return res.status(404).json({
        success: false,
        message: '行程不存在'
      });
    }

    // 生成分享令牌
    const crypto = require('crypto');
    const shareToken = crypto.randomBytes(32).toString('hex');
    
    // 设置分享过期时间
    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + parseInt(expiresIn));

    // 更新行程的分享信息
    trip.isPublic = true;
    trip.shareToken = shareToken;
    await trip.save();

    const shareUrl = `${req.protocol}://${req.get('host')}/shared/trips/${shareToken}`;

    res.json({
      success: true,
      message: '行程分享成功',
      data: {
        shareUrl,
        shareToken,
        expiresAt: expiresAt.toISOString(),
        expiresIn: `${expiresIn}天`
      }
    });
  } catch (error) {
    console.error('分享行程错误:', error);
    res.status(500).json({
      success: false,
      message: '分享失败，请稍后重试'
    });
  }
};

// 11. 导出文件 - POST /api/trips/{id}/export
exports.exportTrip = async (req, res) => {
  try {
    const { id } = req.params;
    const { format = 'json' } = req.body; // json, text, pdf

    const trip = await Trip.findById(id);
    if (!trip) {
      return res.status(404).json({
        success: false,
        message: '行程不存在'
      });
    }

    const items = await TripItem.find({ tripId: id }).sort({ sortOrder: 1 });

    let exportData;
    let contentType;
    let filename;

    switch (format) {
      case 'text':
        // 生成文本格式
        let textContent = `行程: ${trip.title}\n`;
        textContent += `描述: ${trip.description || '无'}\n`;
        textContent += `时间: ${trip.startDate.toDateString()} - ${trip.endDate.toDateString()}\n\n`;
        textContent += '行程安排:\n';
        
        items.forEach((item, index) => {
          textContent += `${index + 1}. ${item.title}\n`;
          textContent += `   地点: ${item.location}\n`;
          if (item.startTime) {
            textContent += `   时间: ${item.startTime.toLocaleString()}\n`;
          }
          if (item.description) {
            textContent += `   描述: ${item.description}\n`;
          }
          textContent += `   状态: ${item.visited ? '已完成' : '未完成'}\n\n`;
        });

        exportData = textContent;
        contentType = 'text/plain';
        filename = `trip-${trip.title}-${new Date().toISOString().split('T')[0]}.txt`;
        break;

      case 'pdf':
        // 这里可以集成PDF生成库，如pdfkit
        // 暂时返回JSON并提示PDF功能开发中
        return res.json({
          success: true,
          message: 'PDF导出功能开发中，当前返回JSON格式',
          data: {
            trip: trip.toObject(),
            items: items.map(item => item.toObject())
          }
        });

      case 'json':
      default:
        exportData = JSON.stringify({
          trip: trip.toObject(),
          items: items.map(item => item.toObject())
        }, null, 2);
        contentType = 'application/json';
        filename = `trip-${trip.title}-${new Date().toISOString().split('T')[0]}.json`;
        break;
    }

    // 设置响应头，触发文件下载
    res.setHeader('Content-Type', contentType);
    res.setHeader('Content-Disposition', `attachment; filename="${filename}"`);
    
    res.send(exportData);
  } catch (error) {
    console.error('导出文件错误:', error);
    res.status(500).json({
      success: false,
      message: '导出失败，请稍后重试'
    });
  }
};