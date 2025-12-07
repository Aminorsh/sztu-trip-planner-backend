// controllers/tripController.js
const DB = require('../utils/db');
const crypto = require('crypto');

class TripController {
  // 1. 获取行程详情 - GET /api/trips/{id}
  async getTrip(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;

      // 查询行程
      const trip = await DB.queryOne(`
        SELECT t.*, u.username, u.avatar 
        FROM trips t 
        LEFT JOIN users u ON t.user_id = u.id 
        WHERE t.id = ? AND t.deleted_at IS NULL 
        AND (t.is_public = TRUE OR t.user_id = ?)
      `, [id, userId]);

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 查询行程项
      const items = await DB.query(`
        SELECT * FROM trip_items 
        WHERE trip_id = ? AND deleted_at IS NULL 
        ORDER BY sort_order ASC
      `, [id]);

      // 将行程项添加到行程对象中
      trip.items = items;

      res.json({
        success: true,
        data: { trip }
      });
    } catch (error) {
      console.error('获取行程详情错误:', error);
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }

  // 2. 保存行程 - PUT /api/trips/{id}
  async saveTrip(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;
      const { title, description, start_date, end_date, is_public, status } = req.body;

      // 检查行程是否存在且属于当前用户
      const existingTrip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
        [id, userId]
      );

