// controllers/tripController.js
const { Trip, TripItem, User } = require('../models/index');
const { Op } = require('sequelize');
const crypto = require('crypto');

class TripController {
  // 1. 获取行程详情 - GET /api/trips/{id}
  async getTrip(req, res) {
    try {
      const { id } = req.params;
      const userId = req.userId;

      const trip = await Trip.findByPk(id, {
        include: [
          {
            model: User,
            as: 'user',
            attributes: ['id', 'username', 'avatar']
          },
          {
            model: TripItem,
            as: 'items',
            order: [['sort_order', 'ASC']]
          }
        ]
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在'
        });
      }

      // 检查权限：非公开行程只能由创建者访问
      if (!trip.is_public && trip.user_id !== userId) {
        return res.status(403).json({
          success: false,
          message: '无权访问此行程'
        });
      }

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

      // 查找行程
      let trip = await Trip.findByPk(id);

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在'
        });
      }

      // 检查权限：只能更新自己的行程
      if (trip.user_id !== userId) {
        return res.status(403).json({
          success: false,
          message: '无权修改此行程'
        });
      }

      // 更新行程
      trip = await trip.update({
        title,
        description,
        start_date,
        end_date,
        is_public: is_public || false,
        status: status || 'planning'
      });

      res.json({
        success: true,
        message: '行程保存成功',
        data: { trip }
      });
    } catch (error) {
      console.error('保存行程错误:', error);
      if (error.name === 'SequelizeValidationError') {
        return res.status(400).json({
          success: false,
          message: '数据验证失败',
          errors: error.errors.map(err => err.message)
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
      const trip = await Trip.findOne({
        where: { id, user_id: userId }
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 获取当前最大排序值
      const lastItem = await TripItem.findOne({
        where: { trip_id: id },
        order: [['sort_order', 'DESC']]
      });

      const sortOrder = lastItem ? lastItem.sort_order + 1 : 0;

      // 创建行程项
      const tripItem = await TripItem.create({
        ...itemData,
        sort_order: sortOrder,
        trip_id: id
      });

      res.status(201).json({
        success: true,
        message: '行程项添加成功',
        data: { tripItem }
      });
    } catch (error) {
      console.error('新增行程项错误:', error);
      if (error.name === 'SequelizeValidationError') {
        return res.status(400).json({
          success: false,
          message: '数据验证失败',
          errors: error.errors.map(err => err.message)
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
      const trip = await Trip.findOne({
        where: { id, user_id: userId }
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 查找行程项
      const tripItem = await TripItem.findOne({
        where: { id: itemId, trip_id: id }
      });

      if (!tripItem) {
        return res.status(404).json({
          success: false,
          message: '行程项不存在'
        });
      }

      // 更新行程项
      await tripItem.update(updateData);

      res.json({
        success: true,
        message: '行程项更新成功',
        data: { tripItem }
      });
    } catch (error) {
      console.error('更新行程项错误:', error);
      if (error.name === 'SequelizeValidationError') {
        return res.status(400).json({
          success: false,
          message: '数据验证失败',
          errors: error.errors.map(err => err.message)
        });
      }
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
      const trip = await Trip.findOne({
        where: { id, user_id: userId }
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 删除行程项
      const result = await TripItem.destroy({
        where: { id: itemId, trip_id: id }
      });

      if (result === 0) {
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

      // 验证行程是否存在且属于当前用户
      const trip = await Trip.findOne({
        where: { id, user_id: userId }
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 使用事务更新排序顺序
      const sequelize = require('../config/database');
      const transaction = await sequelize.transaction();

      try {
        for (let i = 0; i < itemIds.length; i++) {
          await TripItem.update(
            { sort_order: i },
            {
              where: { id: itemIds[i], trip_id: id },
              transaction
            }
          );
        }

        await transaction.commit();

        res.json({
          success: true,
          message: '行程项排序更新成功'
        });
      } catch (error) {
        await transaction.rollback();
        throw error;
      }
    } catch (error) {
      console.error('行程项排序错误:', error);
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
      const trip = await Trip.findOne({
        where: { id, user_id: userId }
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 查找行程项
      const tripItem = await TripItem.findOne({
        where: { id: itemId, trip_id: id }
      });

      if (!tripItem) {
        return res.status(404).json({
          success: false,
          message: '行程项不存在'
        });
      }

      // 更新打卡状态
      await tripItem.update({
        visited: true,
        visited_at: new Date()
      });

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
      const { expiresIn = 7 } = req.body; // 默认7天

      // 验证行程是否存在且属于当前用户
      const trip = await Trip.findOne({
        where: { id, user_id: userId }
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      // 生成唯一的分享令牌
      const shareToken = crypto.randomBytes(16).toString('hex');
      
      // 计算过期时间
      const expiresAt = new Date();
      expiresAt.setDate(expiresAt.getDate() + parseInt(expiresIn));

      // 更新行程的分享信息
      await trip.update({
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
      const trip = await Trip.findOne({
        where: { id, user_id: userId },
        include: [
          {
            model: User,
            as: 'user',
            attributes: ['username', 'email']
          },
          {
            model: TripItem,
            as: 'items',
            order: [['sort_order', 'ASC']]
          }
        ]
      });

      if (!trip) {
        return res.status(404).json({
          success: false,
          message: '行程不存在或无权访问'
        });
      }

      let exportData;
      let contentType;
      let filename;

      switch (format) {
        case 'text':
          // 文本格式导出
          exportData = this.generateTextExport(trip);
          contentType = 'text/plain';
          filename = `trip-${trip.title}-${new Date().toISOString().split('T')[0]}.txt`;
          break;

        case 'csv':
          // CSV格式导出
          exportData = this.generateCsvExport(trip);
          contentType = 'text/csv';
          filename = `trip-${trip.title}-${new Date().toISOString().split('T')[0]}.csv`;
          break;

        case 'json':
        default:
          // JSON格式导出
          exportData = JSON.stringify({
            success: true,
            data: {
              trip: {
                id: trip.id,
                title: trip.title,
                description: trip.description,
                start_date: trip.start_date,
                end_date: trip.end_date,
                status: trip.status,
                created_at: trip.created_at,
                user: trip.user,
                items: trip.items.map(item => ({
                  id: item.id,
                  title: item.title,
                  description: item.description,
                  location: item.location,
                  address: item.address,
                  start_time: item.start_time,
                  end_time: item.end_time,
                  visited: item.visited,
                  visited_at: item.visited_at,
                  category: item.category,
                  cost: item.cost,
                  notes: item.notes,
                  sort_order: item.sort_order
                }))
              }
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
  generateTextExport(trip) {
    let text = `行程: ${trip.title}\n`;
    text += `描述: ${trip.description || '无'}\n`;
    text += `时间: ${new Date(trip.start_date).toLocaleDateString()} - ${new Date(trip.end_date).toLocaleDateString()}\n`;
    text += `状态: ${trip.status}\n`;
    text += `创建者: ${trip.user.username}\n`;
    text += `创建时间: ${new Date(trip.created_at).toLocaleString()}\n\n`;
    
    text += '行程安排:\n';
    text += '='.repeat(50) + '\n\n';
    
    trip.items.forEach((item, index) => {
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
  generateCsvExport(trip) {
    let csv = '行程标题,行程描述,开始日期,结束日期,状态,创建者\n';
    csv += `"${trip.title}","${trip.description || ''}","${new Date(trip.start_date).toLocaleDateString()}","${new Date(trip.end_date).toLocaleDateString()}","${trip.status}","${trip.user.username}"\n\n`;
    
    csv += '项目标题,地点,地址,开始时间,结束时间,描述,费用,类别,状态,完成时间,备注\n';
    
    trip.items.forEach(item => {
      csv += `"${item.title}","${item.location}","${item.address || ''}","${item.start_time ? new Date(item.start_time).toLocaleString() : ''}","${item.end_time ? new Date(item.end_time).toLocaleString() : ''}","${item.description || ''}","${item.cost || 0}","${item.category}","${item.visited ? '已完成' : '未完成'}","${item.visited_at ? new Date(item.visited_at).toLocaleString() : ''}","${item.notes || ''}"\n`;
    });
    
    return csv;
  }
}

module.exports = new TripController();