/**
 * React 应用入口文件
 */

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import 'antd/dist/reset.css' // Ant Design 5.x 样式重置

const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error('Root element #root not found in HTML document. Please check public/index.html.')
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
