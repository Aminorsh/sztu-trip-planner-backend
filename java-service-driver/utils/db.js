// utils/db.js
const pool = require('../config/database');

class DB {
  // 执行查询
  static async query(sql, params = []) {
    const [rows] = await pool.execute(sql, params);
    return rows;
  }
  
  // 执行查询返回第一行
  static async queryOne(sql, params = []) {
    const rows = await this.query(sql, params);
    return rows[0] || null;
  }
  
  // 插入数据
  static async insert(table, data) {
    const keys = Object.keys(data);
    const values = Object.values(data);
    const placeholders = keys.map(() => '?').join(', ');
    
    const sql = `INSERT INTO ${table} (${keys.join(', ')}) VALUES (${placeholders})`;
    const [result] = await pool.execute(sql, values);
    
    return {
      id: result.insertId,
      affectedRows: result.affectedRows
    };
  }
  
  // 更新数据
  static async update(table, id, data) {
    const keys = Object.keys(data);
    const values = Object.values(data);
    const setClause = keys.map(key => `${key} = ?`).join(', ');
    
    const sql = `UPDATE ${table} SET ${setClause}, updated_at = CURRENT_TIMESTAMP WHERE id = ?`;
    const [result] = await pool.execute(sql, [...values, id]);
    
    return {
      affectedRows: result.affectedRows
    };
  }
  
  // 软删除
  static async softDelete(table, id) {
    const sql = `UPDATE ${table} SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`;
    const [result] = await pool.execute(sql, [id]);
    return result.affectedRows;
  }
  
  // 硬删除
  static async hardDelete(table, id) {
    const sql = `DELETE FROM ${table} WHERE id = ?`;
    const [result] = await pool.execute(sql, [id]);
    return result.affectedRows;
  }
  
  // 事务支持
  static async transaction(callback) {
    const connection = await pool.getConnection();
    try {
      await connection.beginTransaction();
      const result = await callback(connection);
      await connection.commit();
      return result;
    } catch (error) {
      await connection.rollback();
      throw error;
    } finally {
      connection.release();
    }
  }
}

module.exports = DB;