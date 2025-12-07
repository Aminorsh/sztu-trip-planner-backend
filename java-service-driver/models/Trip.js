 // models/Trip.js
const { DataTypes } = require('sequelize');
const sequelize = require('../config/database');

const Trip = sequelize.define('Trip', {
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
  start_date: {
    type: DataTypes.DATE,
    allowNull: false
  },
  end_date: {
    type: DataTypes.DATE,
    allowNull: false
  },
  is_public: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  },
  share_token: {
    type: DataTypes.STRING(100),
    unique: true,
    allowNull: true
  },
  cover_image: {
    type: DataTypes.STRING(255),
    defaultValue: ''
  },
  status: {
    type: DataTypes.ENUM('planning', 'ongoing', 'completed', 'cancelled'),
    defaultValue: 'planning'
  },
  user_id: {
    type: DataTypes.INTEGER,
    allowNull: false,
    references: {
      model: 'users',
      key: 'id'
    }
  }
}, {
  tableName: 'trips',
  timestamps: true,
  paranoid: true,
  indexes: [
    {
      fields: ['user_id']
    },
    {
      fields: ['share_token']
    },
    {
      fields: ['start_date']
    }
  ]
});

module.exports = Trip;