      if (!existingTrip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权修改'
        });
      }

      // 更新行程
      const updateData = {
        title,
        description,
        start_date,
        end_date,
        is_public: is_public || false,
        status: status || 'planning'
      };

      await DB.update('trips', id, updateData);

      // 获取更新后的行程
      const updatedTrip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ?',
        [id]
      );

      res.json({
        success: true,
        message: '行程保存成功',
        data: { trip: updatedTrip }
      });
    } catch (error) {
      console.error('保存行程错误:', error);
      
      if (error.code === 'ER_TRUNCATED_WRONG_VALUE') {
        return res.status(400).json({
          success: false,
          message: '日期格式不正确'
        });
      }
      
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }

  // 3. 新增行程项 - POST /api/trips/{id}/items
  async addTripItem(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;
      const itemData = req.body;

      // 验证行程是否存在且属于当前用户
      const trip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
        [id, userId]
      );

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 获取当前最大排序值
      const result = await DB.queryOne(
        'SELECT MAX(sort_order) as max_order FROM trip_items WHERE trip_id = ?',
        [id]
      );

      const sortOrder = (result?.max_order || 0) + 1;

      // 插入行程项
      const insertData = {
        title: itemData.title,
        description: itemData.description || null,
        location: itemData.location,
        address: itemData.address || null,
        latitude: itemData.latitude || null,
        longitude: itemData.longitude || null,
        start_time: itemData.start_time || null,
        end_time: itemData.end_time || null,
        sort_order: sortOrder,
        category: itemData.category || 'other',
        cost: itemData.cost || 0,
        notes: itemData.notes || null,
        trip_id: id
      };

      const insertResult = await DB.insert('trip_items', insertData);

      // 获取刚插入的行程项
      const tripItem = await DB.queryOne(
        'SELECT * FROM trip_items WHERE id = ?',
        [insertResult.id]
      );

      res.status(201).json({
        success: true,
        message: '行程项添加成功',
        data: { tripItem }
      });
    } catch (error) {
      console.error('新增行程项错误:', error);
      
      if (error.code === 'ER_DATA_TOO_LONG') {
        return res.status(400).json({
          success: false,
          message: '输入数据过长'
        });
      }
      
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }

  // 4. 更新行程项 - PUT /api/trips/{id}/items/{itemId}
  async updateTripItem(req, res) {
    try {
      const { id, itemId } = req.params;
      const userId = req.userId;
      const updateData = req.body;

      // 验证行程是否存在且属于当前用户
      const trip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
        [id, userId]
      );

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 更新行程项
      const result = await DB.update('trip_items', itemId, updateData);

      if (result.affectedRows === 0) {
        return res.status(404).json({
          success: false,
          message: '行程项不存在'
        });
      }

      // 获取更新后的行程项
      const tripItem = await DB.queryOne(
        'SELECT * FROM trip_items WHERE id = ?',
        [itemId]
      );

      res.json({
        success: true,
        message: '行程项更新成功',
        data: { tripItem }
      });
    } catch (error) {
      console.error('更新行程项错误:', error);
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }

  // 5. 删除行程项 - DELETE /api/trips/{id}/items/{itemId}
  async deleteTripItem(req, res) {
    try {
      const { id, itemId } = req.params;
      const userId = req.userId;

      // 验证行程是否存在且属于当前用户
      const trip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
        [id, userId]
      );

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 软删除行程项
      const affectedRows = await DB.softDelete('trip_items', itemId);

      if (affectedRows === 0) {
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
        message: '服务器错误'
      });
    }
  }

  // 6. 行程项排序 - PUT /api/trips/{id}/items/reorder
  async reorderTripItems(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;
      const { itemIds } = req.body;

      if (!Array.isArray(itemIds)) {
        return res.status(400).json({
          success: false,
          message: 'itemIds 必须是数组'
        });
      }

      // 使用事务处理
      await DB.transaction(async (connection) => {
        // 验证行程是否存在且属于当前用户
        const [trips] = await connection.execute(
          'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
          [id, userId]
        );

        if (trips.length === 0) {
          throw new Error('行程不存在或无权访问');
        }

        // 批量更新排序顺序
        for (let i = 0; i < itemIds.length; i++) {
          const itemId = itemIds[i];
          await connection.execute(
            'UPDATE trip_items SET sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND trip_id = ?',
            [i, itemId, id]
          );
        }
      });

      res.json({
        success: true,
        message: '行程项排序更新成功'
      });
    } catch (error) {
      console.error('行程项排序错误:', error);
      
      if (error.message === '行程不存在或无权访问') {
        return res.status(404).json({
          success: false,
          message: error.message
        });
      }
      
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }

  // 7. 打卡完成 - POST /api/trips/{id}/items/{itemId}/visit
  async markAsVisited(req, res) {
    try {
      const { id, itemId } = req.params;
      const userId = req.userId;

      // 验证行程是否存在且属于当前用户
      const trip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
        [id, userId]
      );

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 更新打卡状态
      const now = new Date().toISOString().slice(0, 19).replace('T', ' ');
      const result = await DB.update('trip_items', itemId, {
        visited: true,
        visited_at: now
      });

      if (result.affectedRows === 0) {
        return res.status(404).json({
          success: false,
          message: '行程项不存在'
        });
      }

      // 获取更新后的行程项
      const tripItem = await DB.queryOne(
        'SELECT * FROM trip_items WHERE id = ?',
        [itemId]
      );

      res.json({
        success: true,
        message: '打卡成功',
        data: { tripItem }
      });
    } catch (error) {
      console.error('打卡完成错误:', error);
      res.status(500).json({
        success: false,
        message: '服务器错误'
      });
    }
  }

  // 8. 分享行程 - POST /api/trips/{id}/share
  async shareTrip(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;
      const { expiresIn = 7 } = req.body;

      // 验证行程是否存在且属于当前用户
      const trip = await DB.queryOne(
        'SELECT * FROM trips WHERE id = ? AND user_id = ? AND deleted_at IS NULL',
        [id, userId]
      );

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 生成分享令牌
      const shareToken = crypto.randomBytes(16).toString('hex');
      
      // 计算过期时间
      const expiresAt = new Date();
      expiresAt.setDate(expiresAt.getDate() + parseInt(expiresIn));

      // 更新行程的分享信息
      await DB.update('trips', id, {
        is_public: true,
        share_token: shareToken
      });

      // 生成分享链接
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
        message: '服务器错误'
      });
    }
  }

  // 9. 导出文件 - POST /api/trips/{id}/export
  async exportTrip(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;
      const { format = 'json' } = req.body;

      // 验证行程是否存在且属于当前用户
      const trip = await DB.queryOne(`
        SELECT t.*, u.username, u.email 
        FROM trips t 
        LEFT JOIN users u ON t.user_id = u.id 
        WHERE t.id = ? AND t.user_id = ? AND t.deleted_at IS NULL
      `, [id, userId]);

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 查询行程项
      const items = await DB.query(`
        SELECT * FROM trip_items 
        WHERE trip_id = ? AND deleted_at IS NULL 
        ORDER BY sort_order ASC
      `, [id]);

      let exportData;
      let contentType;
      let filename;

      switch (format) {
        case 'text':
          exportData = this.generateTextExport(trip, items);
          contentType = 'text/plain';
          filename = `trip-${trip.title}-${new Date().toISOString().split('T')[0]}.txt`;
          break;

        case 'csv':
          exportData = this.generateCsvExport(trip, items);
          contentType = 'text/csv';
          filename = `trip-${trip.title}-${new Date().toISOString().split('T')[0]}.csv`;
          break;

        case 'json':
        default:
          exportData = JSON.stringify({
            success: true,
            data: {
              trip,
              items
            }
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
        message: '导出失败'
      });
    }
  }

  // 辅助方法：生成文本格式导出
  generateTextExport(trip, items) {
    let text = `行程: ${trip.title}\n`;
    text += `描述: ${trip.description || '无'}\n`;
    text += `时间: ${new Date(trip.start_date).toLocaleDateString()} - ${new Date(trip.end_date).toLocaleDateString()}\n`;
    text += `状态: ${trip.status}\n`;
    text += `创建者: ${trip.username}\n`;
    text += `创建时间: ${new Date(trip.created_at).toLocaleString()}\n\n`;
    
    text += '行程安排:\n';
    text += '='.repeat(50) + '\n\n';
    
    items.forEach((item, index) => {
      text += `${index + 1}. ${item.title}\n`;
      text += `   地点: ${item.location}\n`;
      if (item.address) {
        text += `   地址: ${item.address}\n`;
      }
      if (item.start_time) {
        text += `   开始时间: ${new Date(item.start_time).toLocaleString()}\n`;
      }
      if (item.end_time) {
        text += `   结束时间: ${new Date(item.end_time).toLocaleString()}\n`;
      }
      if (item.description) {
        text += `   描述: ${item.description}\n`;
      }
      if (item.cost > 0) {
        text += `   费用: ¥${item.cost}\n`;
      }
      text += `   类别: ${item.category}\n`;
      text += `   状态: ${item.visited ? '已完成' : '未完成'}\n`;
      if (item.visited_at) {
        text += `   完成时间: ${new Date(item.visited_at).toLocaleString()}\n`;
      }
      if (item.notes) {
        text += `   备注: ${item.notes}\n`;
      }
      text += '\n';
    });
    
    return text;
  }

  // 辅助方法：生成CSV格式导出
  generateCsvExport(trip, items) {
    let csv = '行程标题,行程描述,开始日期,结束日期,状态,创建者\n';
    csv += `"${trip.title}","${trip.description || ''}","${new Date(trip.start_date).toLocaleDateString()}","${new Date(trip.end_date).toLocaleDateString()}","${trip.status}","${trip.username}"\n\n`;
    
    csv += '项目标题,地点,地址,开始时间,结束时间,描述,费用,类别,状态,完成时间,备注\n';
    
    items.forEach(item => {
      csv += `"${item.title}","${item.location}","${item.address || ''}","${item.start_time ? new Date(item.start_time).toLocaleString() : ''}","${item.end_time ? new Date(item.end_time).toLocaleString() : ''}","${item.description || ''}","${item.cost || 0}","${item.category}","${item.visited ? '已完成' : '未完成'}","${item.visited_at ? new Date(item.visited_at).toLocaleString() : ''}","${item.notes || ''}"\n`;
    });
    
    return csv;
  }
}

module.exports = new TripController();