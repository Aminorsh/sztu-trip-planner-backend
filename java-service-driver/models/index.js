// models/index.js
const User = require('./User');
const Trip = require('./Trip');
const TripItem = require('./TripItem');

// 用户与行程的关系：一对多
User.hasMany(Trip, {
  foreignKey: 'user_id',
  as: 'trips'
});

Trip.belongsTo(User, {
  foreignKey: 'user_id',
  as: 'user'
});

// 行程与行程项的关系：一对多
Trip.hasMany(TripItem, {
  foreignKey: 'trip_id',
  as: 'items'
});

TripItem.belongsTo(Trip, {
  foreignKey: 'trip_id',
  as: 'trip'
});

module.exports = {
  User,
  Trip,
  TripItem
};