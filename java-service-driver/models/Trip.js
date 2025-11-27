// models/Trip.js
const mongoose = require('mongoose');

const tripSchema = new mongoose.Schema({
  title: {
    type: String,
    required: [true, '行程标题是必需的'],
    trim: true,
    maxlength: [100, '标题不能超过100个字符']
  },
  description: {
    type: String,
    trim: true,
    maxlength: [500, '描述不能超过500个字符']
  },
  startDate: {
    type: Date,
    required: [true, '开始日期是必需的']
  },
  endDate: {
    type: Date,
    required: [true, '结束日期是必需的']
  },
  userId: {
    type: mongoose.Schema.Types.ObjectId,
    ref: 'User',
    required: true
  },
  isPublic: {
    type: Boolean,
    default: false
  },
  shareToken: {
    type: String,
    unique: true,
    sparse: true
  },
  coverImage: {
    type: String,
    default: ''
  },
  status: {
    type: String,
    enum: ['planning', 'ongoing', 'completed', 'cancelled'],
    default: 'planning'
  }
}, {
  timestamps: true
});

// 创建索引
tripSchema.index({ userId: 1, createdAt: -1 });
tripSchema.index({ shareToken: 1 });

module.exports = mongoose.model('Trip', tripSchema);