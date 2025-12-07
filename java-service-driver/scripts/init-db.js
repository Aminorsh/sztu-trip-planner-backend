// scripts/init-db.js
const fs = require('fs').promises;
const path = require('path');
const pool = require('../config/database');

async function initDatabase() {
  let connection;
  try {
    connection = await pool.getConnection();
    
    // 读取SQL文件
    const sqlPath = path.join(__dirname, '../sql/init.sql');
    const sql = await fs.readFile(sqlPath, 'utf8');
    
    // 分割SQL语句并执行
    const statements = sql.split(';').filter(stmt => stmt.trim());
    
    for (const statement of statements) {
      if (statement.trim()) {
        await connection.query(statement);
        console.log('执行SQL:', statement.substring(0, 100) + '...');
      }
    }
    
    console.log('✅ 数据库初始化成功！');
  } catch (error) {
    console.error('❌ 数据库初始化失败:', error);
  } finally {
    if (connection) connection.release();
    process.exit();
  }
}

initDatabase();