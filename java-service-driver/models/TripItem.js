 const { DataTypes } = require('sequelize');
 const sequelize = require('../config/database');
 const Trip = require('./Trip');

 const TripItem = sequelize.define('TripItem', {
   id: {
     type: DataTypes.INTEGER,
     primaryKey: true,
     autoIncrement: true,
   },
   title: {
     type: DataTypes.STRING(100),
     allowNull: false,
   },
   description: {
     type: DataTypes.TEXT,
     allowNull: true,
   },
   location: {
     type: DataTypes.STRING(255),
     allowNull: false,
   },
   address: {
     type: DataTypes.STRING(255),
     allowNull: true,
   },
   latitude: {
     type: DataTypes.FLOAT,
     allowNull: true,
   },
   longitude: {
     type: DataTypes.FLOAT,
     allowNull: true,
   },
   startTime: {
     type: DataTypes.DATE,
     allowNull: true,
   },
   endTime: {
     type: DataTypes.DATE,
     allowNull: true,
   },
   sortOrder: {
     type: DataTypes.INTEGER,
     defaultValue: 0,
   },
   visited: {
     type: DataTypes.BOOLEAN,
     defaultValue: false,
   },
   visitedAt: {
     type: DataTypes.DATE,
     allowNull: true,
   },
   tripId: {
     type: DataTypes.INTEGER,
     allowNull: false,
     references: {
       model: Trip,
       key: 'id',
     },
   },
   category: {
     type: DataTypes.ENUM('attraction', 'food', 'accommodation', 'transport', 'shopping', 'entertainment', 'other'),
     defaultValue: 'other',
   },
   cost: {
     type: DataTypes.DECIMAL(10, 2),
     defaultValue: 0,
   },
   notes: {
     type: DataTypes.TEXT,
     allowNull: true,
   },
 }, {
   tableName: 'trip_items',
   timestamps: true,
 });

 TripItem.belongsTo(Trip, { foreignKey: 'tripId' });
 Trip.hasMany(TripItem, { foreignKey: 'tripId' });

 module.exports = TripItem;