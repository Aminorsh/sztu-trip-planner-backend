// migrations/create-trips-tables.js
'use strict';

module.exports = {
  up: async (queryInterface, Sequelize) => {
    // 创建trips表
    await queryInterface.createTable('trips', {
      id: {
        type: Sequelize.INTEGER,
        primaryKey: true,
        autoIncrement: true,
        allowNull: false
      },
      title: {
        type: Sequelize.STRING(100),
        allowNull: false
      },
      description: {
        type: Sequelize.TEXT,
        allowNull: true
      },
      start_date: {
        type: Sequelize.DATE,
        allowNull: false
      },
      end_date: {
        type: Sequelize.DATE,
        allowNull: false
      },
      is_public: {
        type: Sequelize.BOOLEAN,
        defaultValue: false
      },
      share_token: {
        type: Sequelize.STRING(100),
        unique: true,
        allowNull: true
      },
      cover_image: {
        type: Sequelize.STRING(255),
        defaultValue: ''
      },
      status: {
        type: Sequelize.ENUM('planning', 'ongoing', 'completed', 'cancelled'),
        defaultValue: 'planning'
      },
      user_id: {
        type: Sequelize.INTEGER,
        allowNull: false,
        references: {
          model: 'users',
          key: 'id'
        },
        onUpdate: 'CASCADE',
        onDelete: 'CASCADE'
      },
      created_at: {
        type: Sequelize.DATE,
        allowNull: false,
        defaultValue: Sequelize.literal('CURRENT_TIMESTAMP')
      },
      updated_at: {
        type: Sequelize.DATE,
        allowNull: false,
        defaultValue: Sequelize.literal('CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP')
      },
      deleted_at: {
        type: Sequelize.DATE,
        allowNull: true
      }
    });

    // 创建trip_items表
    await queryInterface.createTable('trip_items', {
      id: {
        type: Sequelize.INTEGER,
        primaryKey: true,
        autoIncrement: true,
        allowNull: false
      },
      title: {
        type: Sequelize.STRING(100),
        allowNull: false
      },
      description: {
        type: Sequelize.TEXT,
        allowNull: true
      },
      location: {
        type: Sequelize.STRING(100),
        allowNull: false
      },
      address: {
        type: Sequelize.STRING(255),
        allowNull: true
      },
      latitude: {
        type: Sequelize.FLOAT,
        allowNull: true
      },
      longitude: {
        type: Sequelize.FLOAT,
        allowNull: true
      },
      start_time: {
        type: Sequelize.DATE,
        allowNull: true
      },
      end_time: {
        type: Sequelize.DATE,
        allowNull: true
      },
      sort_order: {
        type: Sequelize.INTEGER,
        defaultValue: 0
      },
      visited: {
        type: Sequelize.BOOLEAN,
        defaultValue: false
      },
      visited_at: {
        type: Sequelize.DATE,
        allowNull: true
      },
      category: {
        type: Sequelize.ENUM('attraction', 'food', 'accommodation', 'transport', 'shopping', 'entertainment', 'other'),
        defaultValue: 'other'
      },
      cost: {
        type: Sequelize.DECIMAL(10, 2),
        defaultValue: 0
      },
      notes: {
        type: Sequelize.TEXT,
        allowNull: true
      },
      trip_id: {
        type: Sequelize.INTEGER,
        allowNull: false,
        references: {
          model: 'trips',
          key: 'id'
        },
        onUpdate: 'CASCADE',
        onDelete: 'CASCADE'
      },
      created_at: {
        type: Sequelize.DATE,
        allowNull: false,
        defaultValue: Sequelize.literal('CURRENT_TIMESTAMP')
      },
      updated_at: {
        type: Sequelize.DATE,
        allowNull: false,
        defaultValue: Sequelize.literal('CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP')
      },
      deleted_at: {
        type: Sequelize.DATE,
        allowNull: true
      }
    });

    // 创建索引
    await queryInterface.addIndex('trips', ['user_id']);
    await queryInterface.addIndex('trips', ['share_token']);
    await queryInterface.addIndex('trips', ['start_date']);
    
    await queryInterface.addIndex('trip_items', ['trip_id', 'sort_order']);
    await queryInterface.addIndex('trip_items', ['trip_id', 'visited']);
    await queryInterface.addIndex('trip_items', ['category']);
  },

  down: async (queryInterface, Sequelize) => {
    // 删除表（反向迁移）
    await queryInterface.dropTable('trip_items');
    await queryInterface.dropTable('trips');
  }
};