/**
 * 时间格式化
 * @param {String|Date} time - 时间字符串或日期对象
 * @param {String} format - 格式化模板，默认为'YYYY-MM-DD HH:mm:ss'
 * @returns {String} 格式化后的时间字符串
 */
export function formatTime(time, format = 'YYYY-MM-DD HH:mm:ss') {
  if (!time) return ''
  
  let date
  
  if (typeof time === 'string' || typeof time === 'number') {
    date = new Date(time)
  } else if (time instanceof Date) {
    date = time
  } else {
    return ''
  }
  
  if (isNaN(date.getTime())) return ''
  
  const formatMap = {
    'YYYY': date.getFullYear(),
    'MM': String(date.getMonth() + 1).padStart(2, '0'),
    'DD': String(date.getDate()).padStart(2, '0'),
    'HH': String(date.getHours()).padStart(2, '0'),
    'mm': String(date.getMinutes()).padStart(2, '0'),
    'ss': String(date.getSeconds()).padStart(2, '0')
  }
  
  return format.replace(/YYYY|MM|DD|HH|mm|ss/g, (matched) => formatMap[matched])
}

/**
 * 格式化日期（仅年月日）
 * @param {String|Date} date - 日期字符串或日期对象
 * @returns {String} 格式化后的日期字符串
 */
export function formatDate(date) {
  return formatTime(date, 'YYYY-MM-DD')
}
