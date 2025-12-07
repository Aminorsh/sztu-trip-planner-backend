const multer = require('multer');
const path = require('path');

// 配置存储位置和文件名
const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, 'uploads/'); // 文件存储在 uploads 文件夹
  },
  filename: (req, file, cb) => {
    // 文件名格式：用户ID-时间戳.扩展名
    const uniqueSuffix = Date.now() + '-' + Math.round(Math.random() * 1e9);
    cb(null, `avatar-${req.userId}-${uniqueSuffix}${path.extname(file.originalname)}`);
  }
});

// 文件过滤器：只允许图片
const fileFilter = (req, file, cb) => {
  if (file.mimetype.startsWith('image/')) {
    cb(null, true);
  } else {
    cb(new Error('只允许上传图片文件！'), false);
  }
};

const upload = multer({
  storage: storage,
  limits: { fileSize: 5 * 1024 * 1024 }, // 限制5MB
  fileFilter: fileFilter
});

module.exports = upload;