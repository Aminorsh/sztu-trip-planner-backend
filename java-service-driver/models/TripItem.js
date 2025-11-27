// models/TripItem.js
const mongoose = require('mongoose');

const tripItemSchema = new mongoose.Schema({
  title: {
    type: String,
    required: [true, '行程项标题是必需的'],
    trim: true,
    maxlength: [100, '标题不能超过100个字符']
  },
  description: {
    type: String,
    trim: true,
    maxlength: [300, '描述不能超过300个字符']
  },
  location: {
    type: String,
    required: [true, '地点是必需的'],
    trim: true
  },
  address: {
    type: String,
    trim: true
  },
  latitude: {
    type: Number
  },
  longitude: {
    type: Number
  },
  startTime: {
    type: Date
  },
  endTime: {
    type: Date
  },
  sortOrder: {
    type: Number,
    default: 0
  },
  visited: {
    type: Boolean,
    default: false
  },
  visitedAt: {
    type: Date
  },
  tripId: {
    type: mongoose.Schema.Types.ObjectId,
    ref: 'Trip',
    required: true
  },
  category: {
    type: String,
    enum: ['attraction', 'food', 'accommodation', 'transport', 'shopping', 'entertainment', 'other'],
    default: 'other'
  },
  cost: {
    type: Number,
    min: 0,
    default: 0
  },
  notes: {
    type: String,
    maxlength: [200, '备注不能超过200个字符']
  }
}, {
  timestamps: true
});

// 创建索引
tripItemSchema.index({ tripId: 1, sortOrder: 1 });
tripItemSchema.index({ tripId: 1, visited: 1 });
tripItemSchema.index({ tripId: 1, category: 1 });

module.exports = mongoose.model('TripItem', tripItemSchema);