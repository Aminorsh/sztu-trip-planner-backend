 const { DataTypes } = require('sequelize');
 const sequelize = require('../config/database');
 const User = require('./User');

 const Trip = sequelize.define('Trip', {
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
   startDate: {
     type: DataTypes.DATE,
     allowNull: false,
   },
   endDate: {
     type: DataTypes.DATE,
     allowNull: false,
   },
   userId: {
     type: DataTypes.INTEGER,
     allowNull: false,
     references: {
       model: User,
       key: 'id',
     },
   },
   isPublic: {
     type: DataTypes.BOOLEAN,
     defaultValue: false,
   },
   shareToken: {
     type: DataTypes.STRING(255),
     unique: true,
     allowNull: true,
   },
   coverImage: {
     type: DataTypes.STRING(255),
     allowNull: true,
   },
   status: {
     type: DataTypes.ENUM('planning', 'ongoing', 'completed', 'cancelled'),
     defaultValue: 'planning',
   },
 }, {
   tableName: 'trips',
   timestamps: true,
 });

 Trip.belongsTo(User, { foreignKey: 'userId' });
 User.hasMany(Trip, { foreignKey: 'userId' });

 module.exports = Trip;