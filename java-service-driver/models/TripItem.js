// models/TripItem.js
const { DataTypes } = require('sequelize');
const sequelize = require('../config/database');

const TripItem = sequelize.define('TripItem', {
  id: {
    type: DataTypes.INTEGER,
    primaryKey: true,
    autoIncrement: true
  },
  title: {
    type: DataTypes.STRING(100),
    allowNull: false,
    validate: {
      notEmpty: true,
      len: [1, 100]
    }
  },
  description: {
    type: DataTypes.TEXT,
    allowNull: true
  },
  location: {
    type: DataTypes.STRING(100),
    allowNull: false
  },
  address: {
    type: DataTypes.STRING(255),
    allowNull: true
  },
  latitude: {
    type: DataTypes.FLOAT,
    allowNull: true
  },
  longitude: {
    type: DataTypes.FLOAT,
    allowNull: true
  },
  start_time: {
    type: DataTypes.DATE,
    allowNull: true
  },
  end_time: {
    type: DataTypes.DATE,
    allowNull: true
  },
  sort_order: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  visited: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  },
  visited_at: {
    type: DataTypes.DATE,
    allowNull: true
  },
  category: {
    type: DataTypes.ENUM('attraction', 'food', 'accommodation', 'transport', 'shopping', 'entertainment', 'other'),
    defaultValue: 'other'
  },
  cost: {
    type: DataTypes.DECIMAL(10, 2),
    defaultValue: 0
  },
  notes: {
    type: DataTypes.TEXT,
    allowNull: true
  },
  trip_id: {
    type: DataTypes.INTEGER,
    allowNull: false,
    references: {
      model: 'trips',
      key: 'id'
    }
  }
}, {
  tableName: 'trip_items',
  timestamps: true,
  paranoid: true,
  indexes: [
    {
      fields: ['trip_id', 'sort_order']
    },
    {
      fields: ['trip_id', 'visited']
    },
    {
      fields: ['category']
    }
  ]
});

module.exports = TripItem;