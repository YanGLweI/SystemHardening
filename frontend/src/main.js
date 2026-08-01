import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import ElementUI from 'element-ui'
import Axios from 'axios'
import 'element-ui/lib/theme-chalk/index.css'
import './assets/styles/main.css'

Vue.use(ElementUI, { size: 'medium', zIndex: 3000 })

// 全局挂载 axios
Vue.prototype.$axios = Axios

Vue.config.productionTip = false

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